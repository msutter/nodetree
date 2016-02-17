package models

import (
	"fmt"
	// "errors"
	// "github.com/msutter/go-pulp/pulp"
	// "github.com/msutter/nodetree/log"
	"math/rand"
	// "strings"
	"bytes"
	"math"
	"time"
)

type Node struct {
	Fqdn      string
	ApiUser   string
	ApiPasswd string
	Tags      []string
	Parent    *Node
	Children  []*Node
	SyncPath  []string
	Depth     int
	Errors    []error
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

func (n *Node) AncestorFqdnsWithErrors() (ancestorFqdns []string) {
	for _, ancestor := range n.AncestorsWithErrors() {
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

type SyncProgress struct {
	State      string
	SizeTotal  int
	SizeLeft   int
	ItemsTotal int
	ItemsLeft  int
	Warning    string
	Error      error
}

func (s *SyncProgress) ItemsDone() int {
	return s.ItemsTotal - s.ItemsLeft
}

func (s *SyncProgress) ItemsPercent() int {
	if s.ItemsTotal == 0 {
		return 100
	} else {
		return int(math.Floor(float64(s.ItemsDone()) / float64(s.ItemsTotal) * float64(100)))

	}
}

func (s *SyncProgress) SizeDone() int {
	return s.SizeTotal - s.SizeLeft
}

func (s *SyncProgress) SizePercent() int {
	if s.SizeTotal == 0 {
		return 100
	} else {
		return int(math.Floor(float64(s.SizeDone()) / float64(s.SizeTotal) * float64(100)))
	}
}

func (n *Node) Sync(repository string) (err error) {

	if !n.IsRoot() {
		randTime := time.Duration(rand.Intn(2000) + 100)
		time.Sleep(randTime * time.Millisecond)
	}
	// 	if n.AncestorsHaveError() {
	// 		sp := SyncProgress{
	// 			State:   "skipped",
	// 			Warning: fmt.Sprintf("skipping sync due to errors on ancestor node %v", n.AncestorFqdnsWithErrors()[0]),
	// 		}
	// 		progressChannels <- sp
	// 		time.Sleep(20 * time.Millisecond)
	// 		return
	// 	}
	// 	// create the API client
	// 	client, err := pulp.NewClient(n.Fqdn, n.ApiUser, n.ApiPasswd, nil)
	// 	if err != nil {
	// 		sp := SyncProgress{
	// 			State: "error",
	// 			Error: err,
	// 		}
	// 		progressChannels <- sp
	// 		time.Sleep(20 * time.Millisecond)
	// 		return err
	// 	}

	// 	callReport, _, err := client.Repositories.SyncRepository(repository)

	// 	if err != nil {
	// 		sp := SyncProgress{
	// 			State: "error",
	// 			Error: err,
	// 		}

	// 		progressChannels <- sp
	// 		time.Sleep(20 * time.Millisecond)
	// 		return err
	// 	}

	// 	syncTaskId := callReport.SpawnedTasks[0].TaskId

	// 	state := "init"
	// 	for (state != "finished") && (state != "error") {

	// 		task, _, err := client.Tasks.GetTask(syncTaskId)
	// 		if err != nil {
	// 			sp := SyncProgress{
	// 				State: "error",
	// 				Error: err,
	// 			}
	// 			progressChannels <- sp
	// 			time.Sleep(20 * time.Millisecond)
	// 			// close(progressChannels)
	// 			return err
	// 		}

	// 		if task.State == "error" {
	// 			errorMsg := task.ProgressReport.YumImporter.Metadata.Error
	// 			err = errors.New(errorMsg)
	// 			sp := SyncProgress{
	// 				State: "error",
	// 				Error: err,
	// 			}

	// 			progressChannels <- sp
	// 			time.Sleep(20 * time.Millisecond)
	// 			// close(progressChannels)
	// 			return err
	// 		}

	// 		state = task.State
	// 		sp := SyncProgress{
	// 			State: state,
	// 		}

	// 		if task.ProgressReport.YumImporter.Content != nil {
	// 			sp.SizeTotal = task.ProgressReport.YumImporter.Content.SizeTotal
	// 			sp.SizeLeft = task.ProgressReport.YumImporter.Content.SizeLeft
	// 			sp.ItemsTotal = task.ProgressReport.YumImporter.Content.ItemsTotal
	// 			sp.ItemsLeft = task.ProgressReport.YumImporter.Content.ItemsLeft
	// 		} else {
	// 			fmt.Printf("%v: missing content\n", n.Fqdn)
	// 		}

	// 		progressChannels <- sp
	// 		time.Sleep(500 * time.Millisecond)

	// 	}
	// }

	return
}

func (n *Node) Show() (err error) {
	fmt.Printf(n.GetTreeRaw(n.Fqdn))
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
