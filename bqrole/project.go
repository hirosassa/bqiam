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
	"github.com/rs/zerolog/log"
)

func ProjectRole(role string) (string, error) {
	switch role {
	case READER:
		return "roles/viewer", nil
	case WRITER:
		return "roles/editor", nil
	}

	return "", fmt.Errorf("failed to parse %s", role)
}

func PermitProject(role, project string, users []string, yes bool) error {
	ctx := context.Background()
	client, err := bq.NewClient(ctx, project)
	if err != nil {
		return errors.New("failed to create bigquery Client")
	}
	defer client.Close()

	fmt.Printf("PERMIT following PROJECT-WIDE permission\n")
	fmt.Printf("project_id: %s\n", project)
	fmt.Printf("role:       %s\n", role)
	fmt.Printf("users:      %s\n", users)

	if !yes {
		fmt.Printf("If you proceeds, PROJECT-WIDE permission will be added. Are you sure? [y/n]")

		reader := bufio.NewReader(os.Stdin)
		res, err := reader.ReadString('\n')

		if err != nil || strings.TrimSpace(res) != "y" {
			fmt.Println("Abort.")
			return nil
		}
	}

	policy, err := FetchCurrentPolicy(project)
	if err != nil {
		return fmt.Errorf("failed to fetch current policy: %s", err)
	}

	// grant project-wide role if needed
	for _, user := range users {
		err = grantProjectRole(project, user, role, policy)
		if err != nil {
			return err
		}
		fmt.Printf("Permit %s to %s access as %s\n", user, project, role)
	}

	return nil
}

func RevokeProject(role, project string, users []string, yes bool) error {
	ctx := context.Background()
	client, err := bq.NewClient(ctx, project)
	if err != nil {
		return errors.New("failed to create bigquery Client")
	}
	defer client.Close()

	fmt.Printf("REVOKE following PROJECT-WIDE permission\n")
	fmt.Printf("project_id: %s\n", project)
	fmt.Printf("role:       %s\n", role)
	fmt.Printf("users:      %s\n", users)

	if !yes {
		fmt.Printf("If you proceeds, PROJECT-WIDE permission will be added. Are you sure? [y/n]")

		reader := bufio.NewReader(os.Stdin)
		res, err := reader.ReadString('\n')

		if err != nil || strings.TrimSpace(res) != "y" {
			fmt.Println("Abort.")
			return nil
		}
	}

	policy, err := FetchCurrentPolicy(project)
	if err != nil {
		return fmt.Errorf("failed to fetch current policy: %s", err)
	}

	// revoke project-wide role if needed
	for _, user := range users {
		err = revokeProjectRole(project, user, role, policy)
		if err != nil {
			return err
		}
		fmt.Printf("Revoked %s's permission of %s access as %s\n", user, project, role)
	}

	return nil
}

func grantProjectRole(project, user, role string, policy *ProjectPolicy) error {
	if hasProjectRole(policy, user, role) { // already has roles/viewer
		log.Info().Msgf("%s already has a role: %s, project: %s. skipped.", user, role, project)
		return nil
	}

	var member string
	if isServiceAccount(user) {
		member = "serviceAccount:" + user
	} else {
		member = "user:" + user
	}

	cmd := exec.Command("gcloud", "projects", "add-iam-policy-binding", project, "--member", member, "--role", role, "--condition=None")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update policy bindings to grant %s %s: %s\n%s", user, role, err, err.(*exec.ExitError).Stderr)
	}

	return nil
}

func revokeProjectRole(project, user, role string, policy *ProjectPolicy) error {
	if !hasProjectRole(policy, user, role) {
		log.Info().Msgf("%s doesn't have a role: %s, project: %s. skipped.", user, role, project)
		return nil
	}

	var member string
	if isServiceAccount(user) {
		member = "serviceAccount:" + user
	} else {
		member = "user:" + user
	}

	cmd := exec.Command("gcloud", "projects", "remove-iam-policy-binding", project, "--member", member, "--role", role)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update policy bindings to revoke %s %s: %s\n%s", user, role, err, err.(*exec.ExitError).Stderr)
	}

	return nil
}

func hasProjectRole(p *ProjectPolicy, user, role string) bool {
	for _, b := range p.Bindings {
		if b.Role == role {
			for _, m := range b.Members {
				if strings.HasSuffix(m, user) { // format of m is (user|service):[user-email]
					return true
				}
			}
		}
	}
	return false
}
