package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	bq "cloud.google.com/go/bigquery"
	"github.com/spf13/cobra"
)

func projectRole(role string) (string, error) {
	switch role {
	case READER:
		return "roles/reader", nil
	case WRITER:
		return "roles/writer", nil
	case OWNER:
		return "roles/owner", nil
	}

	return "", fmt.Errorf("failed to parse %s", role)
}

func permitpj(role, project string, users []string) error {
	ctx := context.Background()
	client, err := bq.NewClient(ctx, project)
	if err != nil {
		return errors.New("failed to create bigquery Client")
	}

	fmt.Printf("project_id: %s\n", project)
	fmt.Printf("role:       %s\n", role)
	fmt.Printf("users:      %s\n", users)
	fmt.Printf("If you proceeds, PROJECT-WIDE permission will be added. Are you sure? [y/n]")

	reader := bufio.NewReader(os.Stdin)
	res, err := reader.ReadString('\n')

	if err != nil || strings.TrimSpace(res) != "y" {
		fmt.Println("Abort.")
		return nil
	}

	defer client.Close()

	// grant project-wide role if needed
	for _, user := range users {
		err = grantProjectRole(project, user, role)
		if err != nil {
			return err
		}
	}

	return nil
}

func grantProjectRole(project, user, role string) error {
	policy, err := fetchCurrentPolicy(project)
	if err != nil {
		return fmt.Errorf("failed to fetch current policy: error %s", err)
	}

	if hasProjectRole(*policy, user, role) {
		return nil
	}

	cmd := fmt.Sprintf("gcloud projects add-iam-policy-binding %s --member user:%s --role %s", project, user, role)
	err = exec.Command(cmd).Run()
	if err != nil {
		return fmt.Errorf("failed to update policy bindings to grant roles/bigquery.jobUser: error: %s", err)
	}

	return nil
}

func hasProjectRole(p ProjectPolicy, user, role string) bool {
	for _, b := range p.Bindings {
		if b.Role == role {
			for _, m := range b.Members {
				if m == user {
					return true
				}
			}
		}
	}
	return false
}

// permitpjCmd represents the permitpj command
var permitpjCmd = &cobra.Command{
	Use:   "permitpj [READER | WRITER | OWNER] -p [bq-project-id (required)] -u [user(s) (required)]",
	Short: "permitpj permits some users to some project-wide access",
	Long: `permitpj some users to some project-wide access as READER or WRITER or OWNER
For example:

bqiam permitpj READER -p bq-project-id -u user1@email.com -u user2@email.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		role, err := projectRole(args[0])
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

		err = permitpj(role, project, users)

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
	rootCmd.AddCommand(permitpjCmd)

	permitCmd.Flags().StringP("project", "p", "", "Specify GCP project id")
	err := permitCmd.MarkFlagRequired("project")
	if err != nil {
		panic(err)
	}

	permitCmd.Flags().StringSliceP("users", "u", []string{}, "Specify user email(s)")
}
