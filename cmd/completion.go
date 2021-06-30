package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hirosassa/bqiam/bqrole"
	"github.com/hirosassa/bqiam/completion"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const completionDisplaySizeLimit = 100

func init() {
	rootCmd.AddCommand(newCompletionCmd())
}

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generates shell completion scripts",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("creating completion file: %s\n", completionFilePath())
			if err := createCompletionFile(); err != nil {
				panic(err)
			}
		},
	}

	cmd.AddCommand(
		newCompletionBashCmd(),
		newCompletionZshCmd(),
	)

	return cmd
}

func newCompletionBashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bash",
		Short: "Generates bash completion scripts",
		Run: func(cmd *cobra.Command, args []string) {
			if err := createCompletionFile(); err != nil {
				panic(err)
			}
			if err := rootCmd.GenBashCompletion(os.Stdout); err != nil {
				panic(err)
			}
		},
	}

	return cmd
}

func newCompletionZshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zsh",
		Short: "Generates zsh completion scripts",
		Run: func(cmd *cobra.Command, args []string) {
			if err := createCompletionFile(); err != nil {
				panic(err)
			}
			if err := rootCmd.GenZshCompletion(os.Stdout); err != nil {
				panic(err)
			}
		},
	}

	return cmd
}

func completionFilePath() string {
	// Currently, the config by viper is loaded after loading commands.
	// Loading completion file is on loading commands, so we can't specify config file path by config viper.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return path.Join(home, ".bqiam-completion-file.toml")
}

func createCompletionFile() error {
	var list completion.List
	ctx := context.Background()
	list.Projects = config.BigqueryProjects

	for _, project := range list.Projects {
		datasets, err := listDataSets(ctx, project)
		if err != nil {
			return err
		}

		list.Datasets = append(list.Datasets, *datasets...)
	}

	for _, project := range list.Projects {
		policy, err := bqrole.FetchCurrentPolicy(project)
		if err != nil {
			return err
		}

		for _, b := range policy.Bindings {
			for _, m := range b.Members {
				splited := strings.Split(m, ":") // member format like ((user|serviceAccount):[user-email])
				if len(splited) > 1 {
					list.Users = append(list.Users, splited[1])
				}
			}

		}
	}

	list.DisplaySizeLimit = completionDisplaySizeLimit

	if err := list.Save(completionFilePath()); err != nil {
		return err
	}

	return nil
}

func registerDatasetsCompletions(cmd *cobra.Command) error {
	var list completion.List
	if err := list.Load(completionFilePath()); err != nil {
		return err
	}

	if err := cmd.RegisterFlagCompletionFunc("datasets", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var res []string

		for _, d := range list.Datasets {
			if strings.HasPrefix(d, toComplete) {
				res = append(res, d)
				if len(res) >= list.DisplaySizeLimit {
					break
				}
			}
		}

		return res, cobra.ShellCompDirectiveDefault
	}); err != nil {
		return err
	}
	return nil
}

func registerProjectsCompletions(cmd *cobra.Command) error {
	var list completion.List
	if err := list.Load(completionFilePath()); err != nil {
		return err
	}

	if err := cmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var res []string

		for _, p := range list.Projects {
			if strings.HasPrefix(p, toComplete) {
				res = append(res, p)
				if len(res) >= list.DisplaySizeLimit {
					break
				}
			}
		}

		return res, cobra.ShellCompDirectiveDefault
	}); err != nil {
		return err
	}
	return nil
}

func registerUsersCompletions(cmd *cobra.Command) error {
	var list completion.List
	if err := list.Load(completionFilePath()); err != nil {
		return err
	}

	if err := cmd.RegisterFlagCompletionFunc("users", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var res []string

		for _, u := range list.Users {
			if strings.HasPrefix(u, toComplete) {
				res = append(res, u)
				if len(res) >= list.DisplaySizeLimit {
					break
				}
			}
		}

		return res, cobra.ShellCompDirectiveDefault
	}); err != nil {
		return err
	}
	return nil
}
