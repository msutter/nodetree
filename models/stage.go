package models

import (
	"nodetree/log"
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
			time.Sleep(time.Millisecond)

			// read in channel o start
			<-inc[n.Fqdn]

			if !n.AncestorsHaveError() {
				err := f(n)
				if err != nil {
					log.Error.Println(err)
					n.Errors = append(n.Errors, err)
				}
			} else {
				log.Warning.Printf("Skipping node %v due to errors on folowing Ancestor: %v", n.Fqdn, n.AncestorFqdnsWithErrors())
			}

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

func (s *Stage) Sync() {
	s.SyncedNodeTreeWalker(func(n *Node) error {
		err := n.Sync()
		if err != nil {
			return err
		}
		return err
	})
}

// filter the nodes to sync.
func (s *Stage) SyncByFilters(nodeFqdns []string, nodeTags []string) {
	s.SyncedNodeTreeWalker(func(n *Node) error {
		if n.FqdnsAreDescendant(nodeFqdns) ||
			n.TagsInDescendant(nodeTags) ||
			n.ContainsTags(nodeTags) ||
			n.MatchFqdns(nodeFqdns) {
			err := n.Sync()
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Stage) Show() {
	s.NodeTreeWalker(s.PulpRootNode, func(n *Node) {
		n.Show()
	})
}
