package models

import (
	// "fmt"
	"fmt"
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

		// make the errors container
		node.RepositoryError = make(map[string]error)

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
			f(n)
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

func (s *Stage) HasError() bool {
	returnValue := false
	for _, n := range s.Nodes {
		if (len(n.Errors) > 0) || len(n.RepositoryError) > 0 {
			returnValue = true
		}
	}
	return returnValue
}

func (s *Stage) SyncAll(progressChannel chan SyncProgress) {
	return
}

func (s *Stage) Sync(repositories []string, progressChannel chan SyncProgress) {
	defer close(progressChannel)

	// Use the synced walk
	s.SyncedNodeTreeWalker(func(n *Node) (serr error) {
		// Execute the sync
		n.Sync(repositories, progressChannel)
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

func (s *Stage) CheckAll() {
	// get all repositories exising on the root pulp node
	s.PulpRootNode.UpdateRepositories()
	var repositories []string
	fmt.Printf("\nfound following repositories on root node %v\n", s.PulpRootNode.Fqdn)
	for _, rootRepository := range s.PulpRootNode.Repositories {
		repositories = append(repositories, rootRepository.Name)
		fmt.Printf("  - '%v'\n", rootRepository.Name)
	}
	fmt.Printf("\n")
	s.Check(repositories)
}

func (s *Stage) Check(repositories []string) {
	s.Init()
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		n.UpdateRepositories()
		n.CheckRepositories(repositories)
		n.CheckRepositoryFeeds()
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
