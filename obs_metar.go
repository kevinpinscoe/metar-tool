package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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
