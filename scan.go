package desktop

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const bufferSize = 32 * 1024

type scan struct {
	e    map[int][]*Entry
	errs chan error
	in   chan *scanEntry
	sync.Mutex
	sync.WaitGroup
}

type scanEntry struct {
	i int
	f *os.File
}

func Scan(dirs []string) (map[int][]*Entry, error) {
	s := &scan{e: make(map[int][]*Entry), errs: make(chan error), in: make(chan *scanEntry)}

	for i := 0; i < runtime.GOMAXPROCS(-1); i++ {
		go scanner(s)
	}

	for i, dir := range dirs {
		i, dir := i, dir

		s.Add(1)
		go scanDir(i, dir, s)
	}

	done := make(chan bool, 1)
	go func() {
		s.Wait()
		close(s.in)

		done <- true
	}()

	select {
	case err := <-s.errs:
		return nil, err
	case <-done:
		return s.e, nil
	}
}

func scanner(s *scan) {
	var (
		buf       = make([]byte, bufferSize)
		scanEntry *scanEntry
		entry     *Entry
		err       error
	)

	for scanEntry = range s.in {
		entry, err = Parse(scanEntry.f, buf)
		scanEntry.f.Close()
		if err != nil {
			s.errs <- err
			s.Done()
			return
		} else if entry == nil {
			s.Done()
			continue
		}

		s.Lock()
		s.e[scanEntry.i] = append(s.e[scanEntry.i], entry)
		s.Unlock()

		s.Done()
	}
}

func scanFunc(i int, s *scan) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {
		s.Add(1)

		go func() {
			if err != nil {
				s.errs <- err
				s.Done()
				return
			}

			if f == nil || f.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".desktop") {
				s.Done()
				return
			}

			f, err := os.OpenFile(path, os.O_RDONLY, 0644)
			if err != nil {
				s.errs <- err
				s.Done()
				return
			}

			s.in <- &scanEntry{i: i, f: f}
		}()

		return nil
	}
}

func scanDir(i int, dir string, s *scan) {
	defer s.Done()

	err := filepath.Walk(dir, scanFunc(i, s))
	if err != nil {
		s.errs <- err
	}
}
