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
	"github.com/msutter/nodetree/models"
	"github.com/spf13/cobra"
	"sync"
	// "time"

	tm "github.com/buger/goterm"
)

var pRepository string
var pQuiet bool
var pSilent bool

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync [stage name]",
	Short: "Synchronization of pulp nodes for a given stage",
	Long: `Synchronization of pulp nodes in a given stage

Filters can be set on Fqdns and tags.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			ErrorExitWithUsage(cmd, "sync needs a name for the stage")
		}

		if pRepository == "" {
			ErrorExitWithUsage(cmd, "sync needs a repository name")
		}

		currentStage := stageTree.GetStageByName(args[0])

		// check for flags
		if len(pFqdns) == 0 && len(pTags) == 0 && !pAll {
			fmt.Printf("\nWARNING: This will sync the complete tree for the '%v' stage!\n", args[0])
			fmt.Println("")
			currentStage.Show()
			fmt.Println("")
			fmt.Printf("you can get rid of this warning by setting the --all flag\n")
			fmt.Printf("Are you sure you want to continue? (yes/no)\n")
			userConfirm := askForConfirmation()
			if !userConfirm {
				ErrorExit("sync canceled !")
			} else {
				pAll = true
			}
		}

		// Create a progress channel
		progressChannel := make(chan models.SyncProgress)
		// create a state map
		nodeStates := make(map[string]string)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sp := range progressChannel {
				switch sp.State {
				case "skipped":
					line := fmt.Sprintf("%v %v", sp.Node.GetTreeRaw(sp.Node.Fqdn), sp.State)
					tm.Printf(tm.Color(tm.Bold(line), tm.MAGENTA))
					tm.Flush()
				case "error":
					line := fmt.Sprintf("%v %v", sp.Node.GetTreeRaw(sp.Node.Fqdn), sp.State)
					tm.Printf(tm.Color(tm.Bold(line), tm.RED))
					tm.Flush()
				case "running":
					// only output state changes
					if nodeStates[sp.Node.Fqdn] != sp.State {
						line := fmt.Sprintf("%v %v", sp.Node.GetTreeRaw(sp.Node.Fqdn), sp.State)
						tm.Printf(tm.Color(line, tm.BLUE))
						tm.Flush()
					}
					nodeStates[sp.Node.Fqdn] = sp.State
				case "finished":
					line := fmt.Sprintf("%v %v", sp.Node.GetTreeRaw(sp.Node.Fqdn), sp.State)
					tm.Printf(tm.Color(tm.Bold(line), tm.GREEN))
					tm.Flush()
				}
			}
		}()

		var err models.SyncErrors

		if pAll {
			err = currentStage.Sync(pRepository, progressChannel)
		} else {
			filteredStage := currentStage.Filter(pFqdns, pTags)
			err = filteredStage.Sync(pRepository, progressChannel)
		}

		wg.Wait()
		if err.Any() {
			fmt.Printf("\n")
			fmt.Printf("Error Summary:\n")
			fmt.Printf("\n")
			fmt.Println(err.Error())
		}

	},
}

func init() {
	pulpCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVarP(&pRepository, "repository", "r", "", "the repository to be synced.")
	syncCmd.Flags().BoolVarP(&pQuiet, "quiet", "q", false, "simple output")
	syncCmd.Flags().BoolVarP(&pSilent, "silent", "s", false, "no output")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	//syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().StringSlice("fqdns", []string{}, "Filter on Fqdns")
}
