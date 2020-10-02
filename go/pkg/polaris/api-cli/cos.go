package api_cli

import (
	"encoding/json"
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type CosArgs struct {
	PolarisURL string
	Email      string
	Password   string
	Type       string
}

func SetupCosCommand() *cobra.Command {
	args := &CosArgs{}
	command := &cobra.Command{
		Use:   "cos",
		Short: "cos commands",
		Long:  "use commands for common object server",
		Run: func(cmd *cobra.Command, as []string) {
			runCos(args)
		},
	}

	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "", "polaris URL")
	command.MarkFlagRequired("polaris-url")

	command.Flags().StringVar(&args.Email, "email", "", "email")
	command.MarkFlagRequired("email")

	command.Flags().StringVar(&args.Password, "password", "", "password")
	command.MarkFlagRequired("password")

	command.Flags().StringVar(&args.Type, "type", "projects", "type of data to fetch: projects or tools")

	return command
}

func runCos(args *CosArgs) {
	log.Infof("cos args: %+v", args)

	client := api.NewClient(args.PolarisURL, args.Email, args.Password)

	DoOrDie(client.Authenticate())

	switch args.Type {
	case "projects":
		printJson(client.GetProjects(10))
	case "tools":
		printJson(client.GetTools(10))
	default:
		panic(errors.Errorf("invalid type: %s", args.Type))
	}
}

func printJson(obj interface{}, err error) {
	DoOrDie(err)
	bytes, err := json.MarshalIndent(obj, "", "  ")
	DoOrDie(err)
	fmt.Printf("%s\n", bytes)
}
