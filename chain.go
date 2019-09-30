package mux

import "net/http"

type Decorator func(http.Handler) http.Handler

type Chain struct {
	decorators []Decorator
}

func NewChain(decorators ...Decorator) Chain {
	c := Chain{}
	c.decorators = append(c.decorators, decorators...)
	return c
}

func (c Chain) Append(decorators ...Decorator) Chain {
	newCons := make([]Decorator, len(c.decorators))
	copy(newCons, c.decorators)
	newCons = append(newCons, decorators...)
	newChain := NewChain(newCons...)
	return newChain
}

func (c Chain) Then(h http.Handler) http.Handler {
	final := h
	for i := len(c.decorators) - 1; i >= 0; i-- {
		final = c.decorators[i](final)
	}
	return final
}

func (c Chain) ThenFunc(f http.HandlerFunc) http.Handler {
	return c.Then(f)
}
