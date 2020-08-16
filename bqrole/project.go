package bqrole

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	bq "cloud.google.com/go/bigquery"
)

// This tool does not grant project OWNER permission for safety
const (
	VIEWER = "VIEWER"
	EDITOR = "EDITOR"
)

func ProjectRole(role string) (string, error) {
	switch role {
	case VIEWER:
		return "roles/viewer", nil
	case EDITOR:
		return "roles/editor", nil
	}

	return "", fmt.Errorf("failed to parse %s", role)
}

func PermitProject(role, project string, users []string) error {
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
