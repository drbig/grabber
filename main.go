package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"
)

// Current version, with date
const (
	VERSION      = "0.8.2"
	NAME         = "grabber"
	URL          = "https://github.com/drbig/grabber"
	AUTHOR_NAME  = "Piotr S. Staszewski"
	AUTHOR_EMAIL = "p.staszewski@gmail.com"
)

// Internal global variables
var (
	client      *http.Client
	stats       *Stats
	run         *Switch
	parsers     chan *Job
	downloaders chan *Job
	wg          sync.WaitGroup
)

// Global argument variables
var (
	quiet bool
)

func main() {
	start := time.Now()

	var flgDls = flag.Int("dls", 4, "number of downloaders")
	var flgPrs = flag.Int("prs", 8, "number of parsers")
	var flgStd = flag.Bool("stdout", false, "log to stdout")
	var flgVer = flag.Bool("ver", false, "show version and exit")
	flag.BoolVar(&quiet, "quiet", false, "suppress additional output")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] file.json\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *flgVer {
		fmt.Println(NAME, VERSION)
		return
	}

	if *flgStd {
		log.SetOutput(os.Stdout)
	}

	if len(flag.Args()) != 1 {
		log.Fatalln("Please specify single rules file")
	}

	log.Println("Loading rules..")
	rules, err := loadRules(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	run = &Switch{}

	ctrl := make(chan bool)
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	parsers = make(chan *Job, *flgPrs)
	for i := 0; i < *flgPrs; i++ {
		go parser()
	}
	downloaders = make(chan *Job, *flgDls)
	for i := 0; i < *flgDls; i++ {
		go downloader()
	}

	log.Println("Executing...")
loop:
	for _, target := range rules {
		startTarget := time.Now()
		log.Println("Target:", target.Name)

		root, err := filepath.Abs(target.Path)
		if err != nil {
			log.Fatalln("ERROR", err)
		}

		if _, err := os.Stat(root); os.IsNotExist(err) {
			log.Fatalln("ERROR", err)
		}

		base, err := url.Parse(target.URL)
		if err != nil {
			log.Fatalln("ERROR", err)
		}

		target.Path = root
		stats = &Stats{}
		job := &Job{
			Target: &target,
			Rule:   target.Do,
			URL:    base,
		}

		run.set(true)
		addJob(parsers, job, true)
		go func(c chan bool) {
			wg.Wait()
			c <- true

			return
		}(ctrl)

	inner:
		for {
			select {
			case <-ctrl:
				break inner
			case <-sig:
				if run.get() {
					log.Println("SIGNAL: Interrupted, graceful exit...")
					log.Println("SIGNAL: Another signal will exit now")
					run.set(false)
				} else {
					break loop
				}
			}
		}

		log.Printf("Finished target: %s (took %s)\n", target.Name, time.Since(startTarget))
		if !stats.isEmpty() {
			log.Println("Download statistics:", stats)
		}
	}

	log.Printf("All done (took %s).", time.Since(start))
}

// vim: ts=4 sw=4 sts=4
