package mux

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func NewTiming() *Timing {
	return &Timing{
		time: time.Now(),
	}
}

type Timing struct {
	time  time.Time
	marks []string
}

func (ts *Timing) Mark(key string) {
	ms := float64(time.Since(ts.time)) / float64(time.Millisecond)
	ts.marks = append(ts.marks, fmt.Sprintf("%s=%.1f", key, ms))
	ts.time = time.Now()
}

func (ts *Timing) WriteTo(header http.Header) {
	header.Add("Server-Timing", strings.Join(ts.marks, ", "))
}
