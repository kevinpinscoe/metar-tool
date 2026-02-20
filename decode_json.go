package main

import (
	"fmt"
	"strings"
	"time"
)

type awCloud struct {
	Cover string `json:"cover"` // FEW/SCT/BKN/OVC/VV
	Base  *int   `json:"base"`  // feet AGL (often)
}

type awMetar struct {
	RawOb    string    `json:"rawOb"`
	ICAOId   string    `json:"icaoId"`
	ObsTime  string    `json:"obsTime"`
	WDir     *int      `json:"wdir"`
	WSpd     *int      `json:"wspd"`
	WGst     *int      `json:"wgst"`
	Visib    *string   `json:"visib"`
	Altim    *string   `json:"altim"`
	Temp     *string   `json:"temp"`
	Dewp     *string   `json:"dewp"`
	WxString *string   `json:"wxString"`
	Clouds   []awCloud `json:"clouds"`
}

func printHumanFromAWJSON(m awMetar) {
	station := strings.TrimSpace(m.ICAOId)
	if station == "" {
		station = "(unknown station)"
	}

	fmt.Printf("Station: %s\n", station)

	if t, err := time.Parse(time.RFC3339, strings.TrimSpace(m.ObsTime)); err == nil {
		fmt.Printf("Observed: %s UTC\n", t.UTC().Format("2006-01-02 15:04"))
	} else if strings.TrimSpace(m.ObsTime) != "" {
		fmt.Printf("Observed: %s\n", strings.TrimSpace(m.ObsTime))
	}

	fmt.Printf("Wind: %s\n", humanWindFromJSON(m.WDir, m.WSpd, m.WGst))

	if m.Visib != nil && strings.TrimSpace(*m.Visib) != "" {
		fmt.Printf("Visibility: %s SM\n", strings.TrimSpace(*m.Visib))
	}

	if m.WxString != nil && strings.TrimSpace(*m.WxString) != "" {
		fmt.Printf("Weather: %s\n", decodeWeatherTokens(strings.TrimSpace(*m.WxString)))
	}

	if len(m.Clouds) > 0 {
		var parts []string
		for _, c := range m.Clouds {
			parts = append(parts, humanCloudLayer(c))
		}
		fmt.Printf("Sky: %s\n", strings.Join(parts, ", "))
	}

	if (m.Temp != nil && strings.TrimSpace(*m.Temp) != "") || (m.Dewp != nil && strings.TrimSpace(*m.Dewp) != "") {
		fmt.Printf("Temp/Dew: %s°C / %s°C\n", nonEmptyPtr(m.Temp, "?"), nonEmptyPtr(m.Dewp, "?"))
	}

	if m.Altim != nil && strings.TrimSpace(*m.Altim) != "" {
		fmt.Printf("Altimeter: %s inHg\n", strings.TrimSpace(*m.Altim))
	}

	if strings.TrimSpace(m.RawOb) != "" {
		fmt.Printf("Raw: %s\n", strings.TrimSpace(m.RawOb))
	}
}

func nonEmptyPtr(p *string, fallback string) string {
	if p == nil {
		return fallback
	}
	if strings.TrimSpace(*p) == "" {
		return fallback
	}
	return strings.TrimSpace(*p)
}

func humanWindFromJSON(wdir, wspd, wgst *int) string {
	if wspd == nil {
		return "unknown"
	}
	dir := "VRB"
	if wdir != nil && *wdir >= 0 {
		dir = fmt.Sprintf("%03d°", *wdir)
	}
	if wgst != nil && *wgst > 0 {
		return fmt.Sprintf("%s %d kt gusting %d kt", dir, *wspd, *wgst)
	}
	return fmt.Sprintf("%s %d kt", dir, *wspd)
}

func humanCloudLayer(c awCloud) string {
	cover := strings.TrimSpace(strings.ToUpper(c.Cover))
	base := ""
	if c.Base != nil && *c.Base > 0 {
		base = fmt.Sprintf(" at %d ft AGL", *c.Base)
	}
	return fmt.Sprintf("%s%s", decodeCloudCover(cover), base)
}
