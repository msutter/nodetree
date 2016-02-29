package models

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/msutter/go-pulp/pulp"
	"time"
)

type Node struct {
	Fqdn            string
	ApiUser         string
	ApiPasswd       string
	Tags            []string
	Parent          *Node
	Children        []*Node
	SyncPath        []string
	Depth           int
	TreePosition    int
	Errors          []error
	RepositoryError map[string]error
}

// Matches the given fqdn?
func (n *Node) MatchFqdn(fqdn string) bool {
	if n.Fqdn == fqdn {
		return true
	} else {
		return false
	}
}

// Matches the given fqdns?
func (n *Node) MatchFqdns(fqdns []string) bool {
	ret := false
	for _, fqdn := range fqdns {
		if n.MatchFqdn(fqdn) {
			ret = true
		}
	}
	return ret

}

// Contains the given tag?
func (n *Node) ContainsTag(tag string) bool {
	ret := false
	for _, nodeTag := range n.Tags {
		if nodeTag == tag {
			ret = true
			break
		}
	}
	return ret
}

// Contains the given tags?
func (n *Node) ContainsTags(tags []string) bool {
	ret := false
	for _, tag := range tags {
		if n.ContainsTag(tag) {
			ret = true
		}
	}
	return ret
}

// Is a Leaf?
func (n *Node) IsLeaf() bool {
	if len(n.Children) == 0 {
		return true
	} else {
		return false
	}
}

// Is a Root?
func (n *Node) IsRoot() bool {
	if n.Parent == nil {
		return true
	} else {
		return false
	}
}

func (n *Node) AncestorTreeWalker(f func(*Node)) {
	parent := n.Parent
	if parent != nil {
		f(parent) // resurse
		parent.AncestorTreeWalker(f)
	}
}

// Is Fqdn a Ancestor?
func (n *Node) FqdnIsAncestor(ancestorFqdn string) bool {
	returnValue := false
	n.AncestorTreeWalker(func(ancestor *Node) {
		if ancestor.Fqdn == ancestorFqdn {
			returnValue = true
		}
	})
	return returnValue
}

// Are Fqdns a Ancestor?
func (n *Node) FqdnsAreAncestor(ancestorFqdns []string) bool {
	returnValue := false
	for _, ancestorFqdn := range ancestorFqdns {
		if n.FqdnIsAncestor(ancestorFqdn) {
			returnValue = true
		}
	}
	return returnValue
}

// Get Ancestors
func (n *Node) Ancestors() (ancestors []*Node) {
	n.AncestorTreeWalker(func(ancestor *Node) {
		ancestors = append(ancestors, ancestor)
	})
	return
}

// Get Ancestors by Depth id
func (n *Node) GetAncestorByDepth(depth int) (depthAncestor *Node) {
	n.AncestorTreeWalker(func(ancestor *Node) {
		if ancestor.Depth == depth {
			depthAncestor = ancestor
		}
	})
	return
}

// Has Error
func (n *Node) HasError() bool {
	returnValue := false
	if (len(n.Errors) > 0) || len(n.RepositoryError) > 0 {
		returnValue = true
	}
	return returnValue
}

// Ancestor has Error
func (n *Node) AncestorsHaveError() bool {
	returnValue := false
	for _, ancestor := range n.Ancestors() {
		if ancestor.Errors != nil {
			returnValue = true
		}
	}
	return returnValue
}

// Ancestor has Error
func (n *Node) AncestorsHaveRepositoryError(repository string) bool {
	returnValue := false
	for _, ancestor := range n.Ancestors() {
		if ancestor.RepositoryError[repository] != nil {
			returnValue = true
		}
	}
	return returnValue
}

func (n *Node) AncestorFqdnsWithErrors() (ancestorFqdns []string) {
	for _, ancestor := range n.AncestorsWithErrors() {
		ancestorFqdns = append(ancestorFqdns, ancestor.Fqdn)
	}
	return
}

func (n *Node) AncestorFqdnsWithRepositoryError(repository string) (ancestorFqdns []string) {
	for _, ancestor := range n.AncestorsWithRepositoryError(repository) {
		ancestorFqdns = append(ancestorFqdns, ancestor.Fqdn)
	}
	return
}

// Ancestor has Error
func (n *Node) AncestorsWithErrors() (ancestors []*Node) {
	n.AncestorTreeWalker(func(ancestor *Node) {
		if ancestor.Errors != nil {
			ancestors = append(ancestors, ancestor)
		}
	})
	return
}

// Ancestor has Error
func (n *Node) AncestorsWithRepositoryError(repository string) (ancestors []*Node) {
	n.AncestorTreeWalker(func(ancestor *Node) {
		if ancestor.RepositoryError[repository] != nil {
			ancestors = append(ancestors, ancestor)
		}
	})
	return
}

func (n *Node) ChildTreeWalker(f func(*Node)) {
	for _, node := range n.Children {
		f(node) // resurse
		node.ChildTreeWalker(f)
	}
}

func (n *Node) IslastBrother() bool {
	if n.lastBrother() == n {
		return true
	} else {
		return false
	}
}

func (n *Node) BrotherIndex() (iret int) {
	if !n.IsRoot() {
		for i, child := range n.Parent.Children {
			if n == child {
				iret = i
			}
		}
	}
	return iret
}

func (n *Node) lastBrother() (lastBrother *Node) {
	brothers := n.Parent.Children
	lastBrother = brothers[len(brothers)-1]
	return
}

// Is Fqdn a Descendant?
func (n *Node) FqdnIsDescendant(childFqdn string) bool {
	returnValue := false
	n.ChildTreeWalker(func(child *Node) {
		if child.MatchFqdn(childFqdn) {
			returnValue = true
		}
	})
	return returnValue
}

// Are Fqdns a Descendant?
func (n *Node) FqdnsAreDescendant(childFqdns []string) bool {
	returnValue := false
	n.ChildTreeWalker(func(child *Node) {
		if child.MatchFqdns(childFqdns) {
			returnValue = true
		}
	})
	return returnValue
}

// Is Fqdn a Descendant?
func (n *Node) TagsInDescendant(childTags []string) bool {
	returnValue := false
	n.ChildTreeWalker(func(child *Node) {
		if child.ContainsTags(childTags) {
			returnValue = true
		}
	})
	return returnValue
}

func (n *Node) Sync(repositories []string, progressChannel chan SyncProgress) (err error) {

	waitingTimeout := 10
	waitingRetries := 3

	if !n.IsRoot() {
		n.RepositoryError = make(map[string]error)

		// create the API client
		client, err := pulp.NewClient(n.Fqdn, n.ApiUser, n.ApiPasswd, nil)
		if err != nil {
			n.Errors = append(n.Errors, err)
			sp := SyncProgress{
				Node:  n,
				State: "error",
			}
			progressChannel <- sp
		}

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
					sp.SizeTotal = task.ProgressReport.YumImporter.Content.SizeTotal
					sp.SizeLeft = task.ProgressReport.YumImporter.Content.SizeLeft
					sp.ItemsTotal = task.ProgressReport.YumImporter.Content.ItemsTotal
					sp.ItemsLeft = task.ProgressReport.YumImporter.Content.ItemsLeft
				}

				progressChannel <- sp
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
	return
}

func (n *Node) Show() (err error) {
	fmt.Println(n.GetTreeRaw(n.Fqdn))
	return nil
}

func (n *Node) GetTreeRaw(msg string) (treeRaw string) {
	var buffer bytes.Buffer
	if n.Depth == 0 {
		buffer.WriteString(fmt.Sprintf("\n├─ %v", msg))
	} else {
		buffer.WriteString(fmt.Sprintf("   "))
	}
	for i := 1; i < n.Depth; i++ {
		if n.Depth != 0 {
			// is my ancestor at Depth x the last brother
			depthAncestor := n.GetAncestorByDepth(i)
			if depthAncestor.IslastBrother() {
				buffer.WriteString(fmt.Sprintf("   "))
			} else {
				buffer.WriteString(fmt.Sprintf("│  "))
			}
		} else {
			buffer.WriteString(fmt.Sprintf("   "))
		}
	}
	if n.Depth != 0 {
		if n.IslastBrother() {
			buffer.WriteString(fmt.Sprintf("└─ %v", msg))
		} else {
			buffer.WriteString(fmt.Sprintf("├─ %v", msg))
		}
	}
	return buffer.String()
}
