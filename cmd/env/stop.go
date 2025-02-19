package env

import (
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shipyard/shipyard-cli/constants"
	"github.com/shipyard/shipyard-cli/display"
	"github.com/shipyard/shipyard-cli/requests"
	"github.com/shipyard/shipyard-cli/requests/uri"
)

func NewStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stop",
		GroupID: constants.GroupEnvironments,
		Short:   "Stop a running environment",
		Long:    `This command stops a running environment. You can ONLY stop an environment if it is currently running.`,
		Example: `  # Stop environment ID 12345
  shipyard stop environment 12345`,
	}

	cmd.AddCommand(newStopEnvironmentCmd())

	return cmd
}

func newStopEnvironmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Aliases: []string{"env"},
		Use:     "environment [environment ID]",
		Short:   "Stop a running environment",
		Long:    `This command stops a running environment. You can ONLY stop an environment if it is currently running.`,
		Example: `  # Stop environment ID 12345
  shipyard stop environment 12345`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return stopEnvironmentByID(args[0])
			}
			return errNoEnvironment
		},
	}

	return cmd
}

func stopEnvironmentByID(id string) error {
	client, err := requests.NewHTTPClient(os.Stdout)
	if err != nil {
		return err
	}

	params := make(map[string]string)
	org := viper.GetString("org")
	if org != "" {
		params["org"] = org
	}

	_, err = client.Do(http.MethodPost, uri.CreateResourceURI("stop", "environment", id, "", params), nil)
	if err != nil {
		return err
	}

	out := display.NewSimpleDisplay()
	out.Println("Environment stopped.")
	return nil
}
