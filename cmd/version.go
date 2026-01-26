package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var verbose bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Printf("ipq %s\n", Version)
			fmt.Printf("  commit: %s\n", GitCommit)
			fmt.Printf("  built:  %s\n", BuildDate)
			fmt.Printf("  go:     %s\n", runtime.Version())
			fmt.Printf("  os:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
		} else {
			fmt.Printf("ipq %s\n", Version)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose version info")

	// -V/--version: 直接打印版本并以 0 退出（避免进入主逻辑）
	// CLI Guidelines: 每个 CLI 都应该支持 --version
	rootCmd.PersistentFlags().BoolP("version", "V", false, "Print version")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Printf("ipq %s\n", Version)
			os.Exit(0)
		}
	}
}
