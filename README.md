# grabber

Grabber is a concurrent declarative web scraper and downloader.

Features:
  * Simple tree-like JSON configuration
  * XPath and Regexp extractors
  * Parallel parsing and extraction
  * Parallel download
  * Ability to bail out early (e.g. for updating)
  * Fails fast on config errors, tolerates web errors
  * Follow, every, and single extraction modes
  * Multiple XPaths or Regexps per stage
  * Multi-grouped regexps with a separator (e.g. extract to CSV)
  * It's rather fast

Run `grabber -h` to see command-line options.

See `examples/` directory and consult the code to learn the format
of the config files.

Note that for `tumblr.json` you'll need to replace all occurrences of 
`{{name}}` with a proper account (subdomain) name and all occurrences of 
`{{paging}}` with the (XPath's text() operator) contents of what your target 
blog uses for 'next page' (or semantically equivalent). You may also notice 
that the format is already template-friendly, so you can easily write a script 
for generating per-blog templates.

The examples provided are certainly not exhaustive.

Advice:

Remember you can build your config iteratively by using the `log` command,
so that you make sure the current level works as it should before going
further.

When downloading:

For the first run set `bail` to `0` and use options `-quiet -stdout`,
you may also wish to pipe the output of the run to `tee log`.
Then inspect the output/logfile for any errors. If it looks ok set `bail` to
something reasonable e.g. if you have 10 assets per page set it to 20.

## Todo / Bugs

  * Needs testing 'in the wild'
  * Better documentation
  * Ability to use `Content-Disposition`
  * Full config parsing and error checking during load
  * Test suite

## Copyright

Copyright (c) 2014 Piotr S. Staszewski

Absolutely no warranty. See LICENSE.txt for details.
