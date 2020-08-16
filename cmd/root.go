
package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var config Config

type Config struct {
	BigqueryProjects []string
	CacheFile        string
	CacheRefreshHour int
}

// rootCmd represents the root command
var rootCmd = &cobra.Command{
	Use:   "bqiam",
	Short: "bqiam is a tool for bigquery administrator",
	Long: `bqiam is a tool for bigquery administrator.
This tool provides following functionalities:
- dataset: returns a set of roles that the input user account has for each dataset
- user: returns a set of users who can access the input dataset
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".bqiam" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".bqiam")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Failed to read Config File:", viper.ConfigFileUsed())
		os.Exit(1)
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println("Failed to read Config File:", viper.ConfigFileUsed())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bqiam.toml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().IntP("refresh", "r", 24, "cache refresh threshold in hour (default is 24 hours)")
	err := viper.BindPFlag("CacheRefreshHour", rootCmd.PersistentFlags().Lookup("refresh")) // overwrite by flag if exists
	if err != nil {
		fmt.Println("Failed to bind flag 'refresh': ", err)
		os.Exit(1)
	}
}
