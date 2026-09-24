package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

// Item is a feed entry normalized across RSS and Atom.
type Item struct {
	Title     string
	Link      string
	GUID      string
	Published time.Time
	Source    string
}

type rssDoc struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title string `xml:"title"`
		Items []struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			GUID    string `xml:"guid"`
			PubDate string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

type atomDoc struct {
	XMLName xml.Name `xml:"feed"`
	Title   string   `xml:"title"`
	Entries []struct {
		Title string `xml:"title"`
		ID    string `xml:"id"`
		Links []struct {
			Href string `xml:"href"`
			Rel  string `xml:"rel"`
		} `xml:"link"`
		Updated   string `xml:"updated"`
		Published string `xml:"published"`
	} `xml:"entry"`
}

// rootElement finds the name of the first XML start element without
// unmarshaling the whole document, so we know whether to parse RSS or Atom.
func rootElement(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", err
		}
		if start, ok := tok.(xml.StartElement); ok {
			return start.Name.Local, nil
		}
	}
}

// parseFeed reads RSS 2.0 or Atom XML and returns normalized items tagged
// with source (typically a filename or "stdin").
func parseFeed(data []byte, source string) ([]Item, error) {
	root, err := rootElement(data)
	if err != nil {
		return nil, fmt.Errorf("%s: not valid XML: %w", source, err)
	}

	switch root {
	case "rss":
		var doc rssDoc
		if err := xml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("%s: %w", source, err)
		}
		items := make([]Item, 0, len(doc.Channel.Items))
		for _, it := range doc.Channel.Items {
			items = append(items, Item{
				Title:     it.Title,
				Link:      it.Link,
				GUID:      it.GUID,
				Published: parseTime(it.PubDate),
				Source:    source,
			})
		}
		return items, nil

	case "feed":
		var doc atomDoc
		if err := xml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("%s: %w", source, err)
		}
		items := make([]Item, 0, len(doc.Entries))
		for _, e := range doc.Entries {
			link := ""
			for _, l := range e.Links {
				if l.Rel == "" || l.Rel == "alternate" {
					link = l.Href
					break
				}
			}
			when := e.Published
			if when == "" {
				when = e.Updated
			}
			items = append(items, Item{
				Title:     e.Title,
				Link:      link,
				GUID:      e.ID,
				Published: parseTime(when),
				Source:    source,
			})
		}
		return items, nil

	default:
		return nil, fmt.Errorf("%s: unrecognized root element <%s>, expected <rss> or <feed>", source, root)
	}
}

// timeLayouts covers the date formats actually seen in the wild: RFC 822
// with numeric or named zones for RSS, RFC 3339 for Atom, and a couple of
// sloppy variants publishers ship anyway.
var timeLayouts = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC3339,
	"2006-01-02T15:04:05Z0700",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// dedupeItems drops items that repeat across feeds, which happens when the
// same post shows up in a podcast feed and its site feed, or a feed is
// listed twice on the command line. GUID is preferred when present since
// it's the field publishers use to identify a post across revisions; link
// is the fallback. Items with neither are kept as-is since we have no way
// to tell them apart from anything else.
func dedupeItems(items []Item) []Item {
	seen := make(map[string]bool, len(items))
	out := make([]Item, 0, len(items))
	for _, it := range items {
		key := ""
		switch {
		case it.GUID != "":
			key = "guid:" + it.GUID
		case it.Link != "":
			key = "link:" + it.Link
		}
		if key != "" {
			if seen[key] {
				continue
			}
			seen[key] = true
		}
		out = append(out, it)
	}
	return out
}

func parseTime(s string) time.Time {
	for _, layout := range timeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
