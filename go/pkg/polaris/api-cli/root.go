package api_cli

import (
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Run() {
	rootCmd := SetupRootCommand()
	if err := errors.Wrapf(rootCmd.Execute(), "run root command"); err != nil {
		log.Fatalf("unable to run root command: %+v", err)
		os.Exit(1)
	}
}

type RootFlags struct {
	LogLevel string
}

func SetupRootCommand() *cobra.Command {
	args := &RootFlags{}
	rootCmd := &cobra.Command{
		Use:   "polaris-api-cli",
		Short: "polaris API client",
		Long:  "polaris API client",
		PersistentPreRunE: func(cmd *cobra.Command, as []string) error {
			return SetUpLogger(args.LogLevel)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&args.LogLevel, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	rootCmd.AddCommand(SetupScanCommand())
	rootCmd.AddCommand(SetupToolsCommand())
	rootCmd.AddCommand(SetupCosCommand())
	rootCmd.AddCommand(SetupAuthCommand())
	//rootCmd.AddCommand(setupExampleCommand())

	return rootCmd
}

func DoOrDie(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %+v\n", err)
	}
}

func SetUpLogger(logLevelStr string) error {
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		return errors.Wrapf(err, "unable to parse the specified log level: '%s'", logLevel)
	}
	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.Infof("log level set to '%s'", log.GetLevel())
	return nil
}
