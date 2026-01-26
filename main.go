package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"workforge/config"
	"workforge/terminal"

	"github.com/spf13/cobra"
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
					terminal.Error("directory error: %v", err)
					return
				}
				if len(entries) > 0 {
					if !gwtFlag {
						for _, entry := range entries {
							if entry.Name() == config.ConfigFileName {
								terminal.Warn("This is a Workforge directory")
								terminal.Warn("You can't clone a new repo here")
								return
							}
						}
					} else {
						terminal.Warn("Directory not empty, aborting")
						return
					}
				}
				terminal.GitClone(url, &path)
				if gwtFlag {
					config_file_path := "./" + repo_name + "/" + config.ConfigFileName
					_, err = os.Stat(config_file_path)
					if err == nil {
						terminal.Info("Copying Workforge config from the cloned repo")
						CopyFile(config_file_path, config.ConfigFileName)
					} else {
						config.WriteExampleConfig(&path)
					}
				} else {
					config.WriteExampleConfig(&path)
				}
				config.AddWorkforgePrj(repo_name, gwtFlag)
			} else {
				terminal.Info("Initializing a new Workforge project")
				cwd, err := os.Getwd()
				if err != nil {
					terminal.Error("error getting current directory: %v", err)
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
				if err := config.LoadProject(path, false, &loadProfile); err != nil {
					terminal.Error("error loading project: %v", err)
				}
				return
			}
			if err := config.LoadProject(path, false, nil); err != nil {
				terminal.Error("error loading project: %v", err)
			}
		},
	}
	loadCmd.Flags().StringVarP(&loadProfile, "profile", "p", "", "Profile name to use")
	var listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Workforge projects",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			entries, err := config.SortedProjectEntries()
			if err != nil {
				terminal.Error("error loading projects: %v", err)
				return
			}
			for _, entry := range entries {
				fmt.Println(entry.Name)
			}
		},
	}

	var openProfile string
	var openCmd = &cobra.Command{
		Use:   "open <project-name>",
		Short: "Open a Workforge project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			entry, err := config.FindProjectEntry(name)
			if err != nil {
				terminal.Error("%v", err)
				return
			}
			var profile *string
			if openProfile != "" {
				profile = &openProfile
			}
			if err := config.LoadProject(entry.Path, entry.IsGWT, profile); err != nil {
				terminal.Error("error loading project: %v", err)
			}
		},
	}
	openCmd.Flags().StringVarP(&openProfile, "profile", "p", "", "Profile name to use")

	rootCmd.AddCommand(initCmd, loadCmd, listCmd, openCmd)

	var addCmd = &cobra.Command{
		Use:   "add <name> [base-branch]",
		Short: "Add a worktree or create a new branch",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if addNewBranch {
				base := "main"
				if addPrefix == "" {
					addPrefix = "feature"
				}
				if len(args) > 1 {
					base = args[1]
				}
				if err := terminal.AddNewWorkTree(name, addPrefix, base); err != nil {
					terminal.Error("error creating new worktree: %v", err)
					return
				}
				_, err := os.Getwd()
				if err != nil {
					terminal.Error("error getting current directory: %v", err)
					return
				}
			}
			// Use an existing branch as worktree
			if err := terminal.AddWorkTree(name); err != nil {
				terminal.Error("error adding worktree: %v", err)
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
				terminal.Error("error getting current directory: %v", err)
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
				terminal.Warn("could not locate worktree directory for: %s", name)
				return
			}

			if err := config.RunOnDelete(leafPath, true, nil); err != nil {
				terminal.Error("on_delete hook error: %v", err)
				return
			}

			if err := terminal.RunSyncCommand("git", "worktree", "remove", leafPath); err != nil {
				terminal.Error("error removing worktree: %v", err)
				return
			}
		},
	}
	rootCmd.AddCommand(rmCmd)
	rootCmd.Execute()
}
