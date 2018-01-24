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

	"github.com/spf13/cobra"
	. "github.com/tomtom5152/dnsyo/dnsyo"
	"github.com/azer/logger"
	"github.com/miekg/dns"
	"strings"
	"time"
)

var (
	servers      int
	resolverfile string
	country      string
	requestType  string
	requestRate  int
)

var yoLog = logger.New("dnsyo")

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dnsyo <domain>",
	Short: "Compare the DNS results of 1000+ DNS servers",
	Long:  `Basically dig, if dig queried over 1000 servers and collated their results.`,
	Args:  cobra.MinimumNArgs(1),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		requestType = strings.ToUpper(requestType)
		t, ok := dns.StringToType[requestType]
		if !ok {
			yoLog.Error("unable to use type %s", requestType)
		}

		// perform a lookup
		sl, err := ServersFromFile(resolverfile)
		if err != nil {
			yoLog.Error(err.Error())
			os.Exit(1)
			return
		}

		if country != "" {
			sl, err = sl.FilterCountry(country)
			if err != nil {
				yoLog.Error(err.Error())
				os.Exit(2)
				return
			}
		}

		if servers != 0 {
			sl, err = sl.NRandom(servers)
			if err != nil {
				yoLog.Error(err.Error())
				os.Exit(2)
				return
			}
		}

		result := sl.Query(args[0], t, time.Second / time.Duration(requestRate))

		fmt.Printf(`
 - RESULTS
I asked %d servers for %s records related to %s,
%d responded with records and %d gave errors
Here are the results;`, servers, requestType, args[0], result.SuccessCount, result.ErrorCount)
		fmt.Print("\n\n\n")

		if result.SuccessCount > 0{
			for result, count := range result.Success {
				fmt.Printf("%d servers responded with;\n%s\n\n", count, result)
			}
		}

		if result.ErrorCount > 0{
			fmt.Print("\nAnd here are the errors;\n\n")

			for err, count := range result.Errors {
				fmt.Printf("%d servers responded with;\n%s\n\n", count, err)
			}
		}
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
	//cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test.yaml)")
	rootCmd.PersistentFlags().IntVarP(&requestRate, "rate", "r", 500, "Number of requests per second")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().IntVarP(&servers, "servers", "q", 500, "Number of servers to query")
	rootCmd.Flags().StringVarP(&resolverfile, "resolverfile", "", "config/resolver-list.yml", "Location of the local yaml resolvers file")
	rootCmd.Flags().StringVarP(&country, "country", "c", "", "Query servers by two letter country code")
	rootCmd.Flags().StringVarP(&requestType, "type", "", "A", "Type of query to perform")
}
