// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
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
	"github.com/spf13/viper"
	"nodetree/models"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here

		// Unmarshal nodetree config file
		var stageTree models.StageTree
		viper.Unmarshal(&stageTree)

	},
}

func init() {

	pulpCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// //  this are only test statements. Will be invoked by the future command line
// config := models.NewConfig()
// stage_tree := config.GetStageTree()

// log.Info.Println("")
// log.Info.Println("-------------- START SYNC LAB TREE --------------------")
// lab_stage := stage_tree.GetStageByName("lab")
// lab_stage.Sync()
// log.Info.Println("-------------- START SYNC LAB TREE --------------------")

// log.Info.Println("")
// log.Info.Println("-------------- START SYNC PRD TREE --------------------")
// prd_stage := stage_tree.GetStageByName("prd")
// prd_stage.Sync()
// log.Info.Println("-------------- START SYNC PRD TREE --------------------")

// log.Info.Println("")
// fqdns := []string{"pulp-lab-12411.local", "pulp-lab-11111.local"}
// log.Info.Printf("-------------- START SYNC FQDNS %v --------------------\n", fqdns)
// lab_stage.SyncByFilters(fqdns, []string{})
// log.Info.Printf("-------------- STOP SYNC FQDNS %v --------------------\n", fqdns)

// log.Info.Println("")
// tags := []string{"111MZ", "12MZ"}
// log.Info.Printf("-------------- START SYNC TAGS %v --------------------\n", tags)
// lab_stage.SyncByFilters([]string{}, tags)
// log.Info.Printf("-------------- STOP SYNC TAGS %v --------------------\n", tags)

// log.Info.Println("")
// log.Info.Printf("-------------- START SYNC BY Filters %v %v --------------------\n", fqdns, tags)
// lab_stage.SyncByFilters(fqdns, tags)
// log.Info.Printf("-------------- STOP SYNC BY Filters %v %v --------------------\n", fqdns, tags)
