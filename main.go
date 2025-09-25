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
					*path = args[1]
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

	var addCmd = &cobra.Command{
		Use:   "add <valore1> [valore2]",
		Short: "add a new branch to the worktree",
		Args:  cobra.RangeArgs(1, 2), 
		Run: func(cmd *cobra.Command, args []string) {
			val1 := args[0]
			var val2 string
			if len(args) > 1 {
				val2 = args[1]
			}
			fmt.Printf("Aggiungo: %s", val1)
			if val2 != "" {
				fmt.Printf(" con extra: %s", val2)
			}
			fmt.Println()
		},
	}
	rootCmd.AddCommand(initCmd, addCmd)

	rootCmd.Execute()
}

