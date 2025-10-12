package config

import (
    "encoding/json"
    "fmt"
    "os"
)

type Projects map[string]Project
type Project struct {
	Name string `json:"name"`
	Path string `json:"path"`
	GitWorkTree bool `json:"git_work_tree"`
}
func SaveProjects(filename string, projects Projects) error {
    data, err := json.MarshalIndent(projects, "", "  ")
    if err != nil {
        return fmt.Errorf("errore nel marshal JSON: %w", err)
    }

    return os.WriteFile(filename, data, 0644)
}

func LoadProjects(filename string) (Projects, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("errore nella lettura file: %w", err)
    }

    var projects Projects
    if err := json.Unmarshal(data, &projects); err != nil {
        return nil, fmt.Errorf("errore nel parse JSON: %w", err)
    }

    return projects, nil
}
