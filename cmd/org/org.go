package org

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"shipyard/display"
	"shipyard/requests"
	"shipyard/requests/uri"
)

func NewGetAllOrgsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "orgs",
		Aliases: []string{"organizations"},
		Short:   "Get all orgs",
		Long: `Lists all orgs, to which the user belongs.
Note that this command requires a user-level access token.`,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlag("json", cmd.Flags().Lookup("json"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return getAllOrgs()
		},
	}

	cmd.Flags().Bool("json", false, "JSON output")

	return cmd
}

func NewGetCurrentOrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "org",
		Aliases:      []string{"organization"},
		Short:        "Get the currently configured org",
		Long:         "Gets the org that is currently set in the default or custom config",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			getCurrentOrg()
		},
	}

	return cmd
}

func getCurrentOrg() {
	writer := display.NewSimpleDisplay()
	org := viper.GetString("org")
	msg := org
	if msg == "" {
		msg = "no org is found in the config"
	}
	writer.Println(msg)
}

func getAllOrgs() error {
	client, err := requests.NewHTTPClient(os.Stdout)
	if err != nil {
		return err
	}

	body, err := client.Do(http.MethodGet, uri.CreateResourceURI("", "org", "", "", nil), nil)
	if err != nil {
		return err
	}

	if viper.GetBool("json") {
		return client.Write(body)
	}

	var resp orgsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal orgs response: %w", err)
	}

	var orgs []string
	for _, item := range resp.Data {
		orgs = append(orgs, item.Attributes.Name)
	}

	return client.Write(strings.Join(orgs, "\n") + "\n")
}

type orgsResponse struct {
	Data []struct {
		Attributes struct {
			Name string `json:"name"`
		} `json:"attributes"`
	} `json:"data"`
}
