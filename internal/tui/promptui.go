package tui

import (
	"fmt"
	"os"

	core "github.com/Sabique-Islam/catalyst/internal/config"
	"github.com/manifoldco/promptui"
)

// RunMainMenu displays the main menu and returns the selected option
func RunMainMenu() (string, error) {
	prompt := promptui.Select{
		Label: "Select an option",
		Items: []string{
			"Init (Create catalyst.yml)",
			"Scan (Find dependencies)",
			"Install (Install dependencies)",
			"Build",
			"Run",
			"Clean",
			"Exit",
		},
	}

	_, result, err := prompt.Run()

	if err != nil {
		if err == promptui.ErrInterrupt {
			return "", fmt.Errorf("operation cancelled by user")
		}
		return "", fmt.Errorf("prompt failed: %v", err)
	}

	return result, nil
}

// RunInitWizard guides the user through creating a new catalyst.yml configuration
// Returns: (*core.Config, automate bool, error)
func RunInitWizard() (*core.Config, bool, error) {
	cfg := &core.Config{}

	// Step 1: Get Project Name
	projectPrompt := promptui.Prompt{
		Label: "Enter project name",
	}
	projectName, err := projectPrompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil, false, fmt.Errorf("operation cancelled by user")
		}
		return nil, false, fmt.Errorf("project name prompt failed: %v", err)
	}
	cfg.ProjectName = projectName

	// Step 2: Get Automation Preference
	automationPrompt := promptui.Select{
		Label: "How do you want to handle dependencies?",
		Items: []string{
			"Automate (Recommended) - Scans all .c and .h files for #include statements",
			"Manual - You add dependencies to catalyst.yml yourself",
		},
	}

	idx, _, err := automationPrompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil, false, fmt.Errorf("operation cancelled by user")
		}
		return nil, false, fmt.Errorf("automation preference prompt failed: %v", err)
	}

	// Return the config and automation preference
	// If automate is true, the caller will handle scanning and dependency detection
	// If automate is false, the caller will handle creating manual instructions
	automate := (idx == 0)

	// If automation is enabled, ask for an optional entry point (main source file)
	if automate {
		entryPrompt := promptui.Prompt{
			Label: "Entry point (path to main source file) — leave blank to auto-scan",
			Validate: func(input string) error {
				if input == "" {
					return nil
				}
				// Check file exists
				if _, err := os.Stat(input); err != nil {
					return fmt.Errorf("file does not exist: %v", err)
				}
				return nil
			},
		}

		entry, err := entryPrompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, false, fmt.Errorf("operation cancelled by user")
			}
			return nil, false, fmt.Errorf("entry point prompt failed: %v", err)
		}

		if entry != "" {
			// Set the entry point as the sole source in the config; generator will respect this
			cfg.Sources = []string{entry}
		}
	}

	return cfg, automate, nil
}
