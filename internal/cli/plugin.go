package cli

import (
	"fmt"
	"strings"

	"workforge/internal/app/plugin"
	"workforge/internal/infra/log"

	"github.com/spf13/cobra"
)

func NewPluginCmd() *cobra.Command {
	pluginsDir := plugin.DefaultPluginsDir()
	registry := plugin.NewPluginRegistryService(plugin.DefaultRegistryPath())
	installer := plugin.NewPluginInstallerService(pluginsDir, registry)
	pluginSvc := plugin.NewPluginService(pluginsDir, plugin.DefaultSocketsDir())

	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
	}

	addCmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Install a plugin from a git repository",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]
			entry, err := installer.Install(url)
			if err != nil {
				log.Error("install plugin: %v", err)
				return
			}
			fmt.Printf("Installed plugin: %s\n", entry.Name)
			fmt.Printf("  Config key: %s\n", entry.ConfigKey)
			fmt.Printf("  Hooks: %s\n", strings.Join(entry.Hooks, ", "))
		},
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed plugins",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			plugins, err := registry.List()
			if err != nil {
				log.Error("list plugins: %v", err)
				return
			}
			if len(plugins) == 0 {
				fmt.Println("No plugins installed")
				return
			}
			for _, p := range plugins {
				fmt.Printf("%s\n", p.Name)
				fmt.Printf("  config_key: %s\n", p.ConfigKey)
				fmt.Printf("  hooks: %s\n", strings.Join(p.Hooks, ", "))
				fmt.Printf("  url: %s\n", p.URL)
			}
		},
	}

	rmCmd := &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove an installed plugin",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if err := installer.Uninstall(name); err != nil {
				log.Error("remove plugin: %v", err)
				return
			}
			fmt.Printf("Removed plugin: %s\n", name)
		},
	}

	registerCmd := &cobra.Command{
		Use:   "register <name>",
		Short: "Register an existing plugin directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if err := installer.Register(name); err != nil {
				log.Error("register plugin: %v", err)
				return
			}
			fmt.Printf("Registered plugin: %s\n", name)
		},
	}

	healthcheckCmd := &cobra.Command{
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

	runCmd := &cobra.Command{
		Use:   "run <name>",
		Short: "Wake up a plugin asynchronously and run on_plugin_wakeup hook",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			resultCh := pluginSvc.WakeupAsync(name)
			fmt.Printf("Starting plugin %s...\n", name)

			result := <-resultCh
			if result.Error != nil {
				log.Error("wakeup plugin: %v", result.Error)
				return
			}
			log.Info("Plugin %s is running", name)

			payload := map[string]interface{}{"project": projectNameFromCwd()}
			resp, err := pluginSvc.Call(name, "on_plugin_wakeup", payload)
			if err != nil {
				log.Error("call on_plugin_wakeup: %v", err)
				return
			}
			if len(resp) > 0 {
				fmt.Printf("%s\n", string(resp))
			}
		},
	}

	killCmd := &cobra.Command{
		Use:   "kill <name>",
		Short: "Kill a running plugin",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if !pluginSvc.IsRunning(name) {
				fmt.Printf("Plugin %s is not running\n", name)
				return
			}
			if err := pluginSvc.Kill(name); err != nil {
				log.Error("kill plugin: %v", err)
				return
			}
			fmt.Printf("Killed plugin: %s\n", name)
		},
	}

	pluginCmd.AddCommand(addCmd, listCmd, rmCmd, registerCmd, healthcheckCmd, runCmd, killCmd)
	return pluginCmd
}
