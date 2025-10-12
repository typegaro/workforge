
package main

import (
	"os"
	"fmt"
	"sort"
	"errors"
	"strings"
	"workforge/config"
	"workforge/terminal"

	"github.com/spf13/cobra"
	"github.com/ktr0731/go-fuzzyfinder"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mio-cli",
		Short: "CLI di esempio con comandi ordinati",
	}
	var gwtFlag bool
	var initCmd = &cobra.Command{
		Use:   "init <url> <path>",
		Short: "Inizializza un progetto",
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
					fmt.Println("error directory:", err)
					return
				}
				if len(entries) > 0 {
					if !gwtFlag {
						for _, entry := range entries {
							if entry.Name() == config.ConfigFileName{
								fmt.Println("This is a workforge dir")
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
				fmt.Println("project path:", *path)
				config.WriteExampleConfig(path)
			}
			config.AddWorkforgePrj(repo_name, path, gwtFlag)
			fmt.Println("Project added to the forge")
		},
	}

	initCmd.Flags().BoolVarP(&gwtFlag, "gwt", "t", false, "Use git work tree")

	var loadCmd = &cobra.Command{
		Use:   "load [dir]",
		Short: "Load a workforge project",
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
		IsGWT bool // true se è una subdir derivata da GitWorkTree
	}

	var openCmd = &cobra.Command{
		Use:   "open",
		Short: "fuzzy finder to open a workforge project",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			projs, hitmap, err := config.ListProjectsExpanded()
			if err != nil {
				fmt.Println("errore caricamento progetti:", err)
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
						return "[GWT] " + items[i].Name
					}
					return "[Repo] " + items[i].Name
				},
				fuzzyfinder.WithPromptString(" Scegli progetto > "),
			)
			if err != nil {
				if errors.Is(err, fuzzyfinder.ErrAbort) {
					return
				}
				fmt.Println("errore fuzzy:", err)
				return
			}

			chosen := items[idx]
			config.LoadProject(chosen.Path,chosen.IsGWT, nil)
		},
	}

	rootCmd.AddCommand(initCmd, loadCmd, openCmd)
	rootCmd.Execute()
}

