package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mio-cli",
		Short: "CLI di esempio con comandi ordinati",
	}

	// ----- Comando INIT -----
	var initCmd = &cobra.Command{
		Use:   "init <url> <path>",
		Short: "Inizializza un progetto",
		Args:  cobra.ExactArgs(2), // esattamente 2 argomenti
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]
			path := args[1]
			fmt.Printf("Inizializzo progetto con URL=%s in PATH=%s\n", url, path)
		},
	}

	// ----- Comando ADD -----
	var addCmd = &cobra.Command{
		Use:   "add <valore1> [valore2]",
		Short: "Aggiunge un elemento",
		Args:  cobra.RangeArgs(1, 2), // da 1 a 2 argomenti
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
	// Aggiungo i comandi al root
	rootCmd.AddCommand(initCmd, addCmd)

	// Eseguo
	rootCmd.Execute()
}

