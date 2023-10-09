/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/cobra"
)

var memberOfRepos bitbucket.RepositoriesRes
var repoNames []string
var selectedRepo bitbucket.Repository

// bitbucketCmd represents the bitbucket command
var bitbucketCmd = &cobra.Command{
	Use:   "bitbucket",
	Short: "Interface to bitbucket",
	Long:  `I want to be able to clone repos from bitbucket,`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bitbucket called")
		username := ""
		password := ""

		c := bitbucket.NewBasicAuth(username, password)

		opts := &bitbucket.RepositoriesOptions{
			Role: "member",
		}

		repos, err := c.Repositories.ListForAccount(opts)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		memberOfRepos = *repos
		for _, repo := range memberOfRepos.Items {
			repoNames = append(repoNames, repo.Name)
		}

		promptForRepo()
	},
}

func init() {
	rootCmd.AddCommand(bitbucketCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bitbucketCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bitbucketCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
		if strings.ToLower(repo.Name) == strings.ToLower(in) {
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
