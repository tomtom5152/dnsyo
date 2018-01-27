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
	"github.com/spf13/cobra"
	"github.com/tomtom5152/dnsyo/dnsyo"
	log "github.com/sirupsen/logrus"
	"github.com/tomtom5152/dnsyo/api"
	"os"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the basic API Server",
	Run: func(cmd *cobra.Command, args []string) {
		var working dnsyo.ServerList
		if yml := cmd.Flag("resolverfile").Value.String(); yml != "" {
			var err error
			working, err = dnsyo.ServersFromFile(yml)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			toTest, err := dnsyo.ServersFromCSVURL(cmd.Flag("csvurl").Value.String())
			if err != nil {
				log.Fatal(err)
			}

			working = toTest.TestAll(numThreads)
		}

		// start the server
		server := api.NewAPIServer(working)

		port := os.Getenv("PORT")
		if port == "" {
			port = cmd.Flag("port").Value.String()
		}
		log.Infof("Starting API server with %d nameservers on port %s", len(working), port)
		server.Run(port)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serveCmd.Flags().StringP("port", "p", ":3000", "Port to bind to")
	serveCmd.Flags().StringVar(&csvURL, "csvurl", "https://public-dns.info/nameservers.csv", "URL to fetch the list form")
	serveCmd.Flags().String("resolverfile", "", "Local resolvers file to use")
}
