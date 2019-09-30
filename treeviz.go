package mux

import (
	"bytes"
	"fmt"
	"sort"
)

const (
	ec = "   "
	sc = "│  "
	mc = "├──"
	lc = "└──"
)

func Tree(n *Route) string {
	if n == nil {
		return ""
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s [%s]", n.path, n.kind))
	ms := n.Methods()
	hasMethods := len(ms) > 0
	if hasMethods {
		buf.WriteString(" (")
		for i, m := range n.Methods() {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(m)
		}
		buf.WriteString(")")
	}
	buf.WriteString("\n")
	children := n.children
	sort.Sort(children)
	for i, c := range children {
		cString := Tree(c)
		sl := siblingLevels(c)
		var prefix string
		for l := 1; l < c.Level(); l++ {
			if contains(sl, l) {
				prefix += sc
			} else {
				prefix += ec
			}
		}
		buf.WriteString(prefix)
		if i < len(n.children)-1 {
			buf.WriteString(mc)
		} else {
			buf.WriteString(lc)
		}
		buf.WriteString(cString)
	}
	return buf.String()
}

func siblingLevels(n *Route) []int {
	if n.IsRoot() {
		return []int{}
	}
	res := siblingLevels(n.parent)
	if !n.IsLastChild() {
		res = append(res, n.Level())
	}
	return res
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
