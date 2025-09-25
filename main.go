package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"workforge/terminal"
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
			if len(args)>0 {
				url := args[0]
				if len(args) > 1 {
					*path = args[1]
				}
				terminal.GitClone(url, path)
			}
		},
	}

	var addCmd = &cobra.Command{
		Use:   "add <valore1> [valore2]",
		Short: "Aggiunge un elemento",
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

