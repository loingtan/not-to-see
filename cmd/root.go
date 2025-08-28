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
		cfg := config.Get()
		if err := logger.InitWithConfig(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output, cfg.Log.FilePath); err != nil {
			// Fallback to simple init if config-based init fails
			logger.Init(verbose)
			logger.Warn("Failed to initialize logger with config, using fallback: %v", err)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra-template.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	if cfgFile != "" {

		viper.SetConfigFile(cfgFile)
	} else {

		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cobra-template")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	config.Init()
}
