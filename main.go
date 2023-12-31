package main

import (
	"github.com/areed829/mulan/cmd"
	"github.com/areed829/mulan/pkg/config"
	"github.com/spf13/viper"
)

func main() {
	configPath := "config.yaml"
	if !config.ConfigExists(configPath) {
		err := config.CreateDefaultConfig(configPath)
		if err != nil {
			panic(err)
		}
	}
	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	cmd.Execute()
}
