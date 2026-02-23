package cmd

import (
	"fmt"
	"os"

	"github.com/LyVanBong/swarm-ctl/internal/audit"
	"github.com/spf13/cobra"
)

const asciiLogo = `
 ███████╗██╗    ██╗ █████╗ ██████╗ ███╗   ███╗      ██████╗████████╗██╗     
 ██╔════╝██║    ██║██╔══██╗██╔══██╗████╗ ████║     ██╔════╝╚══██╔══╝██║     
 ███████╗██║ █╗ ██║███████║██████╔╝██╔████╔██║     ██║        ██║   ██║     
 ╚════██║██║███╗██║██╔══██║██╔══██╗██║╚██╔╝██║     ██║        ██║   ██║     
 ███████║╚███╔███╔╝██║  ██║██║  ██║██║ ╚═╝ ██║     ╚██████╗   ██║   ███████╗
 ╚══════╝ ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝      ╚═════╝   ╚═╝   ╚══════╝`

var (
	cfgFile string
	verbose bool
	noColor bool

	rootCmd = &cobra.Command{
		Use:   "swarm-ctl",
		Short: "🐳 Enterprise Docker Swarm Manager",
		Long: asciiLogo + `

Enterprise-grade Docker Swarm cluster management tool.
Manage nodes, services, storage, and monitoring from a single CLI.

Documentation: https://github.com/LyVanBong/swarm-ctl
`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Ghi log hành động kiểm toán ngoại trừ khi user chạy `help`
			if len(os.Args) > 1 {
				audit.Log(os.Args[1:])
			}
		},
	}
)

// Execute là entry point chính
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.swarm-ctl/config.yml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	rootCmd.AddCommand(clusterCmd)
	rootCmd.AddCommand(nodeCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(secretCmd)
	rootCmd.AddCommand(storageCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(appCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(versionCmd)
}
