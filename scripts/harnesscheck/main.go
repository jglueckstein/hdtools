// harnesscheck runs one mechanical constraint against the tree.
//
// The five subcommands are the parts of the old agent rules that a
// parser can decide. Prose quality, MixedCaps taste, and medical
// copy stay reviews. This program does not fix what it finds.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: harnesscheck naming|layout|errors|docs|preamble")
		os.Exit(2)
	}
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var findings []string
	switch os.Args[1] {
	case "naming":
		findings, err = checkNaming(root)
	case "layout":
		findings, err = checkLayout(root)
	case "errors":
		findings, err = checkErrors(root)
	case "docs":
		findings, err = checkDocs(root)
	case "preamble":
		findings, err = checkPreamble(root)
	default:
		fmt.Fprintf(os.Stderr, "unknown check %q\n", os.Args[1])
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, finding := range findings {
		fmt.Fprintln(os.Stderr, finding)
	}
	if len(findings) > 0 {
		os.Exit(1)
	}
}
