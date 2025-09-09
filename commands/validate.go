package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shipyard/shipyard-cli/pkg/client"
	"github.com/shipyard/shipyard-cli/pkg/requests/uri"
	"github.com/shipyard/shipyard-cli/pkg/types"
)

func NewValidateCmd(c client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "validate [file_name]",
		Short:        "Validate a Docker Compose file",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlag("json", cmd.Flags().Lookup("json"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleValidateCmd(c, args[0])
		},
	}

	cmd.Flags().Bool("json", false, "Print the complete JSON output")
	return cmd
}

func handleValidateCmd(c client.Client, filename string) error {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	params := make(map[string]string)
	if org := c.OrgLookupFn(); org != "" {
		params["org"] = org
	}

	url := uri.CreateResourceURI("", "compose", "", "validate", params)
	
	resp, err := c.Requester.Do(http.MethodPost, url, "text/plain", fileContent)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}

	validation, err := types.UnmarshalValidation(resp)
	if err != nil {
		return fmt.Errorf("failed to parse validation response: %w", err)
	}

	if viper.GetBool("json") {
		fmt.Println(string(resp))
		return nil
	}

	return displayValidationResults(validation)
}

func displayValidationResults(validation *types.ValidationResponse) error {
	if validation.Valid {
		fmt.Printf("✅ %s\n", validation.Message)
		if len(validation.Services) > 0 {
			fmt.Printf("Services found: %v\n", validation.Services)
		}
	} else {
		fmt.Printf("❌ %s\n", validation.Message)
	}

	if len(validation.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range validation.Errors {
			if err.Line > 0 {
				fmt.Printf("  Line %d, Column %d: %s", err.Line, err.Column, err.Message)
			} else if err.Service != "" {
				fmt.Printf("  Service '%s': %s", err.Service, err.Message)
			} else {
				fmt.Printf("  %s", err.Message)
			}
			if err.Code != "" {
				fmt.Printf(" (%s)", err.Code)
			}
			fmt.Println()
		}
	}

	if len(validation.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range validation.Warnings {
			if warning.Service != "" {
				fmt.Printf("  Service '%s': %s\n", warning.Service, warning.Message)
			} else {
				fmt.Printf("  %s\n", warning.Message)
			}
		}
	}

	return nil
}
