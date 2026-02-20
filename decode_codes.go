package main

import (
	"strings"
)

func decodeCloudCover(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "SKC":
		return "Sky clear"
	case "CLR":
		return "Clear below 12,000 ft"
	case "FEW":
		return "Few clouds"
	case "SCT":
		return "Scattered clouds"
	case "BKN":
		return "Broken clouds"
	case "OVC":
		return "Overcast"
	case "VV":
		return "Vertical visibility"
	default:
		if strings.TrimSpace(code) == "" {
			return "Sky condition unknown"
		}
		return "Clouds (" + code + ")"
	}
}

// decodeWeatherTokens: decode a string containing one or more METAR present-weather tokens.
// Example input: "-RA BR" or "TSRA" or "VCTS".
func decodeWeatherTokens(s string) string {
	parts := strings.Fields(s)
	var out []string
	for _, p := range parts {
		out = append(out, decodeWxToken(p))
	}
	return strings.Join(out, ", ")
}

func decodeWxToken(tok string) string {
	t := strings.ToUpper(strings.TrimSpace(tok))
	if t == "" {
		return tok
	}

	intensity := ""
	switch {
	case strings.HasPrefix(t, "+"):
		intensity = "Heavy "
		t = strings.TrimPrefix(t, "+")
	case strings.HasPrefix(t, "-"):
		intensity = "Light "
		t = strings.TrimPrefix(t, "-")
	}

	// Vicinity
	prox := ""
	if strings.HasPrefix(t, "VC") {
		prox = "In the vicinity: "
		t = strings.TrimPrefix(t, "VC")
	}

	// Descriptor (can appear before precip)
	desc := ""
	for _, d := range []string{"MI", "PR", "BC", "DR", "BL", "SH", "TS", "FZ"} {
		if strings.HasPrefix(t, d) {
			desc = map[string]string{
				"MI": "Shallow ",
				"PR": "Partial ",
				"BC": "Patches of ",
				"DR": "Low drifting ",
				"BL": "Blowing ",
				"SH": "Showers of ",
				"TS": "Thunderstorm with ",
				"FZ": "Freezing ",
			}[d]
			t = strings.TrimPrefix(t, d)
			break
		}
	}

	phen := decodeWxPhenomena(t)
	if phen == "" {
		// unknown token: keep original-ish meaning
		return strings.TrimSpace(prox + intensity + desc + tok)
	}

	return strings.TrimSpace(prox + intensity + desc + phen)
}

func decodeWxPhenomena(code string) string {
	m := map[string]string{
		// precip
		"DZ": "drizzle",
		"RA": "rain",
		"SN": "snow",
		"SG": "snow grains",
		"IC": "ice crystals",
		"PL": "ice pellets",
		"GR": "hail",
		"GS": "small hail / snow pellets",
		"UP": "unknown precipitation",

		// obscuration
		"BR": "mist",
		"FG": "fog",
		"FU": "smoke",
		"VA": "volcanic ash",
		"DU": "widespread dust",
		"SA": "sand",
		"HZ": "haze",
		"PY": "spray",

		// other
		"PO": "dust/sand whirls",
		"SQ": "squalls",
		"FC": "funnel cloud / tornado / waterspout",
		"SS": "sandstorm",
		"DS": "duststorm",
	}

	if v, ok := m[code]; ok {
		return v
	}

	// Combined precip like RASN
	if len(code) == 4 {
		a := code[:2]
		b := code[2:]
		va, oka := m[a]
		vb, okb := m[b]
		if oka && okb {
			return va + " and " + vb
		}
	}
	return ""
}
