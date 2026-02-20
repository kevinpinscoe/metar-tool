package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var Version = ""

type options struct {
	forecast  string
	obs       string
	obsJSON   bool
	pretty    bool
	timeout   time.Duration
	userAgent string
	output    string
	verbose   bool
	decode    bool
}

func main() {
	var opt options

	showVersion := flag.Bool("version", false, "Print version and exit")

	flag.StringVar(&opt.forecast, "forecast", "", `Forecast provider. Supported: "nws"`)
	flag.StringVar(&opt.obs, "obs", "", "Fetch current raw METAR observation for a station (e.g. KRDU)")
	flag.BoolVar(&opt.obsJSON, "json", false, "For --obs: output JSON instead of raw METAR text")
	flag.BoolVar(&opt.pretty, "pretty", false, "For --json: pretty-print JSON")
	flag.DurationVar(&opt.timeout, "timeout", 10*time.Second, "HTTP timeout (e.g. 5s, 10s)")
	flag.StringVar(&opt.userAgent, "user-agent", "metar-tool/0.1 (contact: you@example.com)", "User-Agent to send to APIs")
	flag.StringVar(&opt.output, "output", "", "Write normal output to this file (errors still go to stderr)")
	flag.BoolVar(&opt.verbose, "verbose", false, "Verbose logging to stderr")
	flag.BoolVar(&opt.decode, "decode", false, "Decode piped METAR/JSON from stdin into human-readable format")

	flag.Parse()

	if *showVersion {
		fmt.Printf("%s %s\n", "metar-tool", Version)
		return
	}

	// Redirect stdout to file if requested
	if strings.TrimSpace(opt.output) != "" {
		f, err := os.Create(opt.output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: open output file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: close output file: %v\n", err)
			}
		}()
		os.Stdout = f
	}

	if opt.decode {
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: read stdin: %v\n", err)
			os.Exit(1)
		}
		if strings.TrimSpace(string(in)) == "" {
			usageAndExit("--decode expects input on stdin (pipe JSON or raw METAR text)")
		}
		if err := decodeFromStdin(in); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: decode failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// --obs mode
	if strings.TrimSpace(opt.obs) != "" {
		station := normalizeStation(opt.obs)
		if err := printMETARObs(station, opt.timeout, opt.userAgent, opt.obsJSON, opt.pretty); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// --forecast mode
	if strings.TrimSpace(opt.forecast) == "" {
		usageAndExit(`missing --forecast (e.g. --forecast nws mrx) or use --obs KRDU`)
	}
	args := flag.Args()

	switch strings.ToLower(opt.forecast) {
	case "nws":
		if len(args) < 1 {
			usageAndExit(`missing WFO id (e.g. "mrx" or "kmrx")`)
		}
		wfo := normalizeWFO(args[0])
		if err := printLatestAFD(wfo, opt.timeout, opt.userAgent); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
	default:
		usageAndExit(`unsupported --forecast value (supported: "nws")`)
	}
}

func usageAndExit(msg string) {
	fmt.Fprintln(os.Stderr, "ERROR:", msg)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, " metar-tool --version")
	fmt.Fprintln(os.Stderr, " metar-tool --obs KRDU")
	fmt.Fprintln(os.Stderr, " metar-tool --obs KTYS --json --pretty")
	fmt.Fprintln(os.Stderr, " metar-tool --forecast nws mrx")
	fmt.Fprintln(os.Stderr, " metar-tool --decode   # reads stdin (pipe JSON or raw METAR)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, " metar-tool --obs KTYS")
	fmt.Fprintln(os.Stderr, " metar-tool --obs KTYS --json | metar-tool --decode")
	fmt.Fprintln(os.Stderr, " metar-tool --obs KTYS | metar-tool --decode")
	os.Exit(2)
}
