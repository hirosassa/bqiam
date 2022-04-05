package bqrole

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

type ProjectPolicy struct {
	Bindings []struct {
		Role    string   `json:"role"`
		Members []string `json:"members"`
	} `json:"bindings"`
	Etag    string `json:"etag"`
	Version int    `json:"version"`
}

func FetchCurrentPolicy(project string) (*ProjectPolicy, error) {
	cmd := fmt.Sprintf("gcloud projects get-iam-policy %s --format=json", project)
	log.Info().Msg(fmt.Sprintf("execute: %s", cmd))

	policyJson, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run gcloud command to get current iam policy: %s\n%s", err, err.(*exec.ExitError).Stderr)
	}

	log.Info().Msg(fmt.Sprintf("finish: %s", cmd))

	var policy ProjectPolicy
	err = json.Unmarshal(policyJson, &policy)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy json: error: %s", err)
	}

	return &policy, nil
}

func isServiceAccount(user string) bool {
	return strings.HasSuffix(user, "iam.gserviceaccount.com")
}
