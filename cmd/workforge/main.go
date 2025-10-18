package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"

	"workforge/internal/app"
	"workforge/internal/config"
	"workforge/internal/terminal"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "wf",
		Short: "Workforge - Forge your work",
	}

	var (
		addNewBranch bool
		addPrefix    string
		gwtFlag      bool
	)

	initCmd := &cobra.Command{
		Use:   "init <url> <path>",
		Short: "Initialize a project",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var (
				path     string
				url      string
				err      error
				entries  []os.DirEntry
				repoName string
			)
			if len(args) > 0 {
				url = args[0]
				repoName = app.RepoURLToName(url)
				entries, err = os.ReadDir("./")
				if err != nil {
					fmt.Println("directory error:", err)
					return
				}
				if len(entries) > 0 {
					if !gwtFlag {
						for _, entry := range entries {
							if entry.Name() == config.ConfigFileName {
								fmt.Println("This is a Workforge directory")
								fmt.Println("You can't clone a new repo here")
								return
							}
						}
					} else {
						fmt.Println("Directory not empty, aborting")
						return
					}
				}
				terminal.GitClone(url, &path)
				if gwtFlag {
					configFilePath := "./" + repoName + "/" + config.ConfigFileName
					_, err = os.Stat(configFilePath)
					if err == nil {
						fmt.Println("Coping wf config from the cloned repo")
						app.CopyFile(configFilePath, config.ConfigFileName)
					} else {
						config.WriteExampleConfig(&path)
					}
				} else {
					config.WriteExampleConfig(&path)
				}
				config.AddWorkforgePrj(repoName, gwtFlag)
			} else {
				fmt.Println("Initializing a new Workforge project")
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Println("error getting current directory:", err)
					return
				}
				repoName = filepath.Base(cwd)
				config.WriteExampleConfig(nil)
				config.AddWorkforgePrj(repoName, gwtFlag)
			}
		},
	}

	initCmd.Flags().BoolVarP(&gwtFlag, "gwt", "t", false, "Use Git worktree")

	var loadProfile string
	loadCmd := &cobra.Command{
		Use:   "load [dir]",
		Short: "Load a Workforge project",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			path := "../"
			if len(args) > 0 {
				path += args[0]
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

	openCmd := &cobra.Command{
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

	addCmd := &cobra.Command{
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
				if err := terminal.AddNewWorktree(name, addPrefix, base, true); err != nil {
					fmt.Println("error creating new worktree:", err)
					return
				}
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
			if err := terminal.AddWorktree(name); err != nil {
				fmt.Println("error adding worktree:", err)
				return
			}
		},
	}

	addCmd.Flags().BoolVarP(&addNewBranch, "b", "b", false, "Create a new branch and worktree (optional base-branch, default: main)")
	addCmd.Flags().StringVar(&addPrefix, "prefix", "feature", "Branch prefix (default: feature)")
	rootCmd.AddCommand(addCmd)

	rmCmd := &cobra.Command{
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
