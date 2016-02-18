package models

import (
	"bytes"
	// "fmt"
)

type SyncErrors struct {
	Nodes []*Node
}

func (s *SyncErrors) Error() (Errs string) {
	var buffer bytes.Buffer
	for _, n := range s.Nodes {
		buffer.WriteString(n.Fqdn)
		buffer.WriteString(":\n")
		for _, e := range n.Errors {
			buffer.WriteString(" - ")
			buffer.WriteString(e.Error())
			buffer.WriteString("\n")
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func (s *SyncErrors) Any() bool {
	any := false
	for _, n := range s.Nodes {
		if len(n.Errors) > 0 {
			any = true
		}
	}
	return any
}
