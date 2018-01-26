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
	. "github.com/tomtom5152/dnsyo/dnsyo"
	log "github.com/sirupsen/logrus"
	"sync"
	"github.com/tomtom5152/dnsyo/api"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the basic API Server",
	Run: func(cmd *cobra.Command, args []string) {
		var working ServerList
		if yml := cmd.Flag("resolverfile").Value.String(); yml != "" {
			var err error
			working, err = ServersFromFile(yml)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			toTest, err := ServersFromCSVURL(cmd.Flag("csvurl").Value.String())
			if err != nil {
				log.Fatal(err)
			}

			var wg sync.WaitGroup
			var mutex sync.Mutex
			testQueue := make(chan Server, len(toTest))

			// start workers
			for i := 0; i < numThreads; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					for s := range testQueue {
						log.WithField("thread", i).Debug("Testing " + s.Name)
						if ok, err := s.Test(); ok {
							mutex.Lock()
							working = append(working, s)
							mutex.Unlock()
						} else {
							sName := s.Name
							if sName == "" {
								sName = s.Ip
							}
							log.WithFields(log.Fields{
								"thread": i,
								"server": sName,
								"reason": err,
							}).Info("Disabling server")
						}
					}
				}(i)
			}

			// add test servers
			for _, s := range toTest {
				testQueue <- s
			}
			close(testQueue)

			wg.Wait()
		}

		// start the server
		server := api.NewAPIServer(working)

		log.Infof("Starting API server with %d nameservers", len(working))
		server.Run(cmd.Flag("port").Value.String())
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
	serveCmd.Flags().StringVar(&csvUrl, "csvurl", "https://public-dns.info/nameservers.csv", "URL to fetch the list form")
	serveCmd.Flags().String("resolverfile", "", "Local resolvers file to use")
}
