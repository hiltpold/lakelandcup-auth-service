package commands

import (
	"github.com/hiltpold/lakelandcup-auth-service/conf"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var configFile = ""

var rootCmd = cobra.Command{
	Use:   "auth",
	Short: "Lakelandcup Authentication Service",
	Run: func(cmd *cobra.Command, args []string) {
		runWithConfig(cmd, serve)
	},
}

// RootCommand will setup and return the root command
func RootCommand() *cobra.Command {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "the config file to use")
	rootCmd.AddCommand(&serveCmd, &versionCmd)
	return &rootCmd
}

func runWithConfig(cmd *cobra.Command, fn func(conf *conf.Configuration)) {
	conf, err := conf.LoadConfig(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %+v", err)
	}
	fn(conf)
}
