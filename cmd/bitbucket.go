package cmd

import (
	"fmt"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
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
}

var bitbucketCloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone a repo from bitbucket",
	Long:  `Clone a repo from bitbucket`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("clone called")
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
		fmt.Println("Selected repo:", selectedRepo.Name)
	},
}

var bitbucketConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure bitbucket",
	Long:  `Configure bitbucket`,
	Run:   configure,
}

func init() {
	rootCmd.AddCommand(bitbucketCmd)
	bitbucketCmd.AddCommand(bitbucketConfigureCmd)
	bitbucketCmd.AddCommand(bitbucketCloneCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bitbucketCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bitbucketCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func configure(cmd *cobra.Command, args []string) {
	var username string
	var password string

	prompt := []*survey.Question{
		{
			Name:     "username",
			Prompt:   &survey.Input{Message: "Enter username:"},
			Validate: survey.Required,
		},
		{
			Name:     "password",
			Prompt:   &survey.Password{Message: "Enter password:"},
			Validate: survey.Required,
		},
	}

	var answers struct {
		Username string `survey:"username"`
		Password string `survey:"password"`
	}

	err := survey.Ask(prompt, &answers)
	if err != nil {
		fmt.Println(err)
		return
	}

	username = answers.Username
	password = answers.Password

	fmt.Printf("Username: %s\nPassword: %s\n", username, password)

	// use viper to store username and password
	// viper.Set("bitbucket.username", username)
	// viper.Set("bitbucket.password", password)
	// viper.WriteConfig()
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
		// if strings.ToLower(repo.Name) == strings.ToLower(in) {
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
