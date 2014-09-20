// Speedtesting for use with writers.
package speedtest

import (
	"github.com/taruti/monotime"
	"io"
	"strconv"
	"sync"
	"time"
)

type State struct {
	l sync.RWMutex
	m map[string]int64
}

func (s *State) Init() {
	s.l.Lock()
	s.m = map[string]int64{}
	s.l.Unlock()
}

func New() *State {
	var st State
	st.Init()
	return &st
}

func (s *State) Get(remoteHost string) int64 {
	s.l.RLock()
	r := s.m[remoteHost]
	s.l.RUnlock()
	return r
}

func (s *State) Kill(remoteHost string) {
	s.l.Lock()
	delete(s.m, remoteHost)
	s.l.Unlock()
}

func (s *State) WriteSpeedJSON(w io.Writer, remoteHost string) error {
	s.l.RLock()
	r, ok := s.m[remoteHost]
	s.l.RUnlock()
	if !ok {
		el := monotime.NewElapsed()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		total := 0.0
		for total < 1024*1024 {
			select {
			case <-timer.C:
				break
			default:
				n, e := w.Write(spaces)
				if e != nil {
					return e
				}
				total += float64(n)
			}
		}
		r = round(total / el.Current().Seconds())
		s.l.Lock()
		s.m[remoteHost] = r
		s.l.Unlock()
	}
	_, e := io.WriteString(w, strconv.FormatInt(r, 10))
	return e
}

func round(f float64) int64 {
	return int64(f + 0.5)
}

var spaces = func() []byte {
	bs := make([]byte, 4096)
	for i := range bs {
		bs[i] = ' '
	}
	return bs
}()
