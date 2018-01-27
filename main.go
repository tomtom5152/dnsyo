package main

import "github.com/tomtom5152/dnsyo/cmd"

var (
	version    = "devel"
	buildDate  string
	commitHash string
)

func main() {
	cmd.CurrentVersion.Version = version
	cmd.CurrentVersion.BuildDate = buildDate
	cmd.CurrentVersion.CommitHash = commitHash

	cmd.Execute()
}
