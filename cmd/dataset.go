package cmd

import (
	"errors"
	"fmt"

	"github.com/hirosassa/bqiam/metadata"
	"github.com/spf13/cobra"
)

// datasetCmd represents the dataset command
var datasetCmd = &cobra.Command{
	Use:   "dataset [user email (required)]",
	Short: "List datasets that the input user or service account has permissions",
	Long: `
This subcommand returns a list of datasets
that the input user or service account is able to access.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("user email is required")
		}
		return nil
	},
	RunE: runCmdDataset,
}

func runCmdDataset(cmd *cobra.Command, args []string) error {
	var ms metadata.Metas
	if err := ms.Load(config.CacheFile); err != nil {
		return err
	}

	entity := args[0]
	for _, m := range ms.Metas {
		if m.Entity == entity {
			fmt.Println(m.Project, m.Dataset, m.Role)
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(datasetCmd)
}
