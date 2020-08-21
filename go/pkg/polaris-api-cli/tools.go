package polaris_api_cli

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
)

func setupToolsCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "tools",
		Short: "tools functionality",
		Long:  "tools-related functionality",
		Args:  cobra.ExactArgs(0),
	}

	command.AddCommand(setupToolsDebugCommand())
	command.AddCommand(setupPostToolsCommand())

	return command
}

type ToolsDebugArgs struct {
	PolarisURL string
	Email      string
	Password   string
}

func setupToolsDebugCommand() *cobra.Command {
	args := &ToolsDebugArgs{}

	command := &cobra.Command{
		Use:   "debug",
		Short: "check tools state",
		Long:  "check tools state",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runToolsDebug(args)
		},
	}

	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "https://local.dev.polaris.synopsys.com/", "URL of polaris instance")

	command.Flags().StringVar(&args.Email, "email", "", "email of Polaris user")
	command.MarkFlagRequired("email")

	command.Flags().StringVar(&args.Password, "password", "", "Polaris password")
	command.MarkFlagRequired("password")

	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "https://local.dev.polaris.synopsys.com/", "URL of polaris instance")

	return command
}

func runToolsDebug(args *ToolsDebugArgs) {
	polarisClient := api.NewClient(args.PolarisURL, args.Email, args.Password)

	err := polarisClient.Authenticate()
	DoOrDie(err)

	log.Warningf("what do we have?")
	tools, err := polarisClient.GetTools(25)
	DoOrDie(err)
	log.Warningf("my tools: %+v", tools)
}

type PostToolsArgs struct {
	PolarisURL string
	Email      string
	Password   string
	Certfile   string
}

func setupPostToolsCommand() *cobra.Command {
	args := &PostToolsArgs{}

	command := &cobra.Command{
		Use:   "post",
		Short: "issue post to tools",
		Long:  "this should fix an issue with local scans not working",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runPostTools(args)
		},
	}

	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "https://local.dev.polaris.synopsys.com/", "URL of polaris instance")

	command.Flags().StringVar(&args.Email, "email", "", "email of Polaris user")
	command.MarkFlagRequired("email")

	command.Flags().StringVar(&args.Password, "password", "", "Polaris password")
	command.MarkFlagRequired("password")

	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "https://local.dev.polaris.synopsys.com/", "URL of polaris instance")
	command.Flags().StringVar(&args.Certfile, "certfile", "", "path to certfile (may not be necessary)")

	return command
}

func runPostTools(args *PostToolsArgs) {
	polarisClient := api.NewClient(args.PolarisURL, args.Email, args.Password)

	err := polarisClient.Authenticate()
	DoOrDie(err)

	if args.Certfile != "" {
		toolIds, err := polarisClient.QueryV0DiscoveryFilterKeysIssuetoolidValues()
		DoOrDie(err)
		log.Warningf("tool ids: %s", toolIds)

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
