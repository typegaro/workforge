package cli

import (
	"fmt"

	"workforge/internal/app/plugin"

	"github.com/spf13/cobra"
)

func NewHealthcheckCmd() *cobra.Command {
	pluginsDir := plugin.DefaultPluginsDir()
	registry := plugin.NewPluginRegistryService(plugin.DefaultRegistryPath())
	pluginSvc := plugin.NewPluginService(pluginsDir, plugin.DefaultSocketsDir())

	return &cobra.Command{
		Use:     "healthcheck",
		Aliases: []string{"hc"},
		Short:   "Run healthcheck on all plugins that support it",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			payload := map[string]interface{}{"project": projectNameFromCwd()}
			results := pluginSvc.RunHook(registry, "on_healthcheck", payload)
			if len(results) == 0 {
				fmt.Println("No plugins with healthcheck support")
				return
			}
			for name, r := range results {
				if r.Error != nil {
					fmt.Printf("ERR %s: %v\n", name, r.Error)
				} else {
					fmt.Printf("OK  %s: %s\n", name, string(r.Response))
				}
			}
			pluginSvc.KillAll()
		},
	}
}
