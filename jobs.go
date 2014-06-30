package main

import (
	"errors"
	"fmt"
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/xml"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

// Job holds parameters for processing.
//
// Target - current Target
// Rule - current Rule
// URL - URL associated with this job
type Job struct {
	Target *Target
	Rule   *Rule
	URL    *url.URL
}

// doRequest does header injection and HTTP GET.
func (j *Job) doRequest() (res *http.Response, err error) {
	req, err := http.NewRequest("GET", j.URL.String(), nil)
	if err != nil {
		return
	}

	if j.Target.Headers != nil {
		for k, v := range *j.Target.Headers {
			req.Header.Add(k, v)
		}
	}

	res, err = client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		res.Body.Close()
		err = errors.New(fmt.Sprintf("getData: %d %s", res.StatusCode, j.URL.String()))
	}

	return
}

// getBody retrieves res.Body.
func (j *Job) getBody() (data []byte, err error) {
	res, err := j.doRequest()
	if err != nil {
		return
	}
	defer res.Body.Close()

	data, err = ioutil.ReadAll(res.Body)

	return
}

// doXPath does XPath filtering.
func (j *Job) doXPath() (matched []string, err error) {
	var raw []xml.Node
	matched = make([]string, 0)

	data, err := j.getBody()
	if err != nil {
		return
	}

	doc, err := gokogiri.ParseHtml(data)
	if err != nil {
		return
	}
	defer doc.Free()

	for _, v := range j.Rule.Action.Args {
		xpath, ok := v.(string)
		if !ok {
			log.Fatalln("doXPath: not a string")
			return
		}

		raw, err = doc.Search(xpath)
		if err != nil {
			return
		}

		for _, v := range raw {
			matched = append(matched, v.String())
		}
	}

	return
}

// doRegexp does Regexp filtering.
func (j *Job) doRegexp() (matched []string, err error) {
	var rxp *regexp.Regexp
	var tmp, srxp, sep string
	matched = make([]string, 0)

	data, err := j.getBody()
	if err != nil {
		return
	}

	for _, v := range j.Rule.Action.Args {
		args, ok := v.([]interface{})
		if !ok {
			log.Fatalln("doRegexp: not an array")
		}

		if len(args) != 2 {
			log.Fatalln("doRegexp: wrong length")
		}

		srxp, ok = args[0].(string)
		if !ok {
			log.Fatalln("doRegexp: not a string (1)")
		}

		sep, ok = args[1].(string)
		if !ok {
			log.Fatalln("doRegexp: not a string (2)")
		}

		rxp, err = regexp.Compile(srxp)
		if err != nil {
			log.Fatalln("doRegexp:", err)
		}

		raw := rxp.FindAllSubmatch(data, -1)

		for _, a := range raw {
			if len(a) > 1 {
				tmp = string(a[1])

				for _, v := range a[2:] {
					tmp += sep + string(v)
				}
				matched = append(matched, tmp)
			}
		}
	}

	return
}

// doCommand executes command associated with the job.
//
// It will also dispatch appropriate Jobs to downloaders.
func (j *Job) doCommand(raw string) (target *url.URL) {
	var datum string

	wg.Add(1)

	turl, err := url.Parse(raw)
	if err != nil {
		target = nil
		datum = raw
	} else {
		target = j.URL.ResolveReference(turl)
		datum = target.String()
	}

	switch j.Rule.Command {
	case "download":
		if target == nil {
			log.Fatalf("doCommand: %s is not a URL\n", datum)
		}
		njob := *j
		njob.URL = target
		addJob(downloaders, &njob, true)
	case "log":
		log.Println("LOG:", datum)
	case "none":
		// nop
	case "print":
		fmt.Println(datum)
	default:
		log.Fatalf("doCommand: %s unknown command\n", j.Rule.Command)
	}

	wg.Done()
	return
}

// saveTo saves the Job's URL to a file.
func (j *Job) saveTo(fullpath string) (n int64, took time.Duration, err error) {
	start := time.Now()

	res, err := j.doRequest()
	if err != nil {
		return
	}
	defer res.Body.Close()

	handle, err := os.Create(fullpath)
	if err != nil {
		return
	}
	defer handle.Close()

	n, err = io.Copy(handle, res.Body)
	took = time.Since(start)

	return
}

// vim: ts=4 sw=4 sts=4
