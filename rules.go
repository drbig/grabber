package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Action holds parsing parameters.
//
// Mode - as supported by parser()
// Type - as suppotyed by Job (do*() methods)
// Args - variable args, as above
type Action struct {
	Mode string
	Type string
	Args []interface{}
}

// Rule holds action, command and possibly links to a sub-rule.
//
// Command - as supported by Job (doCommand() method)
// Action - as defined by Action
// Do - may be a pointer to a sub-rule, or nil for a leaf
type Rule struct {
	Command string
	Action  Action
	Do      *Rule
}

// Target holds general scraping parameters.
//
// Name - for user information
// URL - base URL, this will be used for the initial Job
// Bail - if more than zero will stop scraping if number of encountred existing files exceeds it
// Path - mandatory path where files will be saved (use "./" if not saving files)
// Headers - for headers injection, or nil
// Do - root rule
type Target struct {
	Name    string
	URL     string
	Bail    int
	Path    string
	Headers *map[string]string
	Do      *Rule
}

// loadRules loads the JSON file without any syntactical checking.
func loadRules(name string) (t []Target, err error) {
	handle, err := os.Open(name)
	if err != nil {
		return
	}
	defer handle.Close()

	raw, err := ioutil.ReadAll(handle)
	if err != nil {
		return
	}

	err = json.Unmarshal(raw, &t)
	if err != nil {
		return
	}

	return
}

// vim: ts=4 sw=4 sts=4
