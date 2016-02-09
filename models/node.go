package models

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Node struct {
	Fqdn     string
	Tags     []string
	Parent   *Node
	Children []*Node
	SyncPath []string
	Depth    int
	Errors   []error
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

func (n *Node) DestroyMe() {
	if !n.IsRoot() {
		index := n.BrotherIndex()
		n.Parent.Children = append(n.Parent.Children[:index], n.Parent.Children[index+1:]...)
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

func (n *Node) Sync() (err error) {
	fmt.Printf("%vstart Syncing %v at Depth %v\n", strings.Repeat("+----", n.Depth), n.Fqdn, n.Depth)
	random := rand.Intn(1000) + 100
	time.Sleep(time.Duration(random) * time.Millisecond)
	fmt.Printf("%vstop Syncing %v at Depth %v\n", strings.Repeat("+----", n.Depth), n.Fqdn, n.Depth)
	if n.Fqdn == "pulp-lab-111.local" {
		msg := fmt.Sprintf("Sync failed on %v!", n.Fqdn)
		err = errors.New(msg)
	}
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) Show() (err error) {
	if n.Depth == 0 {
		fmt.Printf("├── %v\n", n.Fqdn)
	} else {
		fmt.Printf("    ")
	}

	for i := 1; i < n.Depth; i++ {
		if n.Depth != 0 {

			// is my ancestor at Depth x the last brother
			depthAncestor := n.GetAncestorByDepth(i)

			if depthAncestor.IslastBrother() {
				fmt.Printf("    ")
			} else {
				fmt.Printf("│   ")
			}
		} else {
			fmt.Printf("    ")
		}
	}

	if n.Depth != 0 {
		if n.IslastBrother() {
			fmt.Printf("└── %v\n", n.Fqdn)
		} else {
			fmt.Printf("├── %v\n", n.Fqdn)
		}
	}
	return nil
}
