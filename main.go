package main

import (
    "os"
    "fmt"
    "sort"
    "errors"
    "strings"
    "path/filepath"
    "workforge/config"
    "workforge/terminal"

    "github.com/spf13/cobra"
    "github.com/ktr0731/go-fuzzyfinder"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "wf",
        Short: "Workforge - Forge your work",
    }
    // --- add command flags ---
    var addNewBranch bool
    var addPrefix string
	var gwtFlag bool
	var initCmd = &cobra.Command{
		Use:   "init <url> <path>",
		Short: "Initialize a project",
		Args:  cobra.RangeArgs(0, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var path *string
			var url string
			var err error
			var entries []os.DirEntry	
			var repo_name string
			if len(args) > 0 {
				url = args[0]
				tmp := strings.Split(url, "/")
				repo_name = strings.TrimSuffix(tmp[len(tmp)-1], ".git")
				if len(args) > 1 {
					path = &args[1]
					entries, err = os.ReadDir(*path)
				} else {
					entries, err = os.ReadDir("./")
				}

				if err != nil {
					fmt.Println("directory error:", err)
					return
				}
				if len(entries) > 0 {
					if !gwtFlag {
						for _, entry := range entries {
							if entry.Name() == config.ConfigFileName{
								fmt.Println("This is a Workforge directory")
								fmt.Println("You can't clone a new repo here")
								return
							}
						}
					}else{
						fmt.Println("Directory not empty, aborting")
						return
					}

				}
				terminal.GitClone(url, path)
			}
			
			if gwtFlag {
				config.WriteExampleConfig(path)
			}else{
				if path == nil {
					path = &repo_name
				} else {
					*path = *path + "/" + repo_name
				}
				fmt.Println("Project path:", *path)
				config.WriteExampleConfig(path)
			}
			config.AddWorkforgePrj(repo_name, path, gwtFlag)
			fmt.Println("Project added to Workforge")
		},
	}

	initCmd.Flags().BoolVarP(&gwtFlag, "gwt", "t", false, "Use Git worktree")

	var loadCmd = &cobra.Command{
		Use:   "load [dir]",
		Short: "Load a Workforge project",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			path := "../"
			var profile *string
			if len(args) > 0 {
				path = path + args[0]
			}
			config.LoadProject(path,false, profile)
		},
	}
	type projItem struct {
		config.Project
		IsGWT bool
	}

	var openCmd = &cobra.Command{
		Use:   "open",
		Short: "Fuzzy finder to open a Workforge project",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			projs, hitmap, err := config.ListProjectsExpanded()
			if err != nil {
				fmt.Println("error loading projects:", err)
				return
			}
			if len(projs) == 0 {
				return
			}

			items := make([]projItem, 0, len(projs))
			for name, p := range projs {
				if p.Name == "" {
					p.Name = name
				}
				items = append(items, projItem{Project: p, IsGWT: hitmap[name]})
			}
			sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })

			idx, err := fuzzyfinder.Find(
				items,
				func(i int) string {
					if items[i].IsGWT {
						return "  " + items[i].Name
					}
					return "  " + items[i].Name
				},
				fuzzyfinder.WithPromptString(" Select project > "),
			)
			if err != nil {
				if errors.Is(err, fuzzyfinder.ErrAbort) {
					return
				}
				fmt.Println("fuzzy error:", err)
				return
			}

			chosen := items[idx]
			config.LoadProject(chosen.Path,chosen.IsGWT, nil)
		},
	}

    rootCmd.AddCommand(initCmd, loadCmd, openCmd)

    var addCmd = &cobra.Command{
        Use:   "add <name> [base-branch]",
        Short: "Add a worktree or create a new branch",
        Args:  cobra.RangeArgs(1, 2),
        Run: func(cmd *cobra.Command, args []string) {
            name := args[0]
            base := "main"
            if len(args) > 1 {
                base = args[1]
            }
            if addNewBranch {
                if addPrefix == "" {
                    addPrefix = "feature"
                }
                if err := terminal.AddNewWorkTree(name, addPrefix, base, true); err != nil {
                    fmt.Println("error creating new worktree:", err)
                    return
                }
                // Register the new worktree in Workforge registry
                cwd, err := os.Getwd()
                if err != nil {
                    fmt.Println("error getting current directory:", err)
                    return
                }
                leafAbs := filepath.Join(cwd, "..", name)
                if err := config.AddWorkforgeLeaf(leafAbs); err != nil {
                    fmt.Println("error registering worktree:", err)
                }
                return
            }
            // Use an existing branch as worktree
            if err := terminal.AddWorkTree(name); err != nil {
                fmt.Println("error adding worktree:", err)
                return
            }
            // Register the worktree in Workforge registry
            cwd, err := os.Getwd()
            if err != nil {
                fmt.Println("error getting current directory:", err)
                return
            }
            folderName := strings.ReplaceAll(name, "/", "-")
            leafAbs := filepath.Join(cwd, "..", folderName)
            if err := config.AddWorkforgeLeaf(leafAbs); err != nil {
                fmt.Println("error registering worktree:", err)
                return
            }
        },
    }

    addCmd.Flags().BoolVarP(&addNewBranch, "b", "b", false, "Create a new branch and worktree (optional base-branch, default: main)")
    addCmd.Flags().StringVar(&addPrefix, "prefix", "feature", "Branch prefix (default: feature)")
    rootCmd.AddCommand(addCmd)

    // wf rm <name>
    var rmCmd = &cobra.Command{
        Use:   "rm <name>",
        Short: "Remove a worktree (and unregister it)",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            name := args[0]
            // Determine candidate paths: ../name or ../(name with slashes -> dashes)
            cwd, err := os.Getwd()
            if err != nil {
                fmt.Println("error getting current directory:", err)
                return
            }
            cand1 := filepath.Join(cwd, "..", name)
            cand2 := filepath.Join(cwd, "..", strings.ReplaceAll(name, "/", "-"))

            leafPath := ""
            if st, err := os.Stat(cand1); err == nil && st.IsDir() {
                leafPath = cand1
            } else if st, err := os.Stat(cand2); err == nil && st.IsDir() {
                leafPath = cand2
            } else {
                fmt.Println("could not locate worktree directory for:", name)
                return
            }

            // Run on_delete hooks (non-fatal if config missing)
            if err := config.RunOnDelete(leafPath, true, nil); err != nil {
                fmt.Println("on_delete hook error:", err)
                return
            }

            // Remove the worktree via git
            if err := terminal.RunSyncCommand("git", "worktree", "remove", leafPath); err != nil {
                fmt.Println("error removing worktree:", err)
                return
            }

            // Unregister from Workforge registry
            if err := unregisterLeafFromRegistry(cwd, leafPath); err != nil {
                fmt.Println("error unregistering worktree:", err)
                return
            }

            fmt.Println("Worktree removed and unregistered:", leafPath)
        },
    }
    rootCmd.AddCommand(rmCmd)
    rootCmd.Execute()
}

// unregisterLeafFromRegistry removes the leaf from the Workforge registry using the same keying
// logic as AddWorkforgeLeaf.
func unregisterLeafFromRegistry(basePath string, leafAbs string) error {
    workforgePath := os.Getenv("HOME") + "/" + config.WORK_FORGE_PRJ_CONFIG_DIR
    projects, err := config.LoadProjects(workforgePath + "/" + config.WORK_FORGE_PRJ_CONFIG_FILE)
    if err != nil {
        return fmt.Errorf("failed to load existing projects: %w", err)
    }
    // Find base name
    var baseName string
    for name, p := range projects {
        if p.GitWorkTree && p.Path == basePath {
            baseName = name
            break
        }
    }
    leafName := filepath.Base(leafAbs)

    // Try to delete using base/key form first
    if baseName != "" {
        key := baseName + "/" + leafName
        if _, ok := projects[key]; ok {
            delete(projects, key)
            return config.SaveProjects(workforgePath+"/"+config.WORK_FORGE_PRJ_CONFIG_FILE, projects)
        }
    }
    // Fallback: delete plain leaf key
    if _, ok := projects[leafName]; ok {
        delete(projects, leafName)
        return config.SaveProjects(workforgePath+"/"+config.WORK_FORGE_PRJ_CONFIG_FILE, projects)
    }
    return fmt.Errorf("worktree entry not found in registry: %s", leafName)
}
