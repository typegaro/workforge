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
    rootCmd.Execute()
}
