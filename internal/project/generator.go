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
        Author   string `yaml:"author"`
        License  string `yaml:"license"`
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
