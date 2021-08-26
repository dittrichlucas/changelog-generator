<!-- Logo -->
<p align="center">
  <img width="150" src="./doc/assets/icon.png" alt="changelog generator logo" />
</p>

<!-- Attribute the author -->
<div align="center" style="font-size: 8px; margin-top: -15px; margin-bottom: 25px">
    Icons made by
    <a href="https://www.flaticon.com/authors/flat-icons" title="Flat Icons">
        Flat Icons
    </a>
    from
    <a href="https://www.flaticon.com/" title="Flaticon">
        www.flaticon.com
    </a>
</div>


<!-- Name -->
<h1 align="center" style="margin-top:10px">Changelog Generator</h1>

<!-- Badges -->
<div align="center">

A simple Github Action to generate changelog automatically

</div>

## :thinking: Why?

When analyzing some market actions, none of them proposed to generate the incremental changelog and according to the standard established by [keepachangelog](https://keepachangelog.com).

Some, like [github-changelog-generator](https://github.com/github-changelog-generator/github-changelog-generator), become unfeasible in the medium and long term by always regenerating the changelog file with each new release, because, from a performance point of view, the more releases/versions of the product there are, the more time-consuming it will be to generate the changelog automatically.

Thus, this action appears to facilitate changelog generation, since it searches, formats, and adds information only from the new release and not from all.

## :gear: Usage

### :point_right: Basic

Here the release/version changelog will only be generated and the file will be checked with the `cat` command.

```yml
name: Changelog generator
on:
  workflow_dispatch:

jobs:
  generate-changelog:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout
      uses: actions/checkout@v2

    - name: Changelog generator
      uses: dittrichlucas/changelog-generator@main
      with:
        token: ${{ github.token }}
        repo: ${{ github.repository }}

    - name: Check changelog file
      run: cat CHANGELOG.md
```

### :point_right: Advanced

In this other example, the file will be generated and a **PR will be opened** to the repository with the changes made.

```yml
name: Changelog generator
on:
  workflow_dispatch:

jobs:
  generate-changelog:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout
      uses: actions/checkout@v2

    - name: Changelog generator
      uses: dittrichlucas/changelog-generator@main
      with:
        token: ${{ github.token }}
        repo: ${{ github.repository }}

    - name: Create Pull Request
      id: cpr
      uses: peter-evans/create-pull-request@v3
      with:
        token: ${{ github.token}}
        commit-message: "project/ci: update the changelog file with new release deliveries"
        committer: GitHub <noreply@github.com>
        author: ${{ github.actor }} <${{ github.actor }}@users.noreply.github.com>
        signoff: true
        branch: release-branch
        delete-branch: true
        title: "project/ci: generate the changelog file for the new release"
        body: |
          Update the `CHANGELOG.md` file with the deliveries of the new release
        draft: false

    - name: Check outputs
      run: |
        echo "Pull Request Number - ${{ steps.cpr.outputs.pull-request-number }}"
        echo "Pull Request URL - ${{ steps.cpr.outputs.pull-request-url }}"
```

**Note**: Here I am using (and _I suggest_) the [**create-pull-request**](https://github.com/peter-evans/create-pull-request) action because it is very complete and fulfills the purpose I was looking for to open a PR with the updated changelog, but feel free to use the one that is most convenient for your case.

## :memo: Options

All inputs are required.

| Input | Description                                               | Default    |
|-------|-----------------------------------------------------------|------------|
| token | Personal access token (PAT) used to fetch the repository. | Not yet    |
| repo  | Repository name with owner. For example, owner/repo-name. | Not yet    |


## :footprints: Next Steps

Check out the [roadmap](./ROADMAP.md)

## :books: References

I decided to list the articles that served as a basis for me to develop this action in Golang, if it is of interest.

- [Writing Github Actions in Go](https://www.sethvargo.com/writing-github-actions-in-go/) <br>
- [Creating Github Actions in Go](https://jacobtomlinson.dev/posts/2019/creating-github-actions-in-go/) <br>
- [Automate changelog and releases creation in Github](https://renehernandez.io/essays/2020/09/23/automate-changelog-and-releases-creation-in-github/) <br>
- [Changelog pattern](https://keepachangelog.com)

## :scroll: License

The scripts and documentation in this project are released under the [MIT License](LICENSE)
