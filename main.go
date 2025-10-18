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
    var addNewBranch bool
    var addPrefix string
	var gwtFlag bool
	var initCmd = &cobra.Command{
		Use:   "init <url> <path>",
		Short: "Initialize a project",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var path string
			var url string
			var err error
			var entries []os.DirEntry	
			var repo_name string
			if len(args) > 0 {
				url = args[0]
				repo_name = RepoUrlToName(url)
				entries, err = os.ReadDir("./")
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
				terminal.GitClone(url, &path)
				if gwtFlag {
					config_file_path := "./"+ repo_name+"/"+ config.ConfigFileName
			    	_, err = os.Stat(config_file_path)
					if err == nil {
						fmt.Println("Coping wf config from the cloned repo")
						CopyFile(config_file_path, config.ConfigFileName)
					}else{
						config.WriteExampleConfig(&path)
					}
				}else{
					config.WriteExampleConfig(&path)
				}
				config.AddWorkforgePrj(repo_name, gwtFlag)
			} else {
				fmt.Println("Initializing a new Workforge project")
    			cwd, err := os.Getwd()
				if err != nil {
					fmt.Println("error getting current directory:", err)
					return 
				}
				repo_name = filepath.Base(cwd)
				config.WriteExampleConfig(nil)
				config.AddWorkforgePrj(repo_name, gwtFlag)
			}
		},
	}

	initCmd.Flags().BoolVarP(&gwtFlag, "gwt", "t", false, "Use Git worktree")

	var loadProfile string
	var loadCmd = &cobra.Command{
		Use:   "load [dir]",
		Short: "Load a Workforge project",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			path := "../"
			if len(args) > 0 {
				path = path + args[0]
			}
			if loadProfile != "" {
				config.LoadProject(path, false, &loadProfile)
				return
			}
			config.LoadProject(path, false, nil)
		},
	}
	loadCmd.Flags().StringVarP(&loadProfile, "profile", "p", "", "Profile name to use")
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
			// Inspect profiles for the selected project and prompt if multiple
			cfg, err := config.LoadConfig(chosen.Path, chosen.IsGWT)
			if err != nil {
				fmt.Println("error loading config:", err)
				config.LoadProject(chosen.Path, chosen.IsGWT, nil)
				return
			}

			profiles := make([]string, 0, len(cfg))
			for name := range cfg {
				profiles = append(profiles, name)
			}
			sort.Strings(profiles)

			selectedProfile := ""
			if len(profiles) > 1 {
				pidx, err := fuzzyfinder.Find(
					profiles,
					func(i int) string { return profiles[i] },
					fuzzyfinder.WithPromptString(" Select profile > "),
				)
				if err != nil {
					if errors.Is(err, fuzzyfinder.ErrAbort) {
						return
					}
					fmt.Println("fuzzy error:", err)
					return
				}
				selectedProfile = profiles[pidx]
			} else if len(profiles) == 1 {
				selectedProfile = profiles[0]
			} else {
				selectedProfile = config.DefaultProfile
			}

			if selectedProfile != "" {
				config.LoadProject(chosen.Path, chosen.IsGWT, &selectedProfile)
				return
			}
			config.LoadProject(chosen.Path, chosen.IsGWT, nil)
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
        },
    }

    addCmd.Flags().BoolVarP(&addNewBranch, "b", "b", false, "Create a new branch and worktree (optional base-branch, default: main)")
    addCmd.Flags().StringVar(&addPrefix, "prefix", "feature", "Branch prefix (default: feature)")
    rootCmd.AddCommand(addCmd)

    var rmCmd = &cobra.Command{
        Use:   "rm <name>",
        Short: "Remove a worktree (and unregister it)",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            name := args[0]
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

            if err := config.RunOnDelete(leafPath, true, nil); err != nil {
                fmt.Println("on_delete hook error:", err)
                return
            }

            if err := terminal.RunSyncCommand("git", "worktree", "remove", leafPath); err != nil {
                fmt.Println("error removing worktree:", err)
                return
            }
        },
    }
    rootCmd.AddCommand(rmCmd)
    rootCmd.Execute()
}
