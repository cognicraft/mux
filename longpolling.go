package mux

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Signal chan struct{}

type SignalFunc func(key string)

var managerID = 0

func NewManager(options ...func(*Manager) error) (*Manager, error) {
	managerID++
	m := &Manager{
		id:           managerID,
		signals:      map[string]Signal{},
		lastSignals:  map[string]time.Time{},
		sessionEtags: map[string]map[string]string{},
	}
	return m, m.SetOption(options...)
}

type Manager struct {
	id           int
	mutex        sync.RWMutex
	signals      map[string]Signal
	lastSignals  map[string]time.Time
	sessionEtags map[string]map[string]string
}

func (m *Manager) SetOption(options ...func(*Manager) error) error {
	for _, opt := range options {
		if err := opt(m); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Signal(prefix string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, _ := range m.sessionEtags {
		for key, _ := range m.sessionEtags[id] {
			if strings.HasPrefix(key, prefix) {
				delete(m.sessionEtags[id], key)
				if len(m.sessionEtags[id]) == 0 {
					delete(m.sessionEtags, id)
				}
			}
		}
	}
	m.lastSignals[prefix] = time.Now()

	for key, signal := range m.signals {
		if strings.HasPrefix(key, prefix) {
			close(signal)
			delete(m.signals, key)
		}
	}
}

func (m *Manager) Await(key string) Signal {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, exists := m.signals[key]
	if !exists {
		s = make(Signal)
		m.signals[key] = s
	}
	return s
}

func (m *Manager) recordEtag(requestTime time.Time, id string, key string, value string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for prefix, last := range m.lastSignals {
		if strings.HasPrefix(key, prefix) && last.After(requestTime) {
			// do not record an old tag
			return
		}
	}
	if _, ok := m.sessionEtags[id]; !ok {
		m.sessionEtags[id] = map[string]string{}
	}
	m.sessionEtags[id][key] = value
}

func (m *Manager) etag(sid string, key string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	etags, ok := m.sessionEtags[sid]
	if !ok {
		return ""
	}
	etag, _ := etags[key]
	return etag
}

func extractKey(r *http.Request) string {
	return r.URL.Path
}

const (
	HeaderEtag        = "Etag"
	HeaderLongPoll    = "Long-Poll"
	HeaderIfNoneMatch = "If-None-Match"
)

func LongPolling() (middleware Decorator, signal SignalFunc) {

	manager, _ := NewManager()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sid := "default"
			key := extractKey(r)
			longPoll := r.Header.Get(HeaderLongPoll)
			ifNoneMatch := r.Header.Get(HeaderIfNoneMatch)
			if longPoll != "" {
				if ifNoneMatch != "" {
					if etag := manager.etag(sid, key); etag == ifNoneMatch {
						timeout, _ := strconv.Atoi(longPoll)
						if timeout < 1 {
							timeout = 1
						}
						select {
						case <-time.After(time.Duration(timeout) * time.Second):
							w.Header().Add(HeaderEtag, ifNoneMatch)
							w.WriteHeader(http.StatusNotModified)
							return
						case <-manager.Await(key):
						}
					}
				}
			}

			rw := &responseWriter{
				ResponseWriter: w,
				buffer:         &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}

			requestTime := time.Now()

			next.ServeHTTP(rw, r)

			switch rw.statusCode {
			case http.StatusOK:
				etag := rw.Header().Get(HeaderEtag)
				if etag == "" {
					// only calculate an etag if none was generated yet
					etag = fmt.Sprintf(`"%d-%x"`, rw.buffer.Len(), sha1.Sum(rw.buffer.Bytes()))
					w.Header().Add(HeaderEtag, etag)
				}
				manager.recordEtag(requestTime, sid, key, etag)
				if ifNoneMatch == etag {
					w.WriteHeader(http.StatusNotModified)
					return
				}
				w.WriteHeader(http.StatusOK)
				io.Copy(w, rw.buffer)
				return
			default:
				w.WriteHeader(rw.statusCode)
				io.Copy(w, rw.buffer)
				return
			}
		})
	}, manager.Signal
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	buffer     *bytes.Buffer
}

func (w *responseWriter) Write(bs []byte) (int, error) {
	return w.buffer.Write(bs)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
