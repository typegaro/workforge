package config

import (
    "encoding/json"
    "fmt"
    "os"
	"path/filepath"
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
func ListProjects() (Projects, error) {
	workforgePath := os.Getenv("HOME") + "/" + WORK_FORGE_PRJ_CONFIG_DIR
	if _, err := os.Stat(workforgePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workforge config directory does not exist")
	}
	projects, err := LoadProjects(workforgePath + "/" + WORK_FORGE_PRJ_CONFIG_FILE)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing projects: %w", err)
	}
	return projects, nil
}

// ListProjectsExpanded espande i progetti con GitWorkTree=true in tutte le subdir (solo directory, depth 1).
// Restituisce:
// - Projects "flattened": progetti normali + subdir dei GWT (senza la base GWT).
// - hitmap: true per gli elementi che sono subdir GWT.
func ListProjectsExpanded() (Projects, map[string]bool, error) {
	base, err := ListProjects()
	if err != nil {
		return nil, nil, err
	}

	out := make(Projects)
	hitmap := make(map[string]bool)

	for _, p := range base {
		if !p.GitWorkTree {
			// Progetto normale: includilo così com'è
			out[p.Name] = p
			hitmap[p.Name] = false
			continue
		}

		// Progetto GWT: elenca solo le subdir (niente base)
		entries, err := os.ReadDir(p.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("errore lettura GWT path %q: %w", p.Path, err)
		}

		found := false
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			found = true
			subName := p.Name + "/" + e.Name()
			subPath := filepath.Join(p.Path, e.Name())
			out[subName] = Project{
				Name:        subName,
				Path:        subPath,
				GitWorkTree: false, // il leaf non è a sua volta GWT
			}
			hitmap[subName] = true // <- è una subdir proveniente da GWT
		}

		// Se non ci sono subdir, non aggiungiamo la base (come richiesto)
		if !found {
			// opzionale: potresti voler segnalare/avvisare
			// fmt.Fprintf(os.Stderr, "attenzione: nessuna subdir in %s\n", p.Path)
		}
	}

	return out, hitmap, nil
}
