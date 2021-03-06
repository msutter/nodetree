package models

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/msutter/go-pulp/pulp"
	"net/url"
	"strings"
)

type Node struct {
	Fqdn            string
	ApiUser         string
	ApiPasswd       string
	Tags            []string
	Parent          *Node
	Children        []*Node
	Repositories    []Repository
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
	client, err := PulpApiClient(n)
	err = PulpApiSyncRepo(n, client, repositories, progressChannel)
	if err != nil {
		return err
	}
	return
}

func (n *Node) Show() (err error) {
	fmt.Println(n.GetTreeRaw(n.Fqdn))
	return nil
}

func (n *Node) CheckRepositories(repositories []string) (err error) {
	if !n.IsRoot() {
		fmt.Printf("checking repositories on node %v\n", n.Fqdn)
		for _, targetRepository := range repositories {
			fmt.Printf("  - '%v': ", targetRepository)
			if !n.HasRepository(targetRepository) {
				fmt.Printf("error\n")
				fmt.Printf("\n")
				errorMsg := fmt.Sprintf("Could not find repository '%v' on node %v", targetRepository, n.Fqdn)
				err = errors.New(errorMsg)
				n.RepositoryError[targetRepository] = err
				return err
			} else {
				fmt.Printf("pass\n")
			}
		}
		fmt.Printf("\n")

	}
	return
}

func (n *Node) CheckRepositoryFeeds() (err error) {
	if !n.IsRoot() {
		for _, currentRepository := range n.Repositories {
			u, err := url.Parse(currentRepository.Feed)

			// check that the feed is pointing on the parent node
			if u.Host != n.Parent.Fqdn {
				errorMsg := fmt.Sprintf("Repository '%v' has invalid feed '%v'. Parent is '%v'",
					currentRepository.Name,
					currentRepository.Feed,
					n.Parent.Fqdn)

				err = errors.New(errorMsg)
				n.RepositoryError[currentRepository.Name] = err
				return err
			}

			// check that the feed is pointing on an existing repository on the parent node
			pathSlice := strings.Split(u.Path, "/")
			repoInPath := pathSlice[len(pathSlice)-2]

			if !n.Parent.HasRepository(repoInPath) {
				errorMsg := fmt.Sprintf("Repository '%v' does not exist on parent node '%v'",
					repoInPath,
					n.Parent.Fqdn)
				err = errors.New(errorMsg)
				n.RepositoryError[currentRepository.Name] = err
				return err
			}
		}
	}
	return
}

func (n *Node) HasRepository(repository string) bool {
	for _, currentRepository := range n.Repositories {
		if currentRepository.Name == repository {
			return true
		}
	}
	return false
}

func (n *Node) UpdateRepositories() (err error) {
	client, err := PulpApiClient(n)
	if err != nil {
		n.Errors = append(n.Errors, err)
		return err
	}

	var remoteRepos []*pulp.Repository
	remoteRepos, err = PulpApiGetRepos(n, client)

	if err != nil {
		n.Errors = append(n.Errors, err)
		return err
	}

	for _, remoteRepo := range remoteRepos {
		repo := Repository{
			Name: remoteRepo.Id,
			Feed: remoteRepo.Importers[0].ImporterConfig.Feed,
		}
		n.Repositories = append(n.Repositories, repo)
	}

	return
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
