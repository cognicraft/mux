package mux

import (
	"fmt"
	"net/http"
	"sort"
)

func NewRoute(path string) *Route {
	return &Route{
		path:     path,
		kind:     ClassifyKind(path),
		handlers: map[string]http.Handler{},
	}
}

type Route struct {
	parent   *Route
	path     string
	kind     Kind
	children Routes
	handlers map[string]http.Handler
}

func (r *Route) IsRoot() bool {
	return r.parent == nil
}

func (r *Route) IsLeaf() bool {
	return len(r.children) == 0
}

func (r *Route) IsFirstChild() bool {
	if r.IsRoot() {
		return true
	} else {
		return r.parent.FirstChild() == r
	}
}

func (r *Route) IsLastChild() bool {
	if r.IsRoot() {
		return true
	} else {
		return r.parent.LastChild() == r
	}
}

func (r *Route) FirstChild() *Route {
	if len(r.children) > 0 {
		return r.children[0]
	}
	return nil
}

func (r *Route) LastChild() *Route {
	if len(r.children) > 0 {
		return r.children[len(r.children)-1]
	}
	return nil
}

func (r *Route) Siblings() Routes {
	if r.IsRoot() {
		return []*Route{}
	}
	sibs := []*Route{}
	for _, c := range r.Parent().Children() {
		if c != r {
			sibs = append(sibs, c)
		}
	}
	return sibs
}

func (r *Route) Parent() *Route {
	return r.parent
}

func (r *Route) Children() Routes {
	return r.children
}

func (r *Route) Kind() Kind {
	return r.kind
}

func (r *Route) Level() int {
	if r.IsRoot() {
		return 0
	} else {
		return r.parent.Level() + 1
	}
}

func (r *Route) Path() string {
	return r.path
}

func (r *Route) FullPath() string {
	if r.parent == nil {
		return r.path
	} else {
		return r.parent.FullPath() + r.path
	}
}

func (r *Route) GET(h http.Handler) {
	r.SetHandler("GET", h)
}

func (r *Route) POST(h http.Handler) {
	r.SetHandler("POST", h)
}

func (r *Route) OPTIONS(h http.Handler) {
	r.SetHandler("OPTIONS", h)
}

func (r *Route) HEAD(h http.Handler) {
	r.SetHandler("HEAD", h)
}

func (r *Route) PUT(h http.Handler) {
	r.SetHandler("PUT", h)
}

func (r *Route) PATCH(h http.Handler) {
	r.SetHandler("PATCH", h)
}

func (r *Route) DELETE(h http.Handler) {
	r.SetHandler("DELETE", h)
}

func (r *Route) SetHandler(m string, h http.Handler) {
	r.handlers[m] = h
}

func (r *Route) Handler(m string) (http.Handler, bool) {
	h, ok := r.handlers[m]
	return h, ok
}

func (r *Route) Methods() []string {
	ms := []string{}
	for m, _ := range r.handlers {
		ms = append(ms, m)
	}
	sort.Strings(ms)
	return ms
}

func (r *Route) String() string {
	return fmt.Sprintf("[%s, %v]", r.kind, r.Methods())
}

func (r *Route) Set(nr *Route) error {
	if r.path != nr.path {
		return fmt.Errorf("unable to replace route: incorrect path")
	}
	r.kind = nr.kind
	r.children = nr.children
	for _, c := range r.children {
		c.parent = r
	}
	r.handlers = nr.handlers
	return nil
}

func (r *Route) Append(child *Route) error {
	if r.hasComplexChild() && child.isComplex() {
		return fmt.Errorf("not possible")
	}
	child.parent = r
	r.children = append(r.children, child)
	return nil
}

func (r *Route) isComplex() bool {
	return r.Kind() == KindParameter || r.Kind() == KindCatchAll
}

func (r *Route) hasComplexChild() bool {
	return r.Children().FindOne(ByKind(KindParameter, KindCatchAll)) != nil
}

func (r *Route) Route(path string) *Route {
	head, tail := split(path)
	if len(head) > 0 {
		if c := r.Children().FindOne(ByPath(head)); c != nil {
			return c.Route(tail)
		}
		if r.kind == KindCatchAll {
			panic(fmt.Errorf("not possible"))
		}
		c := NewRoute(head)
		err := r.Append(c)
		if err != nil {
			panic(err)
		}
		return c.Route(tail)
	}
	return r
}

func (r *Route) Match(path string) (*Route, map[string]string) {
	if len(path) == 0 {
		return r, map[string]string{}
	}
	head, tail := split(path)
	if c := r.Children().FindOne(ByPath(head)); c != nil {
		return c.Match(tail)
	}
	if c := r.Children().FindOne(ByKind(KindParameter, KindCatchAll)); c != nil {
		switch c.Kind() {
		case KindParameter:
			vars := map[string]string{
				c.path: head,
			}
			h, sub := c.Match(tail)
			for k, v := range sub {
				vars[k] = v
			}
			return h, vars
		case KindCatchAll:
			vars := map[string]string{
				c.path: path,
			}
			return c, vars
		}
	}
	// no suitable route exists
	return nil, map[string]string{}
}

type Routes []*Route

func (rs Routes) Len() int           { return len(rs) }
func (rs Routes) Less(i, j int) bool { return rs[i].path < rs[j].path }
func (rs Routes) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

func (rs Routes) FindOne(f func(r *Route) bool) *Route {
	for _, n := range rs {
		if f(n) {
			return n
		}
	}
	return nil
}

func (rs Routes) Filter(f func(r *Route) bool) Routes {
	fns := []*Route{}
	for _, n := range rs {
		if f(n) {
			fns = append(fns, n)
		}
	}
	return fns
}

func ByKind(ks ...Kind) func(*Route) bool {
	return func(n *Route) bool {
		for _, k := range ks {
			if n.kind == k {
				return true
			}
		}
		return false
	}
}

func ByPath(pcs ...string) func(*Route) bool {
	return func(r *Route) bool {
		for _, pc := range pcs {
			if r.path == pc {
				return true
			}
		}
		return false
	}
}

type Kind uint8

const (
	KindStatic Kind = iota
	KindParameter
	KindCatchAll
)

func ClassifyKind(path string) Kind {
	switch path[0] {
	case ':':
		return KindParameter
	case '*':
		return KindCatchAll
	default:
		return KindStatic
	}
}

func (k Kind) String() string {
	switch k {
	case KindStatic:
		return "Static"
	case KindParameter:
		return "Parameter"
	case KindCatchAll:
		return "CatchAll"
	default:
		return "unknown"
	}
}

func split(path string) (string, string) {
	for i, r := range path {
		if r == '/' {
			if i == 0 {
				return path[0:1], path[1:]
			}
			return path[:i], path[i:]
		}
	}
	return path, ""
}
