package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Sabique-Islam/catalyst/internal/tui"
	"gopkg.in/yaml.v3"
)

func main() {
	choice, err := tui.RunMainMenu()
	if err != nil {
		log.Fatalf("Main menu error: %v", err)
	}

	fmt.Printf("\nYou selected: %s\n\n", choice)

	// Test 2: If user selected "Init", run the wizard
	if choice == "Init (Create catalyst.yml)" {
		fmt.Println("Running Init Wizard...")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		config, automate, err := tui.RunInitWizard()
		if err != nil {
			log.Fatalf("Init wizard error: %v", err)
		}

		// Handle automation preference
		fmt.Println()
		if automate {
			fmt.Println("Automation Mode Selected")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("Scanning for dependencies...")
		} else {
			fmt.Println("Manual Mode Selected")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		}

		yamlData, err := yaml.Marshal(config)
		if err != nil {
			log.Fatalf("Failed to marshal config: %v", err)
		}

		fmt.Println(string(yamlData))

		// Write to file
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Print("Save this to catalyst.yml? (y/n): ")

		var save string
		fmt.Scanln(&save)

		if save == "y" || save == "Y" {
			err = os.WriteFile("catalyst.yml", yamlData, 0644)
			if err != nil {
				log.Fatalf("Failed to write file: %v", err)
			}
			fmt.Println("Configuration saved to catalyst.yml!")
		} else {
			fmt.Println("Configuration not saved.")
		}
	} else if choice == "Exit" {
		fmt.Println("Goodbye!")
	} else {
		fmt.Printf("In a real application, this would execute: %s\n", choice)
	}
}
