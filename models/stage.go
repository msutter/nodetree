package models

import (
	"fmt"
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

func (s *Stage) Init() {
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {})
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

	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		var wg sync.WaitGroup
		inWg[n.Fqdn] = &wg
		inWg[n.Fqdn].Add(1)
	})

	// Set a waitgroup for all leafs
	var leafsWaitGroup sync.WaitGroup
	leafsCount := len(s.Leafs)
	leafsWaitGroup.Add(leafsCount)

	// initialize the routines and the Depth map
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		// initialize the waitgroup
		go func() {
			// Give some time to execute the main process
			// (not sure about this. There are probebly better ways)
			time.Sleep(time.Millisecond * 50)

			// read in channel o start
			inWg[n.Fqdn].Wait()
			// <-inc[n.Fqdn]
			fmt.Printf("got in for %v\n", n.Fqdn)
			// if !n.AncestorsHaveError() {
			err := f(n)
			if err != nil {
				// log.Error.Println(err)
				n.Errors = append(n.Errors, err)
			}

			// Write the out channel to finish the leafs
			if n.IsLeaf() {
				fmt.Printf("leaf done before on %v\n", n.Fqdn)
				leafsWaitGroup.Done()
				fmt.Printf("leaf done after on %v\n", n.Fqdn)
			}

			// write on each children to unlock start
			for _, child := range n.Children {
				fmt.Printf("send in for %v\n", child.Fqdn)
				inWg[child.Fqdn].Done()
			}

		}()
	})

	// start the execucution on root node
	inWg[s.PulpRootNode.Fqdn].Done()

	fmt.Printf("wait on all leafs\n")
	leafsWaitGroup.Wait()
	fmt.Printf("done all leafs\n")

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
