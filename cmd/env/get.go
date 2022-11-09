package env

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"shipyard/requests"
	"shipyard/requests/uri"
)

func NewGetEnvironmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "environment [environment ID]",
		Aliases:      []string{"env"},
		Short:        "Get environment by ID",
		SilenceUsage: true,
		// Due to an issue in viper, bind the 'json' flag in PreRun for each command that uses
		// a flag name already bound to a sibling command.
		// See https://github.com/spf13/viper/issues/233#issuecomment-386791444
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlag("json", cmd.Flags().Lookup("json"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return handleGetEnvironmentByID(args[0])
			}
			return fmt.Errorf("Environment ID argument not provided")
		},
	}

	cmd.Flags().Bool("json", false, "JSON output")

	return cmd
}

func NewGetAllEnvironmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "environments",
		Aliases:      []string{"envs"},
		SilenceUsage: true,
		Short:        "Get all environments",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlag("name", cmd.Flags().Lookup("name"))
			viper.BindPFlag("org-name", cmd.Flags().Lookup("org-name"))
			viper.BindPFlag("repo-name", cmd.Flags().Lookup("repo-name"))
			viper.BindPFlag("branch", cmd.Flags().Lookup("branch"))
			viper.BindPFlag("pull-request-number", cmd.Flags().Lookup("pull-request-number"))
			viper.BindPFlag("deleted", cmd.Flags().Lookup("deleted"))
			viper.BindPFlag("page", cmd.Flags().Lookup("page"))
			viper.BindPFlag("page-size", cmd.Flags().Lookup("page-size"))
			viper.BindPFlag("json", cmd.Flags().Lookup("json"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleGetAllEnvironments()
		},
	}

	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("org-name", "", "Filter by org name")
	cmd.Flags().String("repo-name", "", "Filter by repo name")
	cmd.Flags().String("branch", "", "Filter by branch")
	cmd.Flags().String("pull-request-number", "", "Filter by pull request number")
	cmd.Flags().Bool("deleted", false, "Filter by deleted")
	cmd.Flags().Int("page", 0, "Page number requested")
	cmd.Flags().Int("page-size", 0, "Page size requested")
	cmd.Flags().Bool("json", false, "JSON output")

	return cmd
}

var ErrUnmarshalling = errors.New("failed to unmarshal environment(s)")

func handleGetAllEnvironments() error {
	client, err := requests.NewHTTPClient(os.Stdout)
	if err != nil {
		return err
	}

	params := make(map[string]string)

	if name := viper.GetString("name"); name != "" {
		params["name"] = name
	}
	if orgName := viper.GetString("org-name"); orgName != "" {
		params["org_name"] = orgName
	}
	if repoName := viper.GetString("repo-name"); repoName != "" {
		params["repo_name"] = repoName
	}
	if branch := viper.GetString("branch"); branch != "" {
		params["branch"] = branch
	}
	if pullRequestNumber := viper.GetString("pull-request-number"); pullRequestNumber != "" {
		params["pull_request_number"] = pullRequestNumber
	}
	if deleted := viper.GetBool("deleted"); deleted {
		params["deleted"] = "true"
	}
	if page := viper.GetInt("page"); page != 0 {
		params["page"] = strconv.Itoa(page)
	}
	if pageSize := viper.GetInt("page-size"); pageSize != 0 {
		params["page_size"] = strconv.Itoa(pageSize)
	}
	if org := viper.GetString("org"); org != "" {
		params["org"] = org
	}

	body, err := client.Do(http.MethodGet, uri.CreateResourceURI("", "environment", "", "", params), nil)
	if err != nil {
		return err
	}

	if viper.GetBool("json") {
		return client.Write(body)
	}

	r, err := unmarshalManyEnv(body)
	if err != nil {
		return ErrUnmarshalling
	}

	var data [][]string
	for _, d := range r.Data {
		pr := strconv.Itoa(d.Attributes.Projects[0].PullRequestNumber)
		if pr == "0" {
			pr = ""
		}

		data = append(data, []string{
			d.ID,
			d.Attributes.Projects[0].RepoName,
			d.Attributes.Name,
			pr,
			d.Attributes.URL,
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"UUID", "Repo", "App Name", "PR#", "URL"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	for _, v := range data {
		table.Append(v)
	}
	table.Render()

	return nil
}

func GetEnvironmentByID(client requests.Client, id string) (*Response, error) {
	params := make(map[string]string)
	org := viper.GetString("org")
	if org != "" {
		params["org"] = org
	}

	body, err := client.Do(http.MethodGet, uri.CreateResourceURI("", "environment", id, "", params), nil)
	if err != nil {
		return nil, err
	}

	return unmarshalEnv(body)
}

func handleGetEnvironmentByID(id string) error {
	client, err := requests.NewHTTPClient(os.Stdout)
	if err != nil {
		return err
	}

	params := make(map[string]string)
	org := viper.GetString("org")
	if org != "" {
		params["org"] = org
	}

	body, err := client.Do(http.MethodGet, uri.CreateResourceURI("", "environment", id, "", params), nil)
	if err != nil {
		return err
	}

	if viper.GetBool("json") {
		return client.Write(body)
	}

	r, err := unmarshalEnv(body)
	if err != nil {
		return err
	}

	env := r.Data
	pr := strconv.Itoa(env.Attributes.Projects[0].PullRequestNumber)
	if pr == "0" {
		pr = ""
	}

	data := [][]string{
		[]string{
			env.ID,
			env.Attributes.Projects[0].RepoName,
			env.Attributes.Name,
			pr,
			env.Attributes.URL,
		},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"UUID", "Repo", "App Name", "PR#", "URL"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	for _, v := range data {
		table.Append(v)
	}
	table.Render()

	return nil
}

func unmarshalEnv(p []byte) (*Response, error) {
	var r Response
	err := json.Unmarshal(p, &r)
	if err != nil {
		return nil, ErrUnmarshalling
	}
	return &r, err
}

func unmarshalManyEnv(p []byte) (*respManyEnvs, error) {
	var r respManyEnvs
	err := json.Unmarshal(p, &r)
	if err != nil {
		return nil, ErrUnmarshalling
	}
	return &r, err
}

type environment struct {
	Attributes struct {
		Name string `json:"name"`
		URL  string `json:"url"`

		Projects []struct {
			PullRequestNumber int    `json:"pull_request_number"`
			RepoName          string `json:"repo_name"`
		} `json:"projects"`

		Services map[string]interface{} `json:"services"`
	} `json:"attributes"`

	ID string `json:"id"`
}

type Response struct {
	Data struct {
		environment
	} `json:"data"`
}

type respManyEnvs struct {
	Data []struct {
		environment
	} `json:"data"`
}
