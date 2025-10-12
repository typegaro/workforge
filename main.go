
package main

import (
	"fmt"
	"os"
	"strings"
	"workforge/config"
	"workforge/terminal"

	"github.com/spf13/cobra"
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
			config.LoadProject(path, profile)
		},
	}

	rootCmd.AddCommand(initCmd, loadCmd)
	rootCmd.Execute()
}

