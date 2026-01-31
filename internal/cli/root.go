package cli

import (
	"fmt"
	"path/filepath"

	"workforge/internal/app"
	appgit "workforge/internal/app/git"
	"workforge/internal/app/project"
	"workforge/internal/infra/log"

	"github.com/spf13/cobra"
)

func Execute() {
	orchestrator := app.NewOrchestrator()

	var rootCmd = &cobra.Command{
		Use:   "wf",
		Short: "Workforge - Forge your work",
	}
	var addCreateBranch bool
	var addBaseBranch string
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
			if err := orchestrator.InitProject(url, gwtFlag); err != nil {
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
			path := ".."
			if len(args) > 0 {
				path = filepath.Join(path, args[0])
			}
			if loadProfile != "" {
				if err := orchestrator.LoadProject(path, false, &loadProfile); err != nil {
					log.Error("error loading project: %v", err)
				}
				return
			}
			if err := orchestrator.LoadProject(path, false, nil); err != nil {
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
			entries, err := orchestrator.Projects().SortedProjectEntries()
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
			entry, err := orchestrator.Projects().FindProjectEntry(name)
			if err != nil {
				log.Error("%v", err)
				return
			}
			var profile *string
			if openProfile != "" {
				profile = &openProfile
			}
			if err := orchestrator.LoadProject(entry.Path, entry.IsGWT, profile); err != nil {
				log.Error("error loading project: %v", err)
			}
		},
	}
	openCmd.Flags().StringVarP(&openProfile, "profile", "p", "", "Profile name to use")

	rootCmd.AddCommand(initCmd, loadCmd, listCmd, openCmd)

	var addCmd = &cobra.Command{
		Use:   "add [worktree] <branch>",
		Short: "Add a worktree (optionally create the branch if missing)",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			worktreePath := "."
			branch := args[0]
			if len(args) > 1 {
				entry, err := orchestrator.Projects().FindProjectEntry(args[0])
				if err == nil {
					worktreePath = entry.Path
				} else {
					worktreePath = args[0]
				}
				branch = args[1]
			}
			if err := orchestrator.Git().AddWorktree(worktreePath, branch, addCreateBranch, addBaseBranch); err != nil {
				log.Error("error adding worktree: %v", err)
				return
			}
		},
	}

	addCmd.Flags().BoolVarP(&addCreateBranch, "create-branch", "c", false, "Create the branch if it does not exist")
	addCmd.Flags().StringVar(&addBaseBranch, "base", "main", "Base branch for new branch creation")
	rootCmd.AddCommand(addCmd)

	var rmCmd = &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a worktree (and unregister it)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			leafPath, err := orchestrator.RemoveWorktree(name)
			if err != nil {
				if _, ok := err.(project.WorktreeNotFoundError); ok {
					log.Warn("could not locate worktree directory for: %s", name)
					return
				}
				if _, ok := err.(appgit.OnDeleteError); ok {
					log.Error("on_delete hook error: %v", err)
					return
				}
				if _, ok := err.(appgit.RemoveWorktreeError); ok {
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
	rootCmd.AddCommand(NewPluginCmd())
	rootCmd.AddCommand(NewHealthcheckCmd())
	rootCmd.Execute()
}
