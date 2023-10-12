package bitbucket

import (
	"fmt"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/spf13/viper"
)

func ConfigureBitbucket() {
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
	viper.Set("bitbucket.username", username)
	viper.Set("bitbucket.password", password)
	viper.WriteConfig()
}
