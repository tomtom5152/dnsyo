package main

import "github.com/tomtom5152/dnsyo/cmd"

var (
	version    = "devel"
	buildDate  string
	commitHash string
)

func main() {
	cmd.CurrentVersion = cmd.VersionInfo{
		Version:    version,
		BuildDate:  buildDate,
		CommitHash: commitHash,
	}
	cmd.Execute()
}
