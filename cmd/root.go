/*
Copyright © 2025 Sabique-Islam
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Sabique-Islam/catalyst/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "catalyst",
	Short: "A modern C build tool with dependency management",
	Long: `Catalyst is a modern build tool for C projects that simplifies
dependency management, compilation, and project setup.

Features:
  • Interactive project initialization
  • Automatic dependency scanning
  • Cross-platform package management
  • Simple build and run commands

Run 'catalyst' without arguments to launch the interactive menu,
or use one of the available commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand is provided, show the interactive menu
		return runInteractiveMenu()
	},
}

// runInteractiveMenu displays the main menu and executes the selected command
func runInteractiveMenu() error {
	for {
		choice, err := tui.RunMainMenu()
		if err != nil {
			return err
		}

		switch choice {
		case "Smart Init (Auto-detect & generate config)":
			if err := smartInitCmd.RunE(smartInitCmd, []string{}); err != nil {
				fmt.Printf("Error: Smart Init failed: %v\n\n", err)
			}
		case "Analyze (Show project structure)":
			if err := analyzeCmd.RunE(analyzeCmd, []string{}); err != nil {
				fmt.Printf("Error: Analyze failed: %v\n\n", err)
			}
		case "Init (Create catalyst.yml)":
			if err := initCmd.RunE(initCmd, []string{}); err != nil {
				fmt.Printf("Error: Init failed: %v\n\n", err)
			}
		case "Scan (Find dependencies)":
			if err := scanCmd.RunE(scanCmd, []string{}); err != nil {
				fmt.Printf("Error: Scan failed: %v\n\n", err)
			}
		case "Install (Install dependencies)":
			if err := installCmd.RunE(installCmd, []string{}); err != nil {
				fmt.Printf("Error: Install failed: %v\n\n", err)
			}
		case "Build":
			if err := buildCmd.RunE(buildCmd, []string{}); err != nil {
				fmt.Printf("Error: Build failed: %v\n\n", err)
			}
		case "Run":
			if err := runCmd.RunE(runCmd, []string{}); err != nil {
				fmt.Printf("Error: Run failed: %v\n\n", err)
			}
		case "Clean":
			if err := cleanCmd.RunE(cleanCmd, []string{}); err != nil {
				fmt.Printf("Error: Clean failed: %v\n\n", err)
			}
		case "Exit":
			fmt.Println("Goodbye!")
			return nil
		default:
			fmt.Printf("Unknown option: %s\n", choice)
		}
	}
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
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.catalyst.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

		// Search config in home directory with name ".catalyst" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".catalyst")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
