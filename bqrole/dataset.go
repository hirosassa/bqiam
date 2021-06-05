package bqrole

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"os"
	"strings"

	bq "cloud.google.com/go/bigquery"
	"github.com/rs/zerolog/log"
)

func DatasetRole(role string) (bq.AccessRole, error) {
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

func PermitDataset(role bq.AccessRole, project string, users, datasets []string) error {
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

	// grant roles/bigquery.jobUser and roles/bigquery.user if needed
	for _, user := range users {
		err = grantBQRole(project, user, "roles/bigquery.jobUser")
		if err != nil {
			return err
		}

		err = grantBQRole(project, user, "roles/bigquery.user")
		if err != nil {
			return err
		}
	}

	// grant permissions for each datasets
	for _, dataset := range datasets {
		for _, user := range users {
			err := updateDatasetMetadata(ctx, client, role, dataset, user, bq.UserEmailEntity)
			if err != nil {
				// try as group account
				log.Warn().Msgf("failed to permit using bq.UserEmailEntity, try bq.GroupEmailEnity")
				err = updateDatasetMetadata(ctx, client, role, dataset, user, bq.GroupEmailEntity)
				if err != nil {
					return err
				}
			}
			fmt.Printf("Permit %s to %s access as %s\n", user, dataset, role)
		}
	}

	return nil
}

// grantBQRole grants user roles/bigquery permission
func grantBQRole(project, user string, role string) error {
	policy, err := fetchCurrentPolicy(project)
	if err != nil {
		return fmt.Errorf("failed to fetch current policy: %s", err)
	}

	if hasBQRole(*policy, user, role) {
		log.Info().Msgf("%s already have %s\n", user, role)
		return nil
	}

	var member string
	if isServiceAccount(user) {
		member = "serviceAccount:" + user
	} else {
		member = "user:" + user
	}

	cmd := fmt.Sprintf("gcloud projects add-iam-policy-binding %s --member %s --role %s", project, member, role)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if strings.Contains(string(out), "INVALID_ARGUMENT") { // try to bind to "group" account
		member = "group:" + user
		cmd = fmt.Sprintf("gcloud projects add-iam-policy-binding %s --member %s --role %s", project, member, role)
		err = exec.Command("bash", "-c", cmd).Run()
	}

	if err != nil {
		return fmt.Errorf("failed to update policy bindings to grant %s %s: %s", user, role, err)
	}

	return nil
}

func updateDatasetMetadata(ctx context.Context, client *bq.Client, role bq.AccessRole, dataset string, user string, entityType bq.EntityType) error {
	ds := client.Dataset(dataset)
	meta, err := ds.Metadata(ctx)
	if err != nil {
		return err
	}

	update := bq.DatasetMetadataToUpdate{
		Access: append(meta.Access, &bq.AccessEntry{
			Role:       role,
			EntityType: entityType,
			Entity:     user,
		}),
	}

	if _, err := ds.Update(ctx, update, meta.ETag); err != nil {
		return err
	}
	return nil
}

func hasBQRole(p ProjectPolicy, user string, role string) bool {
	for _, b := range p.Bindings {
		if b.Role == role {
			for _, m := range b.Members {
				if strings.HasSuffix(m, user) { // format of m is (user|serviceAccount):[user-email]
					return true
				}
			}
		}
	}
	return false
}
