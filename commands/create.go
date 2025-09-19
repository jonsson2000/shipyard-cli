package commands

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shipyard/shipyard-cli/pkg/client"
	"github.com/shipyard/shipyard-cli/pkg/display"
	"github.com/shipyard/shipyard-cli/pkg/requests/uri"
	"github.com/shipyard/shipyard-cli/pkg/types"
)

func NewCreateCmd(c client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources",
	}
	cmd.AddCommand(NewCreateApplicationCmd(c))
	return cmd
}

func NewCreateApplicationCmd(c client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "application",
		Aliases:      []string{"app"},
		Short:        "Create a new application",
		SilenceUsage: true,
		Example: `  # Create an application with name and repo:
  shipyard create application --name my-app --repo-name my-repo --branch main

  # Create an application with JSON output:
  shipyard create application --name my-app --repo-name my-repo --json`,
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlag("name", cmd.Flags().Lookup("name"))
			_ = viper.BindPFlag("repo-name", cmd.Flags().Lookup("repo-name"))
			_ = viper.BindPFlag("branch", cmd.Flags().Lookup("branch"))
			_ = viper.BindPFlag("description", cmd.Flags().Lookup("description"))
			_ = viper.BindPFlag("json", cmd.Flags().Lookup("json"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleCreateApplicationCmd(c)
		},
	}

	cmd.Flags().String("name", "", "Name of the application")
	cmd.Flags().String("repo-name", "", "Repository name for the application")
	cmd.Flags().String("branch", "main", "Branch name (default: main)")
	cmd.Flags().String("description", "", "Optional description of the application")
	cmd.Flags().Bool("json", false, "JSON output")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("repo-name")

	return cmd
}

func handleCreateApplicationCmd(c client.Client) error {
	params := make(map[string]string)
	if org := c.OrgLookupFn(); org != "" {
		params["org"] = org
	}

	body := map[string]any{
		"name":        viper.GetString("name"),
		"repo_name":   viper.GetString("repo-name"),
		"branch":      viper.GetString("branch"),
		"description": viper.GetString("description"),
	}

	responseBody, err := c.Requester.Do(http.MethodPost, uri.CreateResourceURI("", "application", "", "", params), "application/json", body)
	if err != nil {
		return err
	}

	if viper.GetBool("json") {
		display.Println(responseBody)
		return nil
	}

	var response types.CreateApplicationResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		display.Println("Application created successfully")
		display.Println(string(responseBody))
		return nil
	}

	display.Println("Application created successfully:")
	data := [][]string{
		{"Name", response.Data.Attributes.Name},
		{"ID", response.Data.ID},
		{"Repository", response.Data.Attributes.RepoName},
		{"Branch", response.Data.Attributes.Branch},
		{"Status", response.Data.Attributes.Status},
	}
	if response.Data.Attributes.Description != "" {
		data = append(data, []string{"Description", response.Data.Attributes.Description})
	}

	columns := []string{"Field", "Value"}
	display.RenderTable(os.Stdout, columns, data)

	return nil
}
