package models

import (
	"fmt"
	// "github.com/gosuri/uiprogress"
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
	//  initialize the channels maps
	inc := make(map[string]chan *Node)
	outc := make(map[string]chan *Node)
	// reset the Leafs slice
	s.Leafs = s.Leafs[:0]

	// initialize the routines and the Depth map
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		// initialize the channels
		inc[n.Fqdn] = make(chan *Node)
		outc[n.Fqdn] = make(chan *Node)
		// start the go routines
		go func() {
			// Give some time to execute the main process
			// (not sure about this. There are probebly better ways)
			time.Sleep(time.Millisecond * 10)

			// read in channel o start
			<-inc[n.Fqdn]

			// if !n.AncestorsHaveError() {
			err := f(n)
			if err != nil {
				// log.Error.Println(err)
				n.Errors = append(n.Errors, err)
			}
			// } else {
			// 	// log.Warning.Printf("Skipping node %v due to errors on folowing Ancestor: %v", n.Fqdn, n.AncestorFqdnsWithErrors())
			// }

			// write on each children to unlock start
			for _, child := range n.Children {
				child.Parent = n
				inc[child.Fqdn] <- child.Parent
				close(inc[child.Fqdn])
			}

			// Write the out channel to finish the leafs
			if n.IsLeaf() {
				outc[n.Fqdn] <- n
				close(outc[n.Fqdn])
			}

		}()
	})

	// start the execucution on root node
	inc[s.PulpRootNode.Fqdn] <- s.PulpRootNode
	close(inc[s.PulpRootNode.Fqdn])

	for _, leaf := range s.Leafs {
		<-outc[leaf.Fqdn]
	}
}

func (s *Stage) Sync(repository string) {
	progressChannels := make(map[string]chan SyncProgress)

	s.SyncedNodeTreeWalker(func(n *Node) (err error) {
		progressChannels[n.Fqdn] = make(chan SyncProgress)
		go func() {
			time.Sleep(time.Millisecond * 10)
			for sp := range progressChannels[n.Fqdn] {
				switch sp.State {
				case "running", "finished":
					fmt.Printf("%v: %v [%v/%v]\n", n.Fqdn, sp.State, sp.ItemsLeft, sp.ItemsTotal)
				default:
					fmt.Printf("%v: %v\n", n.Fqdn, sp.State)
				}
			}
		}()
		err = n.Sync(repository, progressChannels[n.Fqdn])
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
