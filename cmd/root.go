package cmd

import (
	"fmt"
	"os"

	"cobra-template/internal/config"
	"cobra-template/pkg/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "course-registration",
	Short: "University Course Registration System",
	Long: `A comprehensive University Course Registration System built with Go.

This system provides:
- High-performance course registration API
- Real-time waitlist management
- Redis caching for optimal performance
- Concurrent registration handling
- Load testing capabilities
- Docker containerization support

Example usage:
  course-registration registration --port 8080    # Start registration server
  course-registration loadtest --concurrent 100   # Run load tests`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger.Init(verbose)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra-template.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra-template" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cobra-template")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	// Initialize config
	config.Init()
}
