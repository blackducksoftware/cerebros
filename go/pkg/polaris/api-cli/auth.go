package api_cli

import (
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type AuthArgs struct {
	PolarisURL string
	Email      string
	Password   string
	Type       string
}

func SetupAuthCommand() *cobra.Command {
	args := &AuthArgs{}
	command := &cobra.Command{
		Use:   "auth",
		Short: "Auth commands",
		Long:  "use commands for auth server",
		Run: func(cmd *cobra.Command, as []string) {
			runAuth(args)
		},
	}

	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "", "polaris URL")
	command.MarkFlagRequired("polaris-url")

	command.Flags().StringVar(&args.Email, "email", "", "email")
	command.MarkFlagRequired("email")

	command.Flags().StringVar(&args.Password, "password", "", "password")
	command.MarkFlagRequired("password")

	command.Flags().StringVar(&args.Type, "type", "users", "type of data to fetch: orgs, roles, users, or groups")

	return command
}

func runAuth(args *AuthArgs) {
	log.Infof("cos args: %+v", args)

	client := api.NewClient(args.PolarisURL, args.Email, args.Password)

	DoOrDie(client.Authenticate())

	switch args.Type {
	case "orgs", "organizations":
		printJson(client.GetOrganizations())
	case "roles":
		printJson(client.GetRoles())
	case "users":
		printJson(client.GetUsers(0, 10))
	case "groups":
		printJson(client.GetGroups())
	default:
		panic(errors.Errorf("invalid type: %s", args.Type))
	}
}
