/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Sabique-Islam/catalyst/internal/project"
	"github.com/spf13/cobra"
)

var projectName string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the catalyst framework",
	Long: `Initializes the catalyst framework by creating a catalyst.yml file, which includes all the dependancies`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat("catalyst.yml"); err == nil {
				return fmt.Errorf("catalyst.yml already exists")
		}

		content, err := project.GenerateYAML(projectName)
		if err != nil {
				return fmt.Errorf("failed to generate YAML: %w", err)
		}

		err = os.WriteFile("catalyst.yml", []byte(content), 0644)
		if err != nil {
				return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Println("✅ Created catalyst.yml")
		return nil
},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&projectName, "name", "n", "my-catalyst-app", "Project name")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
