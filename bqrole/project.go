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
		fmt.Printf("Permit %s to %s access as %s\n", user, project, role)
	}

	return nil
}

func grantProjectRole(project, user, role string) error {
	policy, err := fetchCurrentPolicy(project)
	if err != nil {
		return fmt.Errorf("failed to fetch current policy: error %s", err)
	}

	if hasProjectRole(*policy, user, role) { // already has roles/viewer
		log.Info().Msgf("%s already has a role: %s, project: %s. skipped.", user, role, project)
		return nil
	}

	var member string
	if isServiceAccount(user) {
		member = "serviceAccount:" + user
	} else {
		member = "user:" + user
	}

	cmd := fmt.Sprintf("gcloud projects add-iam-policy-binding %s --member %s --role %s", project, member, role)
	err = exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		return fmt.Errorf("failed to update policy bindings to grant %s %s: error: %s", user, role, err)
	}

	return nil
}

func hasProjectRole(p ProjectPolicy, user, role string) bool {
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
