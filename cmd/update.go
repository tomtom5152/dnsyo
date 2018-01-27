// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tomtom5152/dnsyo/dnsyo"
)

var (
	csvURL string
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the list of resolvers",
	Long: `Performs a test query on all of the configured name servers to see if they are working and saves the output
to the list of active servers.`,
	Run: func(cmd *cobra.Command, args []string) {
		toTest, err := dnsyo.ServersFromCSVURL(csvURL)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("Testing %d nameservers\n", len(toTest))
		working := toTest.TestAll(numThreads)
		err = working.DumpToFile(resolverfile)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		log.Info("Updated server list, %d active, %d disabled\n", len(working), len(toTest)-len(working))

		return
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	updateCmd.Flags().StringVar(&csvURL, "csvurl", "https://public-dns.info/nameservers.csv", "URL to fetch the list form")
}
