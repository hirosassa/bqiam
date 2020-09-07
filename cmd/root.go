package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
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

var verbose, debug bool // for verbose and debug output

// rootCmd represents the root command
var rootCmd = &cobra.Command{
	Use:   "bqiam",
	Short: "bqiam is a tool for bigquery administrator",
	Long: `bqiam is a tool for bigquery administrator.
This tool provides easier IAM management functionalities
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

	logOutput() // set log level
}

func logOutput() {
	zerolog.SetGlobalLevel(zerolog.Disabled) // default: quiet mode
	switch {
	case verbose:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case debug:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bqiam.toml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().IntP("refresh", "r", 24, "cache refresh threshold in hour (default is 24 hours)")

	// for log output
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable varbose log output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug log output")

	err := viper.BindPFlag("CacheRefreshHour", rootCmd.PersistentFlags().Lookup("refresh")) // overwrite by flag if exists
	if err != nil {
		fmt.Println("Failed to bind flag 'refresh': ", err)
		os.Exit(1)
	}
}
