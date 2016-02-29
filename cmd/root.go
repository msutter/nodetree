// Copyright © 2016 Marc Sutter <marc.sutter@swissflow.ch>
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
	// "github.com/msutter/nodetree/log"
	"github.com/msutter/nodetree/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

// models
var stageTree models.StageTree

// flags
var pFqdns []string
var pTags []string
var pAllNode bool
var pQuiet bool
var pSilent bool

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "nodetree",
	Short: "A node tree manager",
	Long: `A node tree manager

nodetree is a CLI that can manages nodes in a tree through API calls`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nodetree.yaml)")
	RootCmd.PersistentFlags().StringSliceVarP(&pFqdns, "fqdn", "f", []string{}, "Filter on Fqdn. You can define multiple fqdns by repeating the -f flag for each fqdn")
	RootCmd.PersistentFlags().StringSliceVarP(&pTags, "tag", "t", []string{}, "Filter on Tag. You can define multiple tags by repeating the -t flag for each tag")
	RootCmd.PersistentFlags().BoolVarP(&pAllNode, "all", "a", false, "Execute the command on all nodes in this stage tree")
	RootCmd.PersistentFlags().BoolVarP(&pQuiet, "quiet", "q", false, "simple output")
	RootCmd.PersistentFlags().BoolVarP(&pSilent, "silent", "s", false, "no output")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".nodetree") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	viper.AutomaticEnv()             // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); (err == nil) && !pSilent {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	viper.Unmarshal(&stageTree)

}

// askForConfirmation uses Scanln to parse user input. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user. Typically, you should use fmt to print out a question
// before calling askForConfirmation. E.g. fmt.Println("WARNING: Are you sure? (yes/no)")
func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		panic(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

func ErrorExitWithUsage(ctx *cobra.Command, message string) {
	fmt.Printf(message)
	ctx.Usage()
	os.Exit(1)
}

func ErrorExit(message string) {
	fmt.Printf(message)
	os.Exit(1)
}
