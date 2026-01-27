package cli

import (
	"fmt"
	"os"

	"workforge/internal/app"
	"workforge/internal/infra/log"

	"github.com/spf13/cobra"
)

func Execute() {
	service := app.NewService()

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
			var url string
			if len(args) > 0 {
				url = args[0]
			}
			if err := service.InitProject(url, gwtFlag); err != nil {
				log.Error("%v", err)
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
				if err := service.LoadProject(path, false, &loadProfile); err != nil {
					log.Error("error loading project: %v", err)
				}
				return
			}
			if err := service.LoadProject(path, false, nil); err != nil {
				log.Error("error loading project: %v", err)
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
			entries, err := service.SortedProjectEntries()
			if err != nil {
				log.Error("error loading projects: %v", err)
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
			entry, err := service.FindProjectEntry(name)
			if err != nil {
				log.Error("%v", err)
				return
			}
			var profile *string
			if openProfile != "" {
				profile = &openProfile
			}
			if err := service.LoadProject(entry.Path, entry.IsGWT, profile); err != nil {
				log.Error("error loading project: %v", err)
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
				if err := service.AddNewWorkTree(name, addPrefix, base); err != nil {
					log.Error("error creating new worktree: %v", err)
					return
				}
				if _, err := os.Getwd(); err != nil {
					log.Error("error getting current directory: %v", err)
					return
				}
			}
			if err := service.AddWorkTree(name); err != nil {
				log.Error("error adding worktree: %v", err)
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
			leafPath, err := service.RemoveWorktree(name)
			if err != nil {
				if _, ok := err.(app.WorktreeNotFoundError); ok {
					log.Warn("could not locate worktree directory for: %s", name)
					return
				}
				if _, ok := err.(app.OnDeleteError); ok {
					log.Error("on_delete hook error: %v", err)
					return
				}
				if _, ok := err.(app.RemoveWorktreeError); ok {
					log.Error("error removing worktree: %v", err)
					return
				}
				log.Error("error removing worktree: %v", err)
				return
			}
			_ = leafPath
		},
	}
	rootCmd.AddCommand(rmCmd)
	rootCmd.Execute()
}
