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
	"os"
)

var pRepositories []string
var pAllRepositories bool

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

		if len(pRepositories) == 0 {
			ErrorExitWithUsage(cmd, "sync needs a repository name")
		}

		currentStage := stageTree.GetStageByName(args[0])

		// check for flags
		if len(pFqdns) == 0 && len(pTags) == 0 && !pAllNode {
			fmt.Printf("\nWARNING: This will sync the complete tree for the '%v' stage!\n", args[0])
			currentStage.Show()
			fmt.Println("")
			fmt.Printf("you can get rid of this warning by setting the --all flag\n")
			fmt.Printf("Are you sure you want to continue? (yes/no)\n")
			userConfirm := askForConfirmation()

			if !userConfirm {
				ErrorExit("sync canceled !")
			} else {
				pAllNode = true
			}
		}

		var stage *models.Stage

		if pAllNode {
			stage = currentStage
		} else {
			stage = currentStage.Filter(pFqdns, pTags)
		}

		// Create a progress channel
		progressChannel := make(chan models.SyncProgress)

		var renderWg sync.WaitGroup
		renderWg.Add(1)

		switch {
		case pSilent:
			go RenderSilentView(progressChannel, &renderWg)
		case pQuiet:
			go RenderQuietView(progressChannel, &renderWg)
		default:
			go RenderQuietView(progressChannel, &renderWg)
			// go RenderProgressView(stage, progressChannel, &renderWg)
		}

		if pAllRepositories {
			stage.SyncAll(progressChannel)
		} else {
			stage.Sync(pRepositories, progressChannel)
		}

		renderWg.Wait()

		switch {
		case pSilent:
			// no report
		default:
			RenderErrorSummary(stage)
		}

		if stage.HasError() {
			os.Exit(1)
		}
	},
}

func init() {
	pulpCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringSliceVarP(&pRepositories, "repositories", "r", []string{}, "the repositories to be synced.")
	syncCmd.Flags().BoolVar(&pAllRepositories, "all-repositories", false, "sync all repositories")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	//syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().StringSlice("fqdns", []string{}, "Filter on Fqdns")
}

// simple view. No in place updates
func RenderQuietView(progressChannel chan models.SyncProgress, wg *sync.WaitGroup) {
	depthChar := "--- "
	defer wg.Done()
	syncStates := make(map[string]map[string]string)
	for sp := range progressChannel {
		if _, exists := syncStates[sp.Node.Fqdn]; !exists {
			syncStates[sp.Node.Fqdn] = make(map[string]string)
		}
		switch sp.State {
		case "skipped":
			for i := 0; i < sp.Node.Depth; i++ {
				fmt.Printf(depthChar)
			}
			line := fmt.Sprintf("%v %v %v", sp.Node.Fqdn, sp.Repository, sp.State)
			tm.Printf(tm.Color(tm.Bold(line), tm.MAGENTA))
			tm.Flush()
		case "error":
			for i := 0; i < sp.Node.Depth; i++ {
				fmt.Printf(depthChar)
			}
			line := fmt.Sprintf("%v %v %v", sp.Node.Fqdn, sp.Repository, sp.State)
			tm.Printf(tm.Color(tm.Bold(line), tm.RED))
			tm.Flush()
		case "running":
			// only output state changes
			if syncStates[sp.Node.Fqdn][sp.Repository] != sp.State {
				for i := 0; i < sp.Node.Depth; i++ {
					fmt.Printf(depthChar)
				}
				line := fmt.Sprintf("%v %v %v", sp.Node.Fqdn, sp.Repository, sp.State)
				tm.Printf(tm.Color(line, tm.BLUE))
				tm.Flush()
			}
			syncStates[sp.Node.Fqdn][sp.Repository] = sp.State
		case "finished":
			for i := 0; i < sp.Node.Depth; i++ {
				fmt.Printf(depthChar)
			}
			line := fmt.Sprintf("%v %v %v", sp.Node.Fqdn, sp.Repository, sp.State)
			tm.Printf(tm.Color(tm.Bold(line), tm.GREEN))
			tm.Flush()
		}
	}
}

// silent view
func RenderSilentView(progressChannel chan models.SyncProgress, wg *sync.WaitGroup) {
	defer wg.Done()

	for sp := range progressChannel {
		// do nothing
		_ = sp
	}
}

func RenderErrorSummary(s *models.Stage) {
	titleLine := fmt.Sprintf("Found following errors:")
	fmt.Printf("\n")
	fmt.Printf(tm.Bold(titleLine))
	fmt.Printf("\n")
	fmt.Printf("\n")
	_ = "breakpoint"
	for _, n := range s.Nodes {
		if n.HasError() {
			fmt.Printf(tm.Color(tm.Bold(n.Fqdn), tm.RED))
			fmt.Printf("\n")
			for k, v := range n.RepositoryError {
				reposiroryErrorString := fmt.Sprintf("%v: ", k)
				fmt.Printf(" - ")
				fmt.Printf(tm.Color(reposiroryErrorString, tm.RED))
				fmt.Printf(v.Error())
				fmt.Printf("\n")
			}
			fmt.Printf("\n")
		}
	}
}
