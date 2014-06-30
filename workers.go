package main

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Switch is a simple mutex-guarded bool.
type Switch struct {
	State bool
	Mtx   sync.RWMutex
}

// get returns Switch state.
func (s *Switch) get() bool {
	s.Mtx.RLock()
	defer s.Mtx.RUnlock()

	return s.State
}

// set sets the Switch state.
func (s *Switch) set(to bool) {
	s.Mtx.Lock()
	s.State = to
	s.Mtx.Unlock()

	return
}

// parser handles Job processing.
func parser() {
	var matched []string
	var err error

	for job := range parsers {
		switch job.Rule.Action.Type {
		case "regexp":
			matched, err = job.doRegexp()
		case "xpath":
			matched, err = job.doXPath()
		default:
			log.Fatalf("parser: %s unknown action type\n", job.Rule.Action.Type)
		}

		if err != nil {
			log.Println("parser:", err)
			goto finish
		}

		switch job.Rule.Action.Mode {
		case "every":
			if len(matched) < 1 {
				goto finish
			}

			for _, v := range matched {
				target := job.doCommand(v)
				if (job.Rule.Do != nil) && (target != nil) {
					njob := *job
					njob.Rule = job.Rule.Do
					njob.URL = target
					addJob(parsers, &njob, false)
				}
			}
		case "follow":
			if job.Rule.Do != nil {
				njob := *job
				njob.Rule = job.Rule.Do
				addJob(parsers, &njob, false)
			}

			if len(matched) < 1 {
				goto finish
			}

			target := job.doCommand(matched[0])

			njob := *job
			njob.URL = target
			addJob(parsers, &njob, false)
		case "single":
			if len(matched) < 1 {
				goto finish
			}

			target := job.doCommand(matched[0])

			if (job.Rule.Do != nil) && (target != nil) {
				njob := *job
				njob.Rule = job.Rule.Do
				njob.URL = target
				addJob(parsers, &njob, false)
			}
		default:
			log.Fatalf("parser: %s unknown mode\n", job.Rule.Action.Mode)
		}

	finish:
		wg.Done()
	}

	return
}

// downloader downloads new files, updates statistics and stops current run if appropriate.
func downloader() {
	for job := range downloaders {
		if run.get() {
			parts := strings.Split(job.URL.Path, "/")
			name, err := url.QueryUnescape(parts[len(parts)-1])
			if err != nil {
				name = parts[len(parts)-1]
			}

			fullpath := filepath.Join(job.Target.Path, name)

			if _, err := os.Stat(fullpath); err == nil {
				existing := stats.addOld()
				if (job.Target.Bail > 0) && (existing >= job.Target.Bail) {
					if !quiet && run.get() {
						log.Printf("downloader: bail after %d exisiting\n", existing)
					}
					run.set(false)
				}
			} else {
				if !quiet {
					log.Println("SAVE:", fullpath)
				}
				n, took, err := job.saveTo(fullpath)
				if err != nil {
					log.Println("downloader:", err)
				} else {
					stats.addNew(n, took)
				}
			}
		}

		wg.Done()
	}

	return
}

// addJob adds a Job to a queue, in the background and only if appropriate,
// unless forced.
func addJob(where chan *Job, job *Job, force bool) {
	if force || run.get() {
		wg.Add(1)

		go func(c chan *Job, j *Job) {
			c <- j
		}(where, job)

	}

	return
}

// vim: ts=4 sw=4 sts=4
