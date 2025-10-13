package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"workforge/config"
	"workforge/terminal"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "wf",
		Short: "Workforge - Forge your work",
	}

	var gwtFlag bool
	initCmd := &cobra.Command{
		Use:   "init <url> [path]",
		Short: "Initialize a project",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]
			repoName := strings.TrimSuffix(filepath.Base(url), ".git")
			var dest *string
			if len(args) > 1 {
				dest = &args[1]
			}

			if err := ensureInitTarget(dest, gwtFlag); err != nil {
				fmt.Println("directory error:", err)
				return
			}

			if err := terminal.GitClone(url, dest); err != nil {
				fmt.Println("git clone failed:", err)
				return
			}

			clonePath := dest
			if !gwtFlag {
				if clonePath == nil {
					clonePath = &repoName
				} else {
					full := filepath.Join(*clonePath, repoName)
					clonePath = &full
				}
				fmt.Println("Project path:", *clonePath)
			}

			if err := config.WriteExampleConfig(clonePath); err != nil {
				fmt.Println("error writing example config:", err)
			}
			if err := config.AddWorkforgePrj(repoName, clonePath, gwtFlag); err != nil {
				fmt.Println("error registering project:", err)
				return
			}
			fmt.Println("Project added to Workforge")
		},
	}
	initCmd.Flags().BoolVarP(&gwtFlag, "gwt", "t", false, "Use Git worktree")

	loadCmd := &cobra.Command{
		Use:   "load [dir]",
		Short: "Load a Workforge project",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			if err := config.LoadProject(path, nil); err != nil {
				fmt.Println("error loading project:", err)
			}
		},
	}

	type projItem struct {
		config.Project
		isGWT bool
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
				items = append(items, projItem{Project: p, isGWT: hitmap[name]})
			}
			sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })

			idx, err := fuzzyfinder.Find(
				items,
				func(i int) string {
					if items[i].isGWT {
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
			if err := config.LoadProject(chosen.Path, nil); err != nil {
				fmt.Println("error loading project:", err)
			}
		},
	}

	rootCmd.AddCommand(initCmd, loadCmd, openCmd)

	var addNewBranch bool
	var addPrefix string
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
				if err := terminal.AddNewWorkTree(name, addPrefix, base, true); err != nil {
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
			if err := terminal.AddWorkTree(name); err != nil {
				fmt.Println("error adding worktree:", err)
				return
			}
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Println("error getting current directory:", err)
				return
			}
			folderName := strings.ReplaceAll(name, "/", "-")
			leafAbs := filepath.Join(cwd, "..", folderName)
			if err := config.AddWorkforgeLeaf(leafAbs); err != nil {
				fmt.Println("error registering worktree:", err)
			}
		},
	}
	addCmd.Flags().BoolVarP(&addNewBranch, "b", "b", false, "Create a new branch and worktree (optional base-branch, default: main)")
	addCmd.Flags().StringVar(&addPrefix, "prefix", "feature", "Branch prefix (default: feature)")
	rootCmd.AddCommand(addCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func ensureInitTarget(path *string, gwt bool) error {
	target := "./"
	if path != nil && *path != "" {
		target = *path
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	if gwt {
		return fmt.Errorf("directory not empty")
	}
	for _, entry := range entries {
		if entry.Name() == config.ConfigFileName {
			return fmt.Errorf("this is already a Workforge directory")
		}
	}
	return nil
}
