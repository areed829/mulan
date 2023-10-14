package cmd

import (
	"fmt"

	pkg "github.com/areed829/mulan/pkg/bitbucket"
	"github.com/spf13/cobra"
)

var bitbucketCmd = &cobra.Command{
	Use:   "bitbucket",
	Short: "Interface to bitbucket",
}

var bitbucketCloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone a repo from bitbucket",
	Long:  `Clone a repo from bitbucket`,
	Run: func(cmd *cobra.Command, args []string) {
		err := pkg.Clone()
		if err != nil {
			fmt.Println(err)
		}
	},
}

var bitbucketConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure bitbucket",
	Long:  `Configure bitbucket`,
	Run: func(cmd *cobra.Command, args []string) {
		setupSsh, _ := cmd.Flags().GetBool("ssh")
		setupAuth, _ := cmd.Flags().GetBool("auth")
		config := pkg.BitbucketConfigurationSettings{
			SetupSsh:                 setupSsh,
			SetupUsernameAndPassword: setupAuth,
		}
		pkg.ConfigureBitbucket(&config)
	},
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

	bitbucketConfigureCmd.Flags().BoolP("ssh", "s", false, "Configure ssh key for bitbucket")
	bitbucketConfigureCmd.Flags().BoolP("auth", "a", false, "Configure username and app password for bitbucket")
}
