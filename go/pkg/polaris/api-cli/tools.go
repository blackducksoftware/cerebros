package api_cli

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
)

func SetupToolsCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "tools",
		Short: "tools functionality",
		Long:  "tools-related functionality",
		Args:  cobra.ExactArgs(0),
	}

	command.AddCommand(SetupToolsDebugCommand())
	command.AddCommand(SetupPostToolsCommand())
	command.AddCommand(SetupPostToolsGetCurlCommandCommand())

	return command
}

type ToolsDebugArgs struct {
	PolarisURL string
	Email      string
	Password   string
}

func SetupToolsDebugCommand() *cobra.Command {
	args := &ToolsDebugArgs{}

	command := &cobra.Command{
		Use:   "debug",
		Short: "check tools state",
		Long:  "check tools state",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunToolsDebug(args)
		},
	}

	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "https://local.dev.polaris.synopsys.com", "URL of polaris instance")

	command.Flags().StringVar(&args.Email, "email", "", "email of Polaris user")
	command.MarkFlagRequired("email")

	command.Flags().StringVar(&args.Password, "password", "", "Polaris password")
	command.MarkFlagRequired("password")

	return command
}

func RunToolsDebug(args *ToolsDebugArgs) {
	polarisClient := api.NewClient(args.PolarisURL, args.Email, args.Password)

	err := polarisClient.Authenticate()
	DoOrDie(err)

	tools, err := polarisClient.GetTools(25)
	DoOrDie(err)
	fmt.Printf("GET to api/common/v0/tools: %+v\n", tools)

	toolIds, err := polarisClient.QueryV0DiscoveryFilterKeysIssuetoolidValues()
	DoOrDie(err)
	fmt.Printf("GET to api/query/v0/discovery/filter-keys/issue.tool.id/values: %s\n", toolIds)
}

type PostToolsGetCurlCommandArgs struct {
	PolarisURL     string
	Email          string
	Password       string
	UseAccessToken bool
	TokenName      string
}

func SetupPostToolsGetCurlCommandCommand() *cobra.Command {
	args := &PostToolsGetCurlCommandArgs{}

	command := &cobra.Command{
		Use:   "curl",
		Short: "get curl command to issue post to tools",
		Long:  "this should fix an issue with local scans not working",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunPostToolsGetCurlCommand(args)
		},
	}

	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "https://local.dev.polaris.synopsys.com", "URL of polaris instance")

	command.Flags().StringVar(&args.Email, "email", "", "email of Polaris user")
	command.MarkFlagRequired("email")

	command.Flags().StringVar(&args.Password, "password", "", "Polaris password")
	command.MarkFlagRequired("password")

	command.Flags().BoolVar(&args.UseAccessToken, "use-access-token", false, "uses access token instead of auth token; you probably don't want to do this")

	command.Flags().StringVar(&args.TokenName, "token-name", "get-curl-command-cerebros", "name to use for access token in polaris")

	return command
}

const CurlTemplate = `curl --fail --verbose --location --request POST '%s/api/common/v0/tools?=' \
  --header 'Content-Type: application/vnd.api+json' \
  --header 'Authorization: Bearer %s' \
  --data-raw '{
    "data":
      {
        "type": "tool",
        "attributes": {
          "version": "2020.06",
          "name": "Coverity"
        }
      }
    }'
`

func RunPostToolsGetCurlCommand(args *PostToolsGetCurlCommandArgs) {
	polarisClient := api.NewClient(args.PolarisURL, args.Email, args.Password)

	err := polarisClient.Authenticate()
	DoOrDie(err)

	if args.UseAccessToken {
		token, err := polarisClient.GetAccessToken(args.TokenName)
		DoOrDie(err)

		tokenCommand := fmt.Sprintf(CurlTemplate, args.PolarisURL, token.Data.Attributes.AccessToken)
		fmt.Println(tokenCommand)
	} else {
		curlCommand := fmt.Sprintf(CurlTemplate, args.PolarisURL, polarisClient.AuthToken)
		fmt.Println(curlCommand)
	}
}

type PostToolsArgs struct {
	PolarisURL string
	Email      string
	Password   string
	Certfile   string
}

func SetupPostToolsCommand() *cobra.Command {
	args := &PostToolsArgs{}

	command := &cobra.Command{
		Use:   "post",
		Short: "issue post to tools",
		Long:  "this should fix an issue with local scans not working",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunPostTools(args)
		},
	}

	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "https://local.dev.polaris.synopsys.com", "URL of polaris instance")

	command.Flags().StringVar(&args.Email, "email", "", "email of Polaris user")
	command.MarkFlagRequired("email")

	command.Flags().StringVar(&args.Password, "password", "", "Polaris password")
	command.MarkFlagRequired("password")

	command.Flags().StringVar(&args.Certfile, "certfile", "", "path to certfile (may not be necessary)")

	return command
}

func RunPostTools(args *PostToolsArgs) {
	polarisClient := api.NewClient(args.PolarisURL, args.Email, args.Password)

	err := polarisClient.Authenticate()
	DoOrDie(err)

	if args.Certfile != "" {
		// Get the SystemCertPool, continue with an empty pool on error
		rootCAs, err := x509.SystemCertPool()
		DoOrDie(err)
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Read in the cert file
		certs, err := ioutil.ReadFile(args.Certfile)
		DoOrDie(err)

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			panic("No certs appended, using system certs only")
		}

		polarisClient.RestyClient.SetTLSClientConfig(&tls.Config{RootCAs: rootCAs})
		polarisClient.RestyClient.SetRedirectPolicy(resty.FlexibleRedirectPolicy(200000))
	}

	out, err := polarisClient.PostTools()
	log.Infof("post tools response: %s", out)
	DoOrDie(err)
}
