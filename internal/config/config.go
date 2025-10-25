package core

// Resource defines a file to be downloaded
type Resource struct {
	URL  string `yaml:"url"`
	Path string `yaml:"path"`
}

// Config is the main project configuration
type Config struct {
	ProjectName  string              `yaml:"project_name"`
	Sources      []string            `yaml:"sources"`
	Dependencies map[string][]string `yaml:"dependencies"`
	Resources    []Resource          `yaml:"resources"`
}
