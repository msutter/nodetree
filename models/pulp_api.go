package models

import (
	"errors"
	"fmt"
	"github.com/msutter/go-pulp/pulp"
	"github.com/spf13/viper"
	"time"
)

func PulpApiClient(n *Node) (client *pulp.Client, err error) {
	// create the API client
	client, err = pulp.NewClient(n.Fqdn, n.ApiUser, n.ApiPasswd, nil)
	if err != nil {
		return client, err
	}
	return
}

func PulpApiGetRepos(n *Node, client *pulp.Client, repositoriy string) (repos []*pulp.Repository, err error) {

	// repository options
	opt := &pulp.GetRepositoryOptions{
		Details: true,
	}

	repos, _, err = client.Repositories.ListRepositories(opt)
	if err != nil {
		return repos, err
	}

	return repos, err

}

func PulpApiSyncRepo(n *Node, client *pulp.Client, repositories []string, progressChannel chan SyncProgress) (err error) {

	waitingTimeout := 10
	waitingRetries := 3

	if n.ApiUser == "" {
		n.ApiUser = viper.GetString("ApiUser")
	}

	if n.ApiPasswd == "" {
		n.ApiPasswd = viper.GetString("ApiPasswd")
	}

	if !n.IsRoot() {
		n.RepositoryError = make(map[string]error)

		// // create the API client
		// client, err := pulp.NewClient(n.Fqdn, n.ApiUser, n.ApiPasswd, nil)
		// if err != nil {
		// 	n.Errors = append(n.Errors, err)
		// 	sp := SyncProgress{
		// 		Node:  n,
		// 		State: "error",
		// 	}
		// 	progressChannel <- sp
		// }

	REPOSITORY_LOOP:
		for _, repository := range repositories {

			callReport, _, err := client.Repositories.SyncRepository(repository)
			if err != nil {
				n.Errors = append(n.Errors, err)
				n.RepositoryError[repository] = err
				sp := SyncProgress{
					Repository: repository,
					Node:       n,
					State:      "error",
				}
				progressChannel <- sp
				continue REPOSITORY_LOOP
			}

			syncTaskId := callReport.SpawnedTasks[0].TaskId
			state := "init"

			progressTries := 0
		PROGRESS_LOOP:
			for (state != "finished") && (state != "error") {
				progressTries++
				if n.AncestorsHaveRepositoryError(repository) {
					// give some between writes on progressChannel
					warningMsg := fmt.Sprintf("skipping sync due to errors on ancestor repository %v on node %v", repository, n.AncestorFqdnsWithRepositoryError(repository)[0])
					sp := SyncProgress{
						Repository: repository,
						Node:       n,
						State:      "skipped",
						Message:    warningMsg,
					}
					progressChannel <- sp
					// break the process loop
					continue REPOSITORY_LOOP
				}

				task, _, err := client.Tasks.GetTask(syncTaskId)
				if err != nil {
					n.RepositoryError[repository] = err
					sp := SyncProgress{
						Repository: repository,
						Node:       n,
						State:      "error",
					}
					progressChannel <- sp
					continue REPOSITORY_LOOP
				}

				if task.State == "error" {
					errorMsg := task.ProgressReport.YumImporter.Metadata.Error
					err = errors.New(errorMsg)
					n.Errors = append(n.Errors, err)
					n.RepositoryError[repository] = err

					sp := SyncProgress{
						Repository: repository,
						Node:       n,
						State:      "error",
					}

					progressChannel <- sp
					continue REPOSITORY_LOOP
				}

				if task.State == "waiting" {
					if progressTries <= waitingRetries {
						time.Sleep(time.Duration(waitingTimeout) * time.Second)
						continue PROGRESS_LOOP

					} else {
						// In case of infinite waiting, kill the task (TODO) and exit with error

						errorMsg := fmt.Sprintf("sync task '%v' has reached timeout in waiting state", task.Id)
						err = errors.New(errorMsg)
						n.Errors = append(n.Errors, err)
						n.RepositoryError[repository] = err

						sp := SyncProgress{
							Repository: repository,
							Node:       n,
							State:      "error",
						}

						progressChannel <- sp
						return err
					}
				}

				state = task.State
				sp := SyncProgress{
					Repository: repository,
					Node:       n,
					State:      state,
				}

				if task.State == "running" {
					if task.ProgressReport.YumImporter.Content != nil {
						sp.SizeTotal = task.ProgressReport.YumImporter.Content.SizeTotal
						sp.SizeLeft = task.ProgressReport.YumImporter.Content.SizeLeft
						sp.ItemsTotal = task.ProgressReport.YumImporter.Content.ItemsTotal
						sp.ItemsLeft = task.ProgressReport.YumImporter.Content.ItemsLeft
					} else {
						if progressTries <= waitingRetries {
							time.Sleep(time.Duration(waitingTimeout) * time.Second)
							continue PROGRESS_LOOP

						} else {
							// In case of infinite waiting, kill the task (TODO) and exit with error

							errorMsg := fmt.Sprintf("sync task '%v' has reached timeout in running state with missing task content object", task.Id)
							err = errors.New(errorMsg)
							n.Errors = append(n.Errors, err)
							n.RepositoryError[repository] = err

							sp := SyncProgress{
								Repository: repository,
								Node:       n,
								State:      "error",
							}

							progressChannel <- sp
							return err
						}

					}
				}
				progressChannel <- sp
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
	return
}
