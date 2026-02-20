package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type productsList struct {
	Graph []productStub `json:"@graph"`
}

type productStub struct {
	ID          string `json:"id"`
	Issued      string `json:"issued"`
	ProductCode string `json:"productCode"`
	ProductName string `json:"productName"`
	ProductText string `json:"productText"`
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

func parseIssued(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func productURL(idOrURL string) string {
	s := strings.TrimSpace(idOrURL)
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	return "https://api.weather.gov/products/" + s
}
