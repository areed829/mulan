package bitbucket

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	git "github.com/go-git/go-git/v5"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/viper"
)

var (
	memberOfRepos bitbucket.RepositoriesRes
	repoNames     []string
	selectedRepo  bitbucket.Repository
)

func Clone() error {
	reposErr := setRepos()
	if reposErr != nil {
		return reposErr
	}

	promptForRepo()

	sshLink := getSshLink()

	homeDir, directoryErr := os.UserHomeDir()
	if directoryErr != nil {
		return directoryErr
	}
	clonePath := filepath.Join(homeDir, "projects", selectedRepo.Name)
	_, cloneErr := git.PlainClone(clonePath, false, &git.CloneOptions{
		URL:      sshLink,
		Progress: os.Stdout,
	})
	if cloneErr != nil {
		return cloneErr
	}

	return nil
}

func getSshLink() string {
	links := selectedRepo.Links

	cloneLinks := links["clone"].([]interface{})

	var sshLink string
	for _, cloneLink := range cloneLinks {
		if (cloneLink.(map[string]interface{})["name"]) == "ssh" {
			sshLink = cloneLink.(map[string]interface{})["href"].(string)
		}
	}
	return sshLink
}

func setRepos() error {
	username := viper.GetString("bitbucket.username")
	password := viper.GetString("bitbucket.password")

	c := bitbucket.NewBasicAuth(username, password)

	opts := &bitbucket.RepositoriesOptions{
		Role: "member",
	}

	repos, err := c.Repositories.ListForAccount(opts)
	if err != nil {
		return err
	}

	memberOfRepos = *repos
	for _, repo := range memberOfRepos.Items {
		repoNames = append(repoNames, repo.Name)
	}

	return nil
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
