package models

import (
	"fmt"
	tm "github.com/buger/goterm"
	"sync"
	"time"
)

type Stage struct {
	Name         string
	PulpRootNode *Node
	Leafs        []*Node
	Nodes        []*Node
}

// Matches the given fqdn?
func (s Stage) MatchName(name string) bool {
	if s.Name == name {
		return true
	} else {
		return false
	}
}

func (s *Stage) NodeTreeWalker(node *Node, f func(*Node)) {
	f(node)
	for _, n := range node.Children {
		s.NodeTreeWalker(n, f)
	}
}

func (s *Stage) Init() {
	pos := 1
	s.NodeTreeWalker(s.PulpRootNode, func(node *Node) {
		s.Nodes = append(s.Nodes, node)
		// set treePosition
		node.TreePosition = pos
		pos++
		// set the leafs
		if node.IsLeaf() {
			s.Leafs = append(s.Leafs, node)
		}
		for _, n := range node.Children {
			// set the depth
			n.Depth = node.Depth + 1
			// set the parent node
			n.Parent = node
		}
	})
}

func (s *Stage) GetNodeByFqdn(nodeFqdn string) (node *Node) {
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		if n.Fqdn == nodeFqdn {
			node = n
		}
	})
	return node
}

func (s *Stage) SyncedNodeTreeWalker(f func(n *Node) error) {
	s.Init()
	// initialize the tree (waitgroups, prents, depth, etc)
	inWg := make(map[string]*sync.WaitGroup)
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		var wg sync.WaitGroup
		inWg[n.Fqdn] = &wg
		inWg[n.Fqdn].Add(1)
	})

	// Set a waitgroup for synconization of compteted leafs
	var leafsWaitGroup sync.WaitGroup
	leafsCount := len(s.Leafs)
	leafsWaitGroup.Add(leafsCount)
	// Walk the tree with syncronization
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		go func() {
			time.Sleep(time.Millisecond * 50)
			// Wait
			inWg[n.Fqdn].Wait()
			// execute the function
			err := f(n)
			if err != nil {
				// log.Error.Println(err)
				n.Errors = append(n.Errors, err)
			}
			// Set done on waitgroup
			if n.IsLeaf() {
				leafsWaitGroup.Done()
			}
			// set Done on each child unlock start
			for _, child := range n.Children {
				inWg[child.Fqdn].Done()
			}

		}()
	})
	// start the execucution on root node
	inWg[s.PulpRootNode.Fqdn].Done()
	// Wait on all leafs to complete
	leafsWaitGroup.Wait()
}

func (s *Stage) Sync(repository string) (err error) {
	// Use the synced walk
	s.SyncedNodeTreeWalker(func(n *Node) (err error) {
		// Create a progress channel
		progressChannel := make(chan SyncProgress)
		// Read the progressChannel for this node until it's closed
		go func() {
			state := "init"
			for sp := range progressChannel {
				switch sp.State {
				case "skipped":
					line := fmt.Sprintf("%v %v", n.GetTreeRaw(n.Fqdn), sp.State)
					tm.Printf(tm.Color(tm.Bold(line), tm.MAGENTA))
					tm.Flush()
				case "error":
					line := fmt.Sprintf("%v %v", n.GetTreeRaw(n.Fqdn), sp.State)
					tm.Printf(tm.Color(tm.Bold(line), tm.RED))
					tm.Flush()
				case "running":
					if state != sp.State {
						line := fmt.Sprintf("%v %v", n.GetTreeRaw(n.Fqdn), sp.State)
						tm.Printf(tm.Color(line, tm.YELLOW))
						tm.Flush()
					}
					state = sp.State
				case "finished":
					line := fmt.Sprintf("%v %v", n.GetTreeRaw(n.Fqdn), sp.State)
					tm.Printf(tm.Color(tm.Bold(line), tm.GREEN))
					tm.Flush()
				}
			}
		}()
		// Execute the sync
		n.Sync(repository, progressChannel)
		if err != nil {
			return err
		}
		return
	})
	return
}

func (s *Stage) Show() {
	s.Init()
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		n.Show()
	})
}

// get a filtered stage.
func (s *Stage) Filter(nodeFqdns []string, nodeTags []string) (filteredStage *Stage) {
	s.Init()
	filteredStage = s
	s.NodeTreeWalker(filteredStage.PulpRootNode, func(n *Node) {
		childsToKeep := []*Node{}
		for _, child := range n.Children {
			if child.FqdnsAreDescendant(nodeFqdns) ||
				child.MatchFqdns(nodeFqdns) ||
				child.TagsInDescendant(nodeTags) ||
				child.ContainsTags(nodeTags) {
				childsToKeep = append(childsToKeep, child)
			}
		}
		n.Children = childsToKeep
	})
	return filteredStage
}
