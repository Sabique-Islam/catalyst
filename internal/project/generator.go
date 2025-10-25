package project

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type CatalystConfig struct {
	Project struct {
		Name    string `yaml:"name"`
		Created string `yaml:"created"`
	} `yaml:"project"`

	Settings struct {
		Author  string `yaml:"author"`
		License string `yaml:"license"`
	} `yaml:"settings"`
}

func GenerateYAML(projectName string, authorName string, license string) (string, error) {
	cfg := CatalystConfig{}
	cfg.Project.Name = projectName
	cfg.Project.Created = time.Now().Format(time.RFC3339)
	cfg.Settings.Author = authorName
	cfg.Settings.License = license

	out, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", fmt.Errorf("yaml marshal failed: %w", err)
	}
	return string(out), nil
}

// InitializeProject runs the interactive project initialization wizard
func InitializeProject() error {
	tui := getTUIInterface()
	config, automate, err := tui.RunInitWizard()
	if err != nil {
		return fmt.Errorf("initialization wizard failed: %w", err)
	}

	// Save the configuration
	if err := config.Save("catalyst.yml"); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Project initialized successfully!\n")
	if automate {
		fmt.Println("🔧 Automation enabled - dependencies will be installed automatically")
	}

	return nil
}

// TUIInterface allows for easier testing by making TUI calls mockable
type TUIInterface interface {
	RunInitWizard() (ConfigInterface, bool, error)
}

type ConfigInterface interface {
	Save(filename string) error
}

// getRealTUI returns the real TUI implementation
func getTUIInterface() TUIInterface {
	return &realTUI{}
}

type realTUI struct{}

func (r *realTUI) RunInitWizard() (ConfigInterface, bool, error) {
	// We'll need to import and use the actual tui.RunInitWizard
	// For now, return a placeholder
	return &dummyConfig{}, false, fmt.Errorf("TUI wizard not yet integrated - please create catalyst.yml manually")
}

type dummyConfig struct{}

func (d *dummyConfig) Save(filename string) error {
	return nil
}
