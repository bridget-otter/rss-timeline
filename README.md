# rss-timeline

I follow a couple dozen blogs and podcasts by RSS. No reader I've used shows
a single "what's new, across everything, in order" list without asking me to
create an account and hand over my subscription list to some server. I just
want to point a tool at the XML I already have and get a sorted list back.

`rss-timeline` does exactly that: give it one or more RSS or Atom files (or
pipe one in), and it prints the items sorted by publish date, most recent
first. That's the whole tool.

It does not fetch URLs itself — there's no HTTP client in here on purpose,
just a query over XML you already have. Fetch with `curl` or `wget` and pipe
the result in.

## Usage

```
rss-timeline [-n count] [file ...]
```

If no files are given, it reads from stdin. Use `-` as a filename to read
stdin at that position, which lets you mix piped input with files on disk.

```
# a single feed already fetched to disk
rss-timeline blog.xml

# fetch and pipe straight in
curl -s https://example.com/feed.xml | rss-timeline

# merge several feeds into one timeline, keep the 5 newest
rss-timeline -n 5 blog.xml podcast.atom newsletter.xml

# mix a live fetch with a saved copy
curl -s https://example.com/feed.xml | rss-timeline - archive.xml
```

Output is one line per item: timestamp, title, source, link.

```
2026-09-14 09:03  Why we rewrote the scheduler  [blog.xml]        https://example.com/posts/scheduler
2026-09-12 18:41  Episode 88: caches are hard    [podcast.atom]   https://example.com/ep/88
```

`-n 0` prints every item instead of truncating.

## Feed support

Both RSS 2.0 (`<rss><channel><item>`) and Atom (`<feed><entry>`) are
understood. The root element of the XML decides which parser runs, so you
don't need to say which format a file is. Dates are parsed against the
common RFC 822 and RFC 3339 variants publishers actually use; anything that
doesn't match prints as "unknown" rather than failing the whole feed.

If one file in a batch fails to parse, that file is skipped with a warning
on stderr and the rest are still processed.

## Building

Standard library only, no external dependencies:

```
go build -o rss-timeline .
```

## License

MIT, see LICENSE.
