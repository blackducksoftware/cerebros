package polaris_api_cli

import (
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	synopsys_scancli "github.com/blackducksoftware/cerebros/go/pkg/synopsys-scancli"
	"os/user"
	"path"

	"github.com/spf13/cobra"
)

type ScanArgs struct {
	PolarisURL       string
	GithubRepo       string
	Email            string
	Password         string
	OSType           string
	UseLocalAnalysis bool
	ProjectName      string
	CLIPath          string
}

func getHomeDir() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	return user.HomeDir
}

func setupScanCommand() *cobra.Command {
	args := &ScanArgs{}

	command := &cobra.Command{
		Use:   "scan",
		Short: "run Polaris scan",
		Long:  "run Polaris scan",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runScan(args)
		},
	}

	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "https://local.dev.polaris.synopsys.com/", "URL of polaris instance")

	command.Flags().StringVar(&args.GithubRepo, "github-repo", "", "name of github repo to scan")
	command.MarkFlagRequired("github-repo")

	command.Flags().StringVar(&args.Email, "email", "", "email of Polaris user")
	command.MarkFlagRequired("email")

	command.Flags().StringVar(&args.Password, "password", "", "Polaris password")
	command.MarkFlagRequired("password")

	command.Flags().StringVar(&args.ProjectName, "project-name", "", "project name to use in Polaris")
	command.MarkFlagRequired("project-name")

	command.Flags().StringVar(&args.OSType, "ostype", "mac", "linux, mac or windows")
	command.Flags().StringVar(&args.PolarisURL, "polaris-url", "https://local.dev.polaris.synopsys.com/", "URL of polaris instance")
	command.Flags().BoolVar(&args.UseLocalAnalysis, "local", false, "perform coverity analysis locally")

	command.Flags().StringVar(&args.CLIPath, "clipath", path.Join(getHomeDir(), "Downloads", "polaris_cli-macosx"), "path to look for polaris tools at, or download to if not found")

	return command
}

func runScan(args *ScanArgs) {
	polarisConfig := &synopsys_scancli.PolarisConfig{
		CLIPath:  args.CLIPath,
		URL:      args.PolarisURL,
		Email:    args.Email,
		Password: args.Password,
		OSType:   api.MustParseOSType(args.OSType),
	}
	scanConfig := &synopsys_scancli.ScanConfig{
		Key: args.ProjectName,
		ScanType: &synopsys_scancli.ScanTypeConfig{
			Polaris: true,
		},
		CodeLocation: &synopsys_scancli.CodeLocation{
			GitRepo: &synopsys_scancli.GitRepo{Repo: args.GithubRepo},
		},
	}
	scanner, err := synopsys_scancli.NewScannerFromConfig(nil, polarisConfig, nil)
	DoOrDie(err)
	err = scanner.Scan(scanConfig)
	DoOrDie(err)
}
