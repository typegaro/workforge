package main

import (
	"fmt"
	"os"
	"workforge/config"
	"workforge/terminal"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mio-cli",
		Short: "CLI di esempio con comandi ordinati",
	}

	var initCmd = &cobra.Command{
		Use:   "init <url> <path>",
		Short: "Inizializza un progetto",
		Args:  cobra.RangeArgs(0, 2), 
		Run: func(cmd *cobra.Command, args []string) {
			var path *string
			var err error
			var entries []os.DirEntry
			if len(args)>0 {
				url := args[0]
				if len(args) > 1 {
					path = &args[1]
					entries, err = os.ReadDir(*path)
				}else{
					entries, err = os.ReadDir("./")

				}
				if err != nil {
					fmt.Println("error directory:", err)
					return
				}
				if len(entries) > 0 {
					fmt.Println("this directory is dirty")
					return
				}
				terminal.GitClone(url, path)
			}		
			config.WriteExampleConfig(path)
		},
	}
	var loadCmd= &cobra.Command{
		Use:   "load [dir]",
		Short: "load  a workforge project",
		Args:  cobra.RangeArgs(0, 1), 
		Run: func(cmd *cobra.Command, args []string) {
			var path *string
			var profile *string
			if len(args)>0 {
				path = &args[0]
			}else{
				path = new(string)
				*path = "./"
			}
			config.LoadProject(path,profile)
		},
	}
	rootCmd.AddCommand(initCmd,loadCmd)

	rootCmd.Execute()
}

