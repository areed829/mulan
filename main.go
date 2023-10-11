package main

import (
	"github.com/areed829/mulan/cmd"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	cmd.Execute()
}
