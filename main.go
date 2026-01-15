package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

func main() {
	forecast := flag.String("forecast", "", `Forecast provider. Supported: "nws"`)
	obs := flag.String("obs", "", "Fetch current raw METAR observation for a station (e.g. KRDU)")
	obsJSON := flag.Bool("json", false, "For --obs: output JSON instead of raw METAR text")
	pretty := flag.Bool("pretty", false, "For --json: pretty-print JSON")
	timeout := flag.Duration("timeout", 10*time.Second, "HTTP timeout (e.g. 5s, 10s)")
	ua := flag.String("user-agent", "metar-tool/0.1 (contact: you@example.com)", "User-Agent to send to APIs")
	flag.Parse()

	// --obs takes precedence and doesn't require --forecast.
	if strings.TrimSpace(*obs) != "" {
		station := normalizeStation(*obs)
		if err := printMETARObs(station, *timeout, *ua, *obsJSON, *pretty); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *forecast == "" {
		usageAndExit("missing --forecast (e.g. --forecast nws mrx) or use --obs KRDU")
	}

	args := flag.Args()
	switch strings.ToLower(*forecast) {
	case "nws":
		if len(args) < 1 {
			usageAndExit(`missing WFO id (e.g. "mrx" or "kmrx")`)
		}
		wfo := normalizeWFO(args[0])
		if err := printLatestAFD(wfo, *timeout, *ua); err != nil {
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
	fmt.Fprintln(os.Stderr, "  metar-tool --obs <station>")
	fmt.Fprintln(os.Stderr, "  metar-tool --forecast nws <wfo>")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  metar-tool --obs KRDU")
	fmt.Fprintln(os.Stderr, "  metar-tool --forecast nws mrx")
	fmt.Fprintln(os.Stderr, "  metar-tool --forecast nws kmrx")
	fmt.Fprintln(os.Stderr, "  metar-tool --obs KTYS --json")
	fmt.Fprintln(os.Stderr, "  metar-tool --obs KTYS --json --pretty")
	os.Exit(2)
}

// normalizeWFO accepts inputs like "mrx", "MRX", "kmrx" and returns "MRX".
func normalizeWFO(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	if strings.HasPrefix(s, "K") && len(s) == 4 {
		// Some people write KMRX; NWS WFO location is MRX.
		s = s[1:]
	}
	return s
}

// normalizeStation accepts "krdu" or "KRDU" and returns "KRDU".
func normalizeStation(s string) string {
	return strings.TrimSpace(strings.ToUpper(s))
}

type productsList struct {
	Graph []productStub `json:"@graph"`
}

type productStub struct {
	ID          string `json:"id"`
	Issued      string `json:"issued"`
	ProductCode string `json:"productCode"`
	ProductName string `json:"productName"`
	ProductText string `json:"productText"` // usually not present in list responses
	Office      string `json:"office"`
	Station     string `json:"station"`
}

type productDetail struct {
	ID          string `json:"id"`
	Issued      string `json:"issued"`
	ProductText string `json:"productText"`
	ProductName string `json:"productName"`
	ProductCode string `json:"productCode"`
}

func printLatestAFD(wfo string, timeout time.Duration, userAgent string) error {
	listURL := fmt.Sprintf("https://api.weather.gov/products/types/AFD/locations/%s", wfo)
	client := &http.Client{Timeout: timeout}

	listBody, err := httpGET(client, listURL, userAgent, "application/geo+json")
	if err != nil {
		return fmt.Errorf("fetch list: %w", err)
	}

	var pl productsList
	if err := json.Unmarshal(listBody, &pl); err != nil {
		return fmt.Errorf("decode list JSON: %w (first 200 bytes: %q)", err, preview(listBody, 200))
	}
	if len(pl.Graph) == 0 {
		return fmt.Errorf("no AFD products found for WFO %s", wfo)
	}

	sort.Slice(pl.Graph, func(i, j int) bool {
		ti := parseIssued(pl.Graph[i].Issued)
		tj := parseIssued(pl.Graph[j].Issued)
		if ti.Equal(time.Time{}) && tj.Equal(time.Time{}) {
			return pl.Graph[i].ID > pl.Graph[j].ID
		}
		if ti.Equal(time.Time{}) {
			return false
		}
		if tj.Equal(time.Time{}) {
			return true
		}
		return ti.After(tj)
	})

	latestID := strings.TrimSpace(pl.Graph[0].ID)
	if latestID == "" {
		return fmt.Errorf("latest product missing id for WFO %s", wfo)
	}

	detailURL := productURL(latestID)

	detailBody, err := httpGET(client, detailURL, userAgent, "application/geo+json")
	if err != nil {
		return fmt.Errorf("fetch product detail: %w", err)
	}

	var pd productDetail
	if err := json.Unmarshal(detailBody, &pd); err != nil {
		return fmt.Errorf("decode product JSON: %w (first 200 bytes: %q)", err, preview(detailBody, 200))
	}

	issued := pd.Issued
	if issued == "" {
		issued = pl.Graph[0].Issued
	}

	fmt.Printf("%s (%s) - issued %s\n", wfo, nonEmpty(pd.ProductName, "Area Forecast Discussion"), issued)
	fmt.Println(strings.Repeat("-", 72))
	fmt.Print(pd.ProductText)
	if !strings.HasSuffix(pd.ProductText, "\n") {
		fmt.Println()
	}
	return nil
}

func printMETARObs(station string, timeout time.Duration, userAgent string, asJSON bool, pretty bool) error {
	u, _ := url.Parse("https://aviationweather.gov/api/data/metar")
	q := u.Query()
	q.Set("ids", station)
	q.Set("taf", "false")

	if asJSON {
		q.Set("format", "json")
	} else {
		q.Set("format", "raw")
	}

	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: timeout}

	accept := "text/plain"
	if asJSON {
		accept = "application/json"
	}

	body, err := httpGET(client, u.String(), userAgent, accept)
	if err != nil {
		return fmt.Errorf("fetch metar: %w", err)
	}

	if asJSON {
		trim := strings.TrimSpace(string(body))
		if trim == "" || trim == "[]" {
			return fmt.Errorf("no METAR returned for %s", station)
		}

		if pretty {
			var v any
			if err := json.Unmarshal(body, &v); err != nil {
				return fmt.Errorf("decode JSON: %w (first 200 bytes: %q)", err, preview(body, 200))
			}
			out, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Errorf("encode JSON: %w", err)
			}
			fmt.Println(string(out))
			return nil
		}

		// Raw JSON as returned by the API
		fmt.Println(trim)
		return nil
	}

	out := strings.TrimSpace(string(body))
	if out == "" {
		return fmt.Errorf("no METAR returned for %s", station)
	}
	fmt.Println(out)
	return nil
}

func httpGET(client *http.Client, urlStr, userAgent, accept string) ([]byte, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s: HTTP %d: %s", urlStr, resp.StatusCode, preview(body, 300))
	}
	return body, nil
}

func parseIssued(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func preview(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}

func nonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func productURL(idOrURL string) string {
	s := strings.TrimSpace(idOrURL)
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	return "https://api.weather.gov/products/" + s
}
