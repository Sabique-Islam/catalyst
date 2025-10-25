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
var authorName string
var license string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the catalyst framework",
	Long:  `Initializes the catalyst framework by creating a catalyst.yml file, which includes all the dependancies`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat("catalyst.yml"); err == nil {
			return fmt.Errorf("catalyst.yml already exists")
		}

		content, err := project.GenerateYAML(projectName, authorName, license)
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
	initCmd.Flags().StringVarP(&projectName, "project", "n", "my-catalyst-app", "Project name")
	initCmd.Flags().StringVarP(&authorName, "author", "a", "linus-torvalds", "Author name")
	initCmd.Flags().StringVarP(&license, "license", "l", "MIT", "License")
}
