package models

import (
	// "fmt"
	// "github.com/gosuri/uiprogress"
	"sync"
	"time"
)

type Stage struct {
	Name         string
	PulpRootNode *Node
	Leafs        []*Node
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
	if node.IsLeaf() {
		s.Leafs = append(s.Leafs, node)
	}
	// set the leafs
	f(node)
	for _, n := range node.Children {
		// set the depth
		n.Depth = node.Depth + 1
		// set the parent node
		n.Parent = node
		// resurse
		s.NodeTreeWalker(n, f)
	}
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

	inWg := make(map[string]*sync.WaitGroup)
	// initialize the tree (waitgroups, prents, depth, etc)
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		var wg sync.WaitGroup
		inWg[n.Fqdn] = &wg
		inWg[n.Fqdn].Add(1)
	})

	// Set a waitgroup for synconization of compteted leafs
	var leafsWaitGroup sync.WaitGroup
	leafsCount := len(s.Leafs)
	leafsWaitGroup.Add(leafsCount)

	// initialize the routines and the Depth map
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

func (s *Stage) Sync(repository string) {
	s.SyncedNodeTreeWalker(func(n *Node) (err error) {

		err = n.Sync(repository)
		if err != nil {
			return err
		}
		return
	})
}

func (s *Stage) Show() {
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		n.Show()
	})
}

// get a filtered stage.
func (s *Stage) Filter(nodeFqdns []string, nodeTags []string) (filteredStage *Stage) {
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
