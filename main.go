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
	"github.com/spf13/viper"
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

const (
	closedIssues = "\n\n**Closed issues:**\n"
	mergedPR     = "\n\n**Merged pull requests:**\n"
)

var (
	title         = "\n## [%s](%s) (%s)"
	fullChangelog = "\n\n[Full Changelog](https://github.com/%v/%v/compare/%v...%v)"
	issueTemplate = "\n- %s [#%v](%s)"
	prTemplate    = "\n- %s [#%v](%s) ([%s](%s))"
	token         = ga.GetInput("token")
	owner         = ga.GetInput("owner")
	repo          = ga.GetInput("repo")
	ctx           = context.Background()
)

func main() {
	client := setupClient()

	if token == "" || repo == "" {
		ga.Fatalf("missing required inputs")
	}

	ga.AddMask(token)

	previousRelease := getPreviousRelease(client)
	nextRelease := getNextRelease(client)

	closedIssues := filterIssues(client, previousRelease, nextRelease)
	mergedPulls := filterPulls(client, previousRelease, nextRelease)

	generateChangelog(previousRelease, nextRelease, closedIssues, mergedPulls)

	fmt.Println("Finish!")
}

func generateChangelog(previousRelease, nextRelease tagData, issues []issuesData, prs []pullsData) {
	// Leitura do arquivo
	file := filepath.Join("CHANGELOG.md")
	fileRead, _ := ioutil.ReadFile(file)
	lines := strings.Split(string(fileRead), "\n")

	// Lógica: https: //stackoverflow.com/questions/46128016/insert-a-value-in-a-slice-at-a-given-index
	lines = append(lines[:1+1], lines[1:]...)
	formatTitle := fmt.Sprintf(title, nextRelease.tagName, nextRelease.link, nextRelease.publishedAt.Format("2006-01-04"))
	formatFullChangelog := fmt.Sprintf(fullChangelog, owner, repo, previousRelease.tagName, nextRelease.tagName) // TODO: Ajustar para o caso da zup (ou quando o owner ou organização for diferente...)
	lines[1] = formatTitle + formatFullChangelog

	// Valida e formata a parte das issues
	if len(issues) > 0 {
		lines[1] = lines[1] + closedIssues

		for _, issue := range issues {
			lines[1] = lines[1] + fmt.Sprintf(issueTemplate, issue.title, issue.issueNumber, issue.link)
		}
	}

	// Valida e formata a parte das prs
	if len(prs) > 0 {
		lines[1] = lines[1] + mergedPR

		for _, pr := range prs {
			lines[1] = lines[1] + fmt.Sprintf(prTemplate, pr.title, pr.prNumber, pr.link, pr.assigneeUser, pr.assigneeLink)
		}
	}

	// Escreve no arquivo o changelog gerado
	newFile := strings.Join(lines, "\n")
	ioutil.WriteFile(file, []byte(newFile), os.ModePerm)
}

func filterIssues(client *github.Client, pr, nr tagData) []issuesData {
	// Seleciona todas as issues fechadas depois da data de criação da tag
	issues, _, err := client.Issues.ListByRepo(
		context.Background(),
		owner,
		repo,
		&github.IssueListByRepoOptions{State: "closed", Since: pr.publishedAt.Time},
	)
	if err != nil {
		log.Fatalf("error listing issues: %v", err)
	}

	// Coloca todos os títulos das issues elegíveis dentro do slice para uso posterior
	var filteredIssues []issuesData
	for _, issue := range issues {
		if issue.ClosedAt.After(pr.publishedAt.Time) && issue.PullRequestLinks == nil && issue.ClosedAt.Before(nr.publishedAt.Time) {
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

func filterPulls(client *github.Client, psr, nr tagData) []pullsData {
	// Seleciona todas as pr fechadas
	prs, _, err := client.PullRequests.List(
		ctx,
		owner,
		repo,
		&github.PullRequestListOptions{State: "closed"},
	)
	if err != nil {
		log.Fatalf("error listing prs: %v", err)
	}

	// Filtra as prs mergeadas após a data de criação da tag
	// TODO: abrir issue no repo go-github, pois retorna erro ao usar o campo name da struct de user
	var mergedPulls []pullsData
	for _, pr := range prs {
		if pr.MergedAt.After(psr.publishedAt.Time) && pr.MergedAt.Before(nr.publishedAt.Time) {
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

	return mergedPulls
}

func getNextRelease(client *github.Client) tagData {
	// Pega a última tag do repositório
	lastTag, _, err := client.Repositories.GetLatestRelease(
		context.Background(),
		owner,
		repo,
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

func getPreviousRelease(client *github.Client) tagData {
	// Pega a última tag do repositório
	tags, _, err := client.Repositories.ListReleases(
		context.Background(),
		owner,
		repo,
		&github.ListOptions{},
	)
	if err != nil {
		log.Fatalf("error getting the previous tag: %v", err)
	}

	previousRelease := tags[1]

	prData := tagData{
		tagName:     *previousRelease.TagName,
		link:        *previousRelease.HTMLURL,
		publishedAt: previousRelease.PublishedAt,
	}

	return prData
}

func getConfig(key string) string {
	viper.SetConfigFile(".env")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("error while reading config file: %v", err)
	}

	value, ok := viper.Get(key).(string)
	if !ok {
		log.Fatalf("invalid type assertion")
	}

	return value
}

func setupClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client
}
