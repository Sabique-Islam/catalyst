package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Sabique-Islam/catalyst/internal/tui"
	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║    Catalyst UI Interactive Demo     ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()

	// Test 1: Run the Main Menu
	fmt.Println("📋 Testing Main Menu...")
	fmt.Println()

	choice, err := tui.RunMainMenu()
	if err != nil {
		log.Fatalf("❌ Main menu error: %v", err)
	}

	fmt.Printf("\n✅ You selected: %s\n\n", choice)

	// Test 2: If user selected "Init", run the wizard
	if choice == "Init (Create catalyst.yml)" {
		fmt.Println("🧙 Running Init Wizard...")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		config, err := tui.RunInitWizard()
		if err != nil {
			log.Fatalf("❌ Init wizard error: %v", err)
		}

		// Display the generated configuration
		fmt.Println("\n╔══════════════════════════════════════╗")
		fmt.Println("║    Generated Configuration          ║")
		fmt.Println("╚══════════════════════════════════════╝")
		fmt.Println()

		yamlData, err := yaml.Marshal(config)
		if err != nil {
			log.Fatalf("❌ Failed to marshal config: %v", err)
		}

		fmt.Println(string(yamlData))

		// Write to file
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Print("💾 Save this to catalyst.yml? (y/n): ")

		var save string
		fmt.Scanln(&save)

		if save == "y" || save == "Y" {
			err = os.WriteFile("catalyst.yml", yamlData, 0644)
			if err != nil {
				log.Fatalf("❌ Failed to write file: %v", err)
			}
			fmt.Println("✅ Configuration saved to catalyst.yml!")
		} else {
			fmt.Println("ℹ️  Configuration not saved.")
		}
	} else if choice == "Exit" {
		fmt.Println("👋 Goodbye!")
	} else {
		fmt.Printf("ℹ️  In a real application, this would execute: %s\n", choice)
	}
}
