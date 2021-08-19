package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	ga "github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

type tagData struct {
	tagName     string
	link        string
	publishedAt *github.Timestamp
}

type issuesData struct {
	title       string
	issueNumber int
	link        string
}

type pullsData struct {
	title        string
	prNumber     int
	link         string
	assigneeUser string
	assigneeLink string
}

type changelog struct {
	owner           string
	repoName        string
	previousRelease tagData
	nextRelease     tagData
	closedIssues    []issuesData
	mergedPulls     []pullsData
}

type filterData struct {
	client          *github.Client
	previousRelease tagData
	nextRelease     tagData
	owner           string
	repoName        string
}

const (
	fileName     = "CHANGELOG.md"
	closedIssues = "\n\n**Closed issues:**\n"
	mergedPR     = "\n\n**Merged pull requests:**\n"
)

var (
	changelogTitle = "# Changelog\n"
	title          = "\n## [%s](%s) (%s)"
	fullChangelog  = "\n\n[Full Changelog](https://github.com/%v/%v/compare/%v...%v)"
	issueTemplate  = "\n- %s [#%v](%s)"
	prTemplate     = "\n- %s [#%v](%s) ([%s](%s))"
	token          = ga.GetInput("token")
	repo           = ga.GetInput("repo")
	ctx            = context.Background()
)

func main() {
	client := setupClient()

	if token == "" || repo == "" {
		ga.Fatalf("missing required inputs")
	}

	ga.AddMask(token)

	splitRepo := strings.Split(repo, "/")
	owner := splitRepo[0]
	repoName := splitRepo[1]

	previousRelease := getPreviousRelease(client, owner, repoName)
	nextRelease := getNextRelease(client, owner, repoName)

	reqData := filterData{
		client,
		previousRelease,
		nextRelease,
		owner,
		repoName,
	}

	closedIssues := filterIssues(reqData)
	mergedPulls := filterPulls(reqData)

	changelogData := changelog{
		owner,
		repoName,
		previousRelease,
		nextRelease,
		closedIssues,
		mergedPulls,
	}

	generateChangelog(changelogData)

	fmt.Println("Process completed successfully!")
}

func generateChangelog(c changelog) {
	file := filepath.Join(fileName)
	// Cria o arquivo apenas se houver issue ou pr na release gerada
	if !fileExists(file) || (len(c.closedIssues) > 0 || len(c.mergedPulls) > 0) {
		err := ioutil.WriteFile(file, []byte(changelogTitle), os.ModePerm)
		if err != nil {
			log.Fatalf("Unable to write file: %v", err)
		}
	}

	fileRead, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	lines := strings.Split(string(fileRead), "\n")

	// Lógica: https: //stackoverflow.com/questions/46128016/insert-a-value-in-a-slice-at-a-given-index
	lines = append(lines[:1+1], lines[1:]...)
	formatTitle := fmt.Sprintf(
		title,
		c.nextRelease.tagName,
		c.nextRelease.link,
		c.nextRelease.publishedAt.Format("2006-01-04"),
	)
	formatFullChangelog := fmt.Sprintf(
		fullChangelog,
		c.owner,
		c.repoName,
		c.previousRelease.tagName,
		c.nextRelease.tagName,
	)
	lines[1] = formatTitle + formatFullChangelog

	// Valida e formata a parte das issues
	if len(c.closedIssues) > 0 {
		lines[1] = lines[1] + closedIssues

		for _, issue := range c.closedIssues {
			lines[1] = lines[1] + fmt.Sprintf(issueTemplate, issue.title, issue.issueNumber, issue.link)
		}
	}

	// Valida e formata a parte das prs
	if len(c.mergedPulls) > 0 {
		lines[1] = lines[1] + mergedPR

		for _, pr := range c.mergedPulls {
			lines[1] = lines[1] + fmt.Sprintf(
				prTemplate,
				pr.title,
				pr.prNumber,
				pr.link,
				pr.assigneeUser,
				pr.assigneeLink,
			)
		}
	}

	if len(c.closedIssues) > 0 || len(c.mergedPulls) > 0 {
		// Escreve no arquivo o changelog gerado
		newFile := strings.Join(lines, "\n")
		ioutil.WriteFile(file, []byte(newFile), os.ModePerm)
	}
}

func filterIssues(d filterData) []issuesData {
	if d.previousRelease.tagName != "" {
		// Seleciona todas as issues fechadas depois da data de criação da tag
		issues, _, err := d.client.Issues.ListByRepo(
			context.Background(),
			d.owner,
			d.repoName,
			&github.IssueListByRepoOptions{State: "closed", Since: d.previousRelease.publishedAt.Time},
		)
		if err != nil {
			log.Fatalf("error listing issues: %v", err)
		}

		// Coloca todos os títulos das issues elegíveis dentro do slice para uso posterior
		var filteredIssues []issuesData
		for _, issue := range issues {
			if issue.ClosedAt.After(d.previousRelease.publishedAt.Time) && issue.PullRequestLinks == nil && issue.ClosedAt.Before(d.nextRelease.publishedAt.Time) {
				filterIssue := issuesData{
					title:       *issue.Title,
					issueNumber: *issue.Number,
					link:        *issue.HTMLURL,
				}
				filteredIssues = append(filteredIssues, filterIssue)
			}
		}

		return filteredIssues
	}

	return []issuesData{}
}

func filterPulls(d filterData) []pullsData {
	// Seleciona todas as pr fechadas
	prs, _, err := d.client.PullRequests.List(
		ctx,
		d.owner,
		d.repoName,
		&github.PullRequestListOptions{State: "closed"},
	)
	if err != nil {
		log.Fatalf("error listing prs: %v", err)
	}

	// Filtra as prs mergeadas após a data de criação da tag
	// TODO: abrir issue no repo go-github, pois retorna erro ao usar o campo name da struct de user
	var mergedPulls []pullsData
	if d.previousRelease.link != "" {
		for _, pr := range prs {
			if pr.MergedAt.After(d.previousRelease.publishedAt.Time) && pr.MergedAt.Before(d.nextRelease.publishedAt.Time) {
				filterPull := pullsData{
					title:        *pr.Title,
					prNumber:     *pr.Number,
					link:         *pr.HTMLURL,
					assigneeUser: *pr.User.Login,
					assigneeLink: *pr.User.HTMLURL,
				}
				mergedPulls = append(mergedPulls, filterPull)
			}
		}
	}

	return mergedPulls
}

func getNextRelease(client *github.Client, owner, repoName string) tagData {
	// Pega a última tag do repositório
	lastTag, _, err := client.Repositories.GetLatestRelease(
		context.Background(),
		owner,
		repoName,
	)
	if err != nil {
		log.Fatalf("error getting the last tag: %v", err)
	}

	nrData := tagData{
		tagName:     *lastTag.TagName,
		link:        *lastTag.HTMLURL,
		publishedAt: lastTag.PublishedAt,
	}

	return nrData
}

func getPreviousRelease(client *github.Client, owner, repoName string) tagData {
	// Pega a última tag do repositório
	tags, _, err := client.Repositories.ListReleases(
		context.Background(),
		owner,
		repoName,
		&github.ListOptions{},
	)
	if err != nil {
		log.Fatalf("error getting the previous tag: %v", err)
	}

	if len(tags) <= 1 {
		return tagData{}
	}

	previousRelease := tags[1]

	prData := tagData{
		tagName:     *previousRelease.TagName,
		link:        *previousRelease.HTMLURL,
		publishedAt: previousRelease.PublishedAt,
	}

	return prData
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func setupClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client
}
