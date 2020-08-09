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
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"os"
	"strings"

	bq "cloud.google.com/go/bigquery"
	"github.com/spf13/cobra"
)

const (
	READER = "READER"
	WRITER = "WRITER"
	OWNER  = "OWNER"
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

	return "", fmt.Errorf("failed to parse %s", role)
}

func permit(role bq.AccessRole, project string, users, datasets []string) error {
	ctx := context.Background()
	client, err := bq.NewClient(ctx, project)
	if err != nil {
		return errors.New("failed to create bigquery Client")
	}

	fmt.Printf("project_id: %s\n", project)
	fmt.Printf("role:       %s\n", role)
	fmt.Printf("datasets:   %s\n", datasets)
	fmt.Printf("users:      %s\n", users)
	fmt.Printf("Are you sure? [y/n]")

	reader := bufio.NewReader(os.Stdin)
	res, err := reader.ReadString('\n')

	if err != nil || strings.TrimSpace(res) != "y" {
		fmt.Println("Abort.")
		return nil
	}

	defer client.Close()

	// grant roles/bigquery.jobUser if needed
	for _, user := range users {
		err = grantBQJobUser(project, user)
		if err != nil {
			return err
		}
	}

	// grant permissions for each datasets
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

// grantBQJobUser grants user roles/bigquery.jobUser permission to run job on BigQuery
func grantBQJobUser(project, user string) error {
	policy, err := fetchCurrentPolicy(project)
	if err != nil {
		return fmt.Errorf("failed to fetch current policy: error %s", err)
	}

	if hasBQJobUser(*policy, user) { // already has roles/bigquery.jobUser
		return nil
	}

	cmd := fmt.Sprintf("gcloud projects add-iam-policy-binding %s --member user:%s --role roles/bigquery.jobUser", project, user)
	err = exec.Command(cmd).Run()
	if err != nil {
		return fmt.Errorf("failed to update policy bindings to grant roles/bigquery.jobUser: error: %s", err)
	}

	return nil
}

func hasBQJobUser(p ProjectPolicy, user string) bool {
	for _, b := range p.Bindings {
		if b.Role == "roles/bigquery.jobUser" {
			for _, m := range b.Members {
				if m == user {
					return true
				}
			}
		}
	}
	return false
}

// permitCmd represents the permit command
var permitCmd = &cobra.Command{
	Use:   "permit [READER | WRITER | OWNER] -p [bq-project-id (required)] [flags]",
	Short: "permit some users to some datasets access",
	Long: `permit some users to some datasets access as READER or WRITER or OWNER
For example:

bqiam permit READER -p bq-project-id -u user1@email.com -u user2@email.com -d dataset1 -d dataset2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		role, err := accessRole(args[0])
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

		err = permit(role, project, users, datasets)

		if err != nil {
			return fmt.Errorf("failed to permit: %s", err)
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
	err := permitCmd.MarkFlagRequired("project")
	if err != nil {
		panic(err)
	}

	permitCmd.Flags().StringSliceP("users", "u", []string{}, "Specify user email(s)")
	permitCmd.Flags().StringSliceP("datasets", "d", []string{}, "Specify dataset(s)")
}
