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
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tomtom5152/dnsyo/dnsyo"
)

var (
	servers      int
	resolverfile string
	country      string
	requestType  string
	numThreads   int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dnsyo <domain>",
	Short: "Compare the DNS results of 1000+ DNS servers",
	Long:  `Basically dig, if dig queried over 1000 servers and collated their results.`,
	Args:  cobra.MinimumNArgs(1),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// perform a lookup
		q := &dnsyo.Query{
			Domain: args[0],
		}
		err := q.SetType(requestType)
		if err != nil {
			log.Fatal(err.Error())
		}

		sl, err := dnsyo.ServersFromFile(resolverfile)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		if country != "" {
			sl, err = sl.FilterCountry(country)
			if err != nil {
				log.Fatal(err.Error())
			}
		}

		if servers != 0 {
			sl, err = sl.NRandom(servers)
			if err != nil {
				log.Fatal(err.Error())
			}
		}

		q.Results = sl.ExecuteQuery(q, numThreads)

		print(q.ToTextSummary())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	//log.SetFormatter(&log.TextFormatter{})
	//cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test.yaml)")
	rootCmd.PersistentFlags().IntVarP(&numThreads, "threads", "t", 500, "Number of threads to run")
	rootCmd.PersistentFlags().StringVarP(&resolverfile, "resolverfile", "", "dnsyo-resolver-list.yml", "Location of the local yaml resolvers file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().IntVarP(&servers, "servers", "q", 500, "Number of servers to query (0=ALL)")
	rootCmd.Flags().StringVarP(&country, "country", "c", "", "Query servers by two letter country code")
	rootCmd.Flags().StringVarP(&requestType, "type", "", "A", "Type of query to perform")
}
