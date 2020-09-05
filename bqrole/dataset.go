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
		log.Info().Msgf("%s already have bigquery.jobUser\n", user)
		return nil
	}

	var member string
	if isServiceAccount(user) {
		member = "serviceAccount:" + user
	} else {
		member = "user:" + user
	}

	cmd := fmt.Sprintf("gcloud projects add-iam-policy-binding %s --member %s --role roles/bigquery.jobUser", project, member)
	err = exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		return fmt.Errorf("failed to update policy bindings to grant %s roles/bigquery.jobUser: error: %s", user, err)
	}

	return nil
}

func hasBQJobUser(p ProjectPolicy, user string) bool {
	for _, b := range p.Bindings {
		if b.Role == "roles/bigquery.jobUser" {
			for _, m := range b.Members {
				if strings.HasSuffix(m, user) { // format of m is (user|serviceAccount):[user-email]
					return true
				}
			}
		}
	}
	return false
}
