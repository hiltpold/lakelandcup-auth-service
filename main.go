package main

import (
	"github.com/sirupsen/logrus"

	"github.com/hiltpold/lakelandcup-auth-service/commands"
)

func main() {
	if err := commands.RootCommand().Execute(); err != nil {
		logrus.Fatal(err)
	}
}
