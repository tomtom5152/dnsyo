package cmd

import (
	"github.com/spf13/cobra"
	"runtime"
)

type VersionInfo struct {
	Version    string
	BuildDate  string
	CommitHash string
}

var CurrentVersion VersionInfo

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the dnsyo version information",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf(`dnsyo:
	version		: %s
	build date	: %s
	git hash	: %s
	go version	: %s
	go compiler	: %s
	platform	: %s/%s
`,
			CurrentVersion.Version, CurrentVersion.BuildDate, CurrentVersion.CommitHash,
			runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH,
		)
		return
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
