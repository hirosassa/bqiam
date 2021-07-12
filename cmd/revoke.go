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

	"github.com/spf13/cobra"

	"github.com/hirosassa/bqiam/bqrole"
)

func init() {
	rootCmd.AddCommand(newRevokeCommand())
}

func newRevokeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "revokes some users to some access",
		Long: `revokes some users to some datasets or project-wide access as READER or WRITER or OWNER
For example:

bqiam revoke dataset READER -p bq-project-id -u user1@email.com -u user2@email.com -d dataset1 -d dataset2
bqiam revoke project READER -p bq-project-id -u user1@email.com
`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(
		newRevokeDatasetCmd(),
	)

	return cmd
}

func newRevokeProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project [READER | WRITER] -p [bq-project-id (required)] -u [user(s) (required)]",
		Short: "revokes some users to some project-wide access",
		Long: `revoke project revokes some users to some project-wide access as READER or WRITER or OWNER
For example:

bqiam project READER -p bq-project-id -u user1@email.com -u user2@email.com`,
		RunE: runRevokeProjectCmd,
	}

	cmd.Flags().StringP("project", "p", "", "Specify GCP project id")
	err := cmd.MarkFlagRequired("project")
	if err != nil {
		panic(err)
	}

	cmd.Flags().StringSliceP("users", "u", []string{}, "Specify user email(s)")

	return cmd
}

func runRevokeProjectCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("READER or WRITER must be specified")
	}

	role, err := bqrole.ProjectRole(args[0])
	if err != nil {
		return fmt.Errorf("READER or WRITER must be specified: %s", err)
	}

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return fmt.Errorf("failed to parse project flag: %s", err)
	}

	users, err := cmd.Flags().GetStringSlice("users")
	if err != nil {
		return fmt.Errorf("failed to parse users flag: %s", err)
	}

	err = bqrole.RevokeProject(role, project, users)
	if err != nil {
		return fmt.Errorf("failed to revoke: %s", err)
	}

	return nil
}

func newRevokeDatasetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset [READER | WRITER | OWNER] -p [bq-project-id (required)] [flags]",
		Short: "revokes some users to some datasets access",
		Long: `revokes some users to some datasets access as READER or WRITER or OWNER
For example:

bqiam dataset READER -p bq-project-id -u user1@email.com -u user2@email.com -d dataset1 -d dataset2`,
		RunE: runRevokeDatasetCmd,
	}

	cmd.Flags().StringP("project", "p", "", "Specify GCP project id")
	err := cmd.MarkFlagRequired("project")
	if err != nil {
		panic(err)
	}

	cmd.Flags().StringSliceP("users", "u", []string{}, "Specify user email(s)")
	cmd.Flags().StringSliceP("datasets", "d", []string{}, "Specify dataset(s)")

	return cmd
}

func runRevokeDatasetCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("READER or WRITER or OWNER must be specified")
	}

	role, err := bqrole.DatasetRole(args[0])
	if err != nil {
		return fmt.Errorf("READER or WRITER or OWNER must be specified: %s", err)
	}

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return fmt.Errorf("failed to parse project flag: %s", err)
	}

	users, err := cmd.Flags().GetStringSlice("users")
	if err != nil {
		return fmt.Errorf("failed to parse users flag: %s", err)
	}

	datasets, err := cmd.Flags().GetStringSlice("datasets")
	if err != nil {
		return fmt.Errorf("failed to parse datasets flag: %s", err)
	}

	err = bqrole.RevokeDataset(role, project, users, datasets)
	if err != nil {
		return fmt.Errorf("failed to revoke: %s", err)
	}

	return nil
}
