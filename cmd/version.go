package cmd

import (
	"fmt"
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
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose version info")

	// -V 显示版本
	rootCmd.Flags().BoolP("version", "V", false, "Print version")
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if v, _ := cmd.Flags().GetBool("version"); v {
			versionCmd.Run(cmd, args)
			return fmt.Errorf("")
		}
		return nil
	}
}
