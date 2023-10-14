package bitbucket

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/viper"
)

var (
	memberOfRepos bitbucket.RepositoriesRes
	repoNames     []string
	selectedRepo  bitbucket.Repository
)

func Clone() {
	color.White("Grabbing available repos...")
	setRepos()

	color.White("Select a repo to clone...")
	promptForRepo()

	cloneSelectedRepo()
}

func logError(err error) {
	if err != nil {
		color.Red("Error: %s", err)
		panic(err)
	}
}

func cloneSelectedRepo() {
	privateKeyFile := filepath.Join(os.Getenv("HOME"), ".ssh", "bitbucket")
	publicKeys, publicKeyErr := gitssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	logError(publicKeyErr)

	sshUrl := getSshUrl()

	color.White("Cloning %s\n", sshUrl)

	homeDir, err := os.UserHomeDir()
	logError(err)

	clonePath := filepath.Join(homeDir, "projects", selectedRepo.Name)
	_, err = git.PlainClone(clonePath, false, &git.CloneOptions{
		Auth:              publicKeys,
		URL:               sshUrl,
		Progress:          os.Stdout,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			color.Red("Repository already exists")
		} else {
			logError(err)
		}
		return
	}
	color.Green("Cloned %s to %s", sshUrl, clonePath)
}

func getSshUrl() string {
	links := selectedRepo.Links

	cloneLinks := links["clone"].([]interface{})

	var sshUrl string
	for _, cloneLink := range cloneLinks {
		if (cloneLink.(map[string]interface{})["name"]) == "ssh" {
			sshUrl = cloneLink.(map[string]interface{})["href"].(string)
		}
	}
	return sshUrl
}

func setRepos() {
	username := viper.GetString("bitbucket.username")
	password := viper.GetString("bitbucket.password")

	c := bitbucket.NewBasicAuth(username, password)

	opts := &bitbucket.RepositoriesOptions{
		Role: "member",
	}

	repos, err := c.Repositories.ListForAccount(opts)
	logError(err)

	memberOfRepos = *repos
	for _, repo := range memberOfRepos.Items {
		repoNames = append(repoNames, repo.Name)
	}
}

func executor(in string) {
	for _, repo := range memberOfRepos.Items {
		if repo.Name == in {
			selectedRepo = repo
			return
		}
	}
}

func exitChecker(in string, _ bool) bool {
	for _, repo := range memberOfRepos.Items {
		if strings.EqualFold(repo.Name, in) {
			return true
		}
	}
	return false
}

func promptForRepo() {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("Enter repo: "),
		prompt.OptionTitle("Repo Selection"),
		prompt.OptionSetExitCheckerOnInput(exitChecker),
	)
	p.Run()
}

func completer(d prompt.Document) []prompt.Suggest {
	var s []prompt.Suggest
	for _, repo := range memberOfRepos.Items {
		if d.GetWordBeforeCursor() != "" {
			if matched := d.TextBeforeCursor(); matched != "" && len(matched) > 0 {
				if fuzzy.MatchFold(matched, repo.Name) {
					s = append(s, prompt.Suggest{Text: repo.Name})
				}
			}
		}
	}
	return s
}
