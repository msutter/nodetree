package models

import (
	"errors"
	"fmt"
	"github.com/msutter/go-pulp/pulp"
	"github.com/spf13/viper"
	"time"
)

// check if node contains repo
func NodeContainsRepo(pulpRepos []*pulp.Repository, repository string) bool {
	for _, r := range pulpRepos {
		if r.Id == repository {
			return true
		}
	}
	return false
}

// func FeedHostMatchParent(n *Node, pulpRepos []*pulp.Repository, repository string) (bool, err error) {
// 	for _, r := range pulpRepos {
// 		feed := r.Importers[0].ImporterConfig.Feed
// 		u, err := url.Parse(feed)

// 		if u.Host != n.Parent.Fqdn {
// 			errorMsg := fmt.Sprintf("repository '%v' on node %v has invalid feed '%v'", repository, n.Fqdn, feed)
// 			err = errors.New(errorMsg)
// 			return false, err
// 		}

// 		pathSlice := path.Split(u.Path)
// 		repoInPath := pathSlice[len(pathSlice)-2]

// 		if
// 			errorMsg := fmt.Sprintf("repository '%v' on node %v has invalid feed '%v'", repository, n.Fqdn, feed)
// 			err = errors.New(errorMsg)
// 			return false, err
// 		}

// 	}
// 	return true, err
// }

// func ValidateRepoSync(n *Node, pulpRepos []*pulp.Repository, repository string) (bool, err error) {
// 	for _, r := range pulpRepos {
// 		feed := r.Importers[0].ImporterConfig.Feed
// 		u, err := url.Parse(feed)

// 		if r.Id == repository {
// 			errorMsg := fmt.Sprintf("repository '%v' does not exist on node %v", repository, n.Fqdn)
// 			err = errors.New(errorMsg)
// 			return false, err
// 		}

// 		repoInPath = path.Split(u.Path)

// 		if u.Host != n.Parent.Fqdn {
// 			errorMsg := fmt.Sprintf("repository '%v' does not exist on node %v", repository, n.Fqdn)
// 			err = errors.New(errorMsg)
// 			return false, err
// 		}

// 	}
// 	return false, err
// }

func PulpApiClient(n *Node) (client *pulp.Client, err error) {

	// Use default credentials if not specified on node level
	if n.ApiUser == "" {
		n.ApiUser = viper.GetString("ApiUser")
	}
	if n.ApiPasswd == "" {
		n.ApiPasswd = viper.GetString("ApiPasswd")
	}

	// create the API client
	client, err = pulp.NewClient(n.Fqdn, n.ApiUser, n.ApiPasswd, nil)
	if err != nil {
		return client, err
	}
	return
}

// Return a list of all repositories
func PulpApiGetRepos(n *Node, client *pulp.Client) (repos []*pulp.Repository, err error) {

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

	// Get the repos on the target node
	var remoteRepos []*pulp.Repository
	remoteRepos, err = PulpApiGetRepos(n, client)

	if !n.IsRoot() {

	REPOSITORY_LOOP:
		for _, repository := range repositories {

			repoExists := NodeContainsRepo(remoteRepos, repository)
			_ = "breakpoint"

			// check if repo exists on target node
			if !repoExists {
				errorMsg := fmt.Sprintf("repository '%v' does not exist on node %v", repository, n.Fqdn)
				err = errors.New(errorMsg)
				// n.Errors = append(n.Errors, err)
				n.RepositoryError[repository] = err
				sp := SyncProgress{
					Repository: repository,
					Node:       n,
					State:      "error",
				}
				progressChannel <- sp
				continue REPOSITORY_LOOP
			}

			// check if repo feed points to a valid parent node repository
			if !repoExists {
				errorMsg := fmt.Sprintf("repository '%v' does not exist on node %v", repository, n.Fqdn)
				err = errors.New(errorMsg)
				// n.Errors = append(n.Errors, err)
				n.RepositoryError[repository] = err
				sp := SyncProgress{
					Repository: repository,
					Node:       n,
					State:      "error",
				}
				progressChannel <- sp
				continue REPOSITORY_LOOP
			}

			callReport, _, err := client.Repositories.SyncRepository(repository)
			if err != nil {
				// n.Errors = append(n.Errors, err)
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
					// n.Errors = append(n.Errors, err)
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
						// n.Errors = append(n.Errors, err)
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
							// n.Errors = append(n.Errors, err)
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
