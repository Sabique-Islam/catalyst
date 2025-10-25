package tui

import (
"fmt"
"strings"

core "github.com/Sabique-Islam/catalyst/internal/config"
"github.com/manifoldco/promptui"
)

// RunMainMenu displays the main menu and returns the selected option
func RunMainMenu() (string, error) {
	prompt := promptui.Select{
		Label: "Select an option",
		Items: []string{"Build", "Run", "Clean", "Init (Create catalyst.yml)", "Exit"},
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

func RunInitWizard() (*core.Config, error) {
	cfg := &core.Config{
		Dependencies: make(map[string][]string),
		Resources:    []core.Resource{},
	}

	//Project Name
	projectPrompt := promptui.Prompt{
		Label: "Enter project name",
	}
	projectName, err := projectPrompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil, fmt.Errorf("operation cancelled by user")
		}
		return nil, fmt.Errorf("project name prompt failed: %v", err)
	}
	cfg.ProjectName = projectName

	//Sources (looping)
	fmt.Println("\n--- Source Files ---")
	sources := []string{}
	for {
		sourcePrompt := promptui.Prompt{
			Label: "Enter a source file (e.g., src/main.c) (leave empty to finish)",
		}
		source, err := sourcePrompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, fmt.Errorf("operation cancelled by user")
			}
			return nil, fmt.Errorf("source file prompt failed: %v", err)
		}

		source = strings.TrimSpace(source)
		if source == "" {
			break
		}
		sources = append(sources, source)
	}
	cfg.Sources = sources

	//Linux Dependencies (MultiSelect)
	fmt.Println("\n--- Linux Dependencies ---")
	linuxLibs := []string{"pthread", "m"}
	selectedLinux := []string{}

	// Note: promptui doesn't have native multi-select, so I will use a workaround
for _, lib := range linuxLibs {
confirmPrompt := promptui.Select{
Label: fmt.Sprintf("Add '%s' to Linux dependencies?", lib),
Items: []string{"Yes", "No"},
}
_, result, err := confirmPrompt.Run()
if err != nil {
if err == promptui.ErrInterrupt {
return nil, fmt.Errorf("operation cancelled by user")
}
return nil, fmt.Errorf("linux dependencies prompt failed: %v", err)
}
if result == "Yes" {
selectedLinux = append(selectedLinux, lib)
}
}
if len(selectedLinux) > 0 {
cfg.Dependencies["linux"] = selectedLinux
}

//Windows Dependencies (MultiSelect)
fmt.Println("\n--- Windows Dependencies ---")
windowsLibs := []string{"ws2_32.lib"}
selectedWindows := []string{}

for _, lib := range windowsLibs {
confirmPrompt := promptui.Select{
Label: fmt.Sprintf("Add '%s' to Windows dependencies?", lib),
Items: []string{"Yes", "No"},
}
_, result, err := confirmPrompt.Run()
if err != nil {
if err == promptui.ErrInterrupt {
return nil, fmt.Errorf("operation cancelled by user")
}
return nil, fmt.Errorf("windows dependencies prompt failed: %v", err)
}
if result == "Yes" {
selectedWindows = append(selectedWindows, lib)
}
}
if len(selectedWindows) > 0 {
cfg.Dependencies["windows"] = selectedWindows
}

//Resources (looping)
fmt.Println("\n--- Resources ---")
resources := []core.Resource{}
for {
urlPrompt := promptui.Prompt{
Label: "Add a resource URL? (e.g., for a data file) (leave empty to finish)",
}
url, err := urlPrompt.Run()
if err != nil {
if err == promptui.ErrInterrupt {
return nil, fmt.Errorf("operation cancelled by user")
}
return nil, fmt.Errorf("resource URL prompt failed: %v", err)
}

url = strings.TrimSpace(url)
if url == "" {
break
}

pathPrompt := promptui.Prompt{
Label: "Enter the local path to save it (e.g., assets/data.zip)",
}
path, err := pathPrompt.Run()
if err != nil {
if err == promptui.ErrInterrupt {
return nil, fmt.Errorf("operation cancelled by user")
}
return nil, fmt.Errorf("resource path prompt failed: %v", err)
}

resources = append(resources, core.Resource{
URL:  url,
Path: strings.TrimSpace(path),
})
}
cfg.Resources = resources

return cfg, nil
}
