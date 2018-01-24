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
	"os"
	"fmt"
	"time"
)

func Update(source, target string) error {
	toTest, err := ServersFromFile(source)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}

	fmt.Printf("Testing %d nameservers\n", len(toTest))

	var wg sync.WaitGroup
	var mutex sync.Mutex
	var working ServerList
	limiter := time.Tick(time.Second / time.Duration(requestRate))
	for _, s := range toTest {
		wg.Add(1)
		go func(s Server) {
			defer wg.Done()
			<-limiter
			log.Debug("Testing " + s.Reverse)
			if ok, err := s.Test(); ok {
				mutex.Lock()
				working = append(working, s)
				mutex.Unlock()
			} else {
				log.WithFields(log.Fields{
					"server": s.Reverse,
					"reason": err,
				}).Info("Disabling server")
			}
		}(s)
	}

	wg.Wait()
	err = working.DumpToFile(target)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	fmt.Printf("Updated server list, %d active, %d disabled\n", len(working), len(toTest)-len(working))

	return nil
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update <source> <destination>",
	Short: "Update the list of resolvers",
	Long: `Performs a test query on all of the configured name servers to see if they are working and saves the output
to the list of active servers.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := Update(args[0], args[1]); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
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
	updateCmd.Flags().StringP("csvurl", "u", "https://public-dns.info/nameservers.csv", "URL to fetch the list form")
}
