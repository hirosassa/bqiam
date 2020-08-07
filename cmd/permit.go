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
	"context"
	"errors"
	"fmt"
	bq "cloud.google.com/go/bigquery"

	"github.com/spf13/cobra"
)

const (
	READER = "READER"
	WRITER = "WRITER"
	OWNER = "OWNER"
)

func accessRole(role string) (bq.AccessRole, error) {
	switch role {
	case READER:
		return bq.ReaderRole, nil
	case WRITER:
		return bq.WriterRole, nil
	case OWNER:
		return bq.OwnerRole, nil
	}

	return "", errors.New(fmt.Sprintf("failed to parse %s", role))
}

func permit(role bq.AccessRole, project string, users, datasets []string) error {
	ctx := context.Background()
	client, err := bq.NewClient(ctx, project)
	if err != nil {
		return errors.New("failed to create bigquery Client")
	}

	defer client.Close()

	for _, dataset := range datasets {
		for _, user := range users {
			ds := client.Dataset(dataset)
			meta, err := ds.Metadata(ctx)
			if err != nil {
				return err
			}

			update := bq.DatasetMetadataToUpdate{
				Access: append(meta.Access, &bq.AccessEntry{
					Role:       role,
					EntityType: bq.UserEmailEntity,
					Entity:     user,
				}),
			}

			if _, err := ds.Update(ctx, update, meta.ETag); err != nil {
				return err
			}

			fmt.Printf("Permit %s to %s access as %s\n", user, dataset, role)
		}
	}

	return nil
}

// permitCmd represents the permit command
var permitCmd = &cobra.Command{
	Use:   "permit [READER | WRITER | OWNER] -p [bq-project-id (required)] [flags]",
	Short: "permit some users to some datasets access",
	Long: `permit some users to some datasets access as READER or WRITER or OWNER
For example:

./bqiam permit READER -p bq-project-id -u user1@email.com -u user2@email.com -d dataset1 -d dataset2

to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		role, err := accessRole(args[0])
		if err != nil {
			return errors.New(fmt.Sprintf("READER or WRITER or OWNER must be specified: %s", err))
		}

		project, err := cmd.Flags().GetString("project")
		if err != nil {
			return errors.New(fmt.Sprintf("failed to parse project flag: %s", err))
		}

		users, err := cmd.Flags().GetStringSlice("users")
		if err != nil {
			return errors.New(fmt.Sprintf("failed to parse users flag: %s", err))
		}

		datasets, err := cmd.Flags().GetStringSlice("datasets")
		if err != nil {
			return errors.New(fmt.Sprintf("failed to parse datasets flag: %s", err))
		}

		err = permit(role, project, users, datasets)

		if err != nil {
			return errors.New(fmt.Sprintf("failed to permit: %s", err))
		}

		return nil
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("READER or WRITER or OWNER must be specified")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(permitCmd)

	permitCmd.Flags().StringP("project", "p", "", "Specify GCP project id")
	permitCmd.MarkFlagRequired("project")

	permitCmd.Flags().StringSliceP("users", "u", []string{}, "Specify user email(s)")
	permitCmd.Flags().StringSliceP("datasets", "d", []string{}, "Specify dataset(s)")
}
