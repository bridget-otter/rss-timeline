// Command rss-timeline merges one or more RSS/Atom feeds and prints their
// items sorted by publish date, most recent first.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
)

func main() {
	n := flag.Int("n", 20, "number of items to show (0 for all)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-n count] [file ...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "reads XML feeds from the given files, or from stdin if none are given\n")
		fmt.Fprintf(os.Stderr, "use \"-\" as a filename to read stdin at that position\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	sources := flag.Args()
	if len(sources) == 0 {
		sources = []string{"-"}
	}

	var items []Item
	exitCode := 0
	for _, src := range sources {
		data, err := readSource(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rss-timeline: %s: %v\n", src, err)
			exitCode = 1
			continue
		}
		name := src
		if name == "-" {
			name = "stdin"
		}
		parsed, err := parseFeed(data, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rss-timeline: %v\n", err)
			exitCode = 1
			continue
		}
		items = append(items, parsed...)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Published.After(items[j].Published)
	})

	if *n > 0 && len(items) > *n {
		items = items[:*n]
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	for _, it := range items {
		when := "unknown"
		if !it.Published.IsZero() {
			when = it.Published.Format("2006-01-02 15:04")
		}
		fmt.Fprintf(w, "%s\t%s\t[%s]\t%s\n", when, it.Title, it.Source, it.Link)
	}
	w.Flush()

	os.Exit(exitCode)
}

func readSource(name string) ([]byte, error) {
	if name == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(name)
}
