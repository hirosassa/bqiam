/*
Copyright © 2020 Hirohito Sasakawa

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
