package project

import (
    "fmt"
    "time"

    "gopkg.in/yaml.v3"
)

type CatalystConfig struct {
    Project struct {
        Name    string `yaml:"name"`
        Version string `yaml:"version"`
        Created string `yaml:"created"`
    } `yaml:"project"`

    Settings struct {
        Language string `yaml:"language"`
        Author   string `yaml:"author"`
        License  string `yaml:"license"`
    } `yaml:"settings"`
}

func GenerateYAML(projectName string) (string, error) {
    cfg := CatalystConfig{}
    cfg.Project.Name = projectName
    cfg.Project.Version = "0.1.0"
    cfg.Project.Created = time.Now().Format(time.RFC3339)
    cfg.Settings.Language = "go"
    cfg.Settings.Author = "Saketh"
    cfg.Settings.License = "MIT"

    out, err := yaml.Marshal(&cfg)
    if err != nil {
        return "", fmt.Errorf("yaml marshal failed: %w", err)
    }
    return string(out), nil
}
