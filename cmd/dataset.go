package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hirosassa/bqiam/metadata"
	"github.com/spf13/cobra"
	"gopkg.in/djherbis/times.v1"
)

// datasetCmd represents the dataset command
var datasetCmd = &cobra.Command{
	Use:   "dataset [user email (required)]",
	Short: "List datasets that the input user or service account has permissions",
	Long: `
This subcommand returns a list of datasets
that the input user or service account is able to access.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("user email is required")
		}
		return nil
	},
	RunE: runCmdDataset,
}

func runCmdDataset(cmd *cobra.Command, args []string) error {
	refreshCache(cmd) // refresh cache if needed

	var ms metadata.Metas
	if err := ms.Load(config.CacheFile); err != nil {
		return err
	}

	entity := args[0]
	for _, m := range ms.Metas {
		if m.Entity == entity {
			fmt.Println(m.Project, m.Dataset, m.Role)
		}
	}
	return nil
}

func refreshCache(cmd *cobra.Command) {
	isExpired, _ := checkCacheExpired(config.CacheFile) // ignore errors
	if isExpired {
		fmt.Printf("Refresh cache? (takes 30-60 sec) [y/n]")
		reader := bufio.NewReader(os.Stdin)
		res, err := reader.ReadString('\n')

		if err != nil || strings.TrimSpace(res) != "y" {
			fmt.Println("Skip refreshing.")
			return
		}

		runCmdCache(cmd, []string{})  // run cache coomand to refresh
	}
}

func checkCacheExpired(filename string) (bool, error) {
	t, err := times.Stat(filename)
	if err != nil {
		return false, fmt.Errorf("Failed to get file modified timestamp: err: %s", err)
	}

	timePassed := time.Now().Sub(t.ModTime()).Hours()
	return timePassed > float64(config.CacheRefresh), nil
}

func init() {
	rootCmd.AddCommand(datasetCmd)
}
