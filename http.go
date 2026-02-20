package main

import (
	"fmt"
	"io"
	"net/http"
)

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

func preview(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}

func nonEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
