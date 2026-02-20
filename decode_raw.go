package main

import (
	"fmt"
	"strconv"
	"strings"
)

func decodeRawMETARToHuman(raw string) error {
	// If multiple lines, take the first non-empty line.
	lines := strings.Split(raw, "\n")
	var metarLine string
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			metarLine = ln
			break
		}
	}
	if metarLine == "" {
		return fmt.Errorf("no METAR content found on stdin")
	}

	tokens := strings.Fields(metarLine)
	if len(tokens) < 3 {
		fmt.Printf("Raw: %s\n", metarLine)
		return nil
	}

	i := 0
	reportType := ""
	if tokens[i] == "METAR" || tokens[i] == "SPECI" {
		reportType = tokens[i]
		i++
	}

	station := tokens[i]
	i++
	obsTime := tokens[i]
	i++

	fmt.Printf("Station: %s\n", station)
	if reportType != "" {
		fmt.Printf("Report: %s\n", reportType)
	}
	fmt.Printf("Observed: %s (DDHHMMZ)\n", obsTime)

	// Optional AUTO/COR
	if i < len(tokens) && (tokens[i] == "AUTO" || tokens[i] == "COR") {
		if tokens[i] == "AUTO" {
			fmt.Printf("Modifier: Automated\n")
		} else {
			fmt.Printf("Modifier: Corrected\n")
		}
		i++
	}

	// Wind
	if i < len(tokens) && strings.HasSuffix(tokens[i], "KT") {
		fmt.Printf("Wind: %s\n", decodeWindToken(tokens[i]))
		i++
		// optional variable dir 180V240
		if i < len(tokens) && strings.Contains(tokens[i], "V") && len(tokens[i]) == 7 {
			fmt.Printf("Wind variation: %s\n", tokens[i])
			i++
		}
	}

	// Visibility
	if i < len(tokens) {
		vis, used := decodeVisibility(tokens[i:], 0)
		if used > 0 {
			fmt.Printf("Visibility: %s\n", vis)
			i += used
		}
	}

	// Weather tokens until sky/temps/alt/RMK
	var wx []string
	for i < len(tokens) {
		t := tokens[i]
		if isSkyToken(t) || isTempDewToken(t) || isAltimeterToken(t) || t == "RMK" {
			break
		}
		wx = append(wx, t)
		i++
	}
	if len(wx) > 0 {
		fmt.Printf("Weather: %s\n", decodeWeatherTokens(strings.Join(wx, " ")))
	}

	// Sky
	var sky []string
	for i < len(tokens) {
		t := tokens[i]
		if isTempDewToken(t) || isAltimeterToken(t) || t == "RMK" {
			break
		}
		if isSkyToken(t) {
			sky = append(sky, decodeSkyToken(t))
			i++
			continue
		}
		break
	}
	if len(sky) > 0 {
		fmt.Printf("Sky: %s\n", strings.Join(sky, ", "))
	}

	// Temp/Dew
	if i < len(tokens) && isTempDewToken(tokens[i]) {
		tc, dc := decodeTempDew(tokens[i])
		fmt.Printf("Temp/Dew: %s°C / %s°C\n", tc, dc)
		i++
	}

	// Altimeter
	if i < len(tokens) && isAltimeterToken(tokens[i]) {
		fmt.Printf("Altimeter: %s\n", decodeAltimeter(tokens[i]))
		i++
	}

	// RMK (raw)
	for i < len(tokens) {
		if tokens[i] == "RMK" {
			fmt.Printf("Remarks: %s\n", strings.Join(tokens[i+1:], " "))
			break
		}
		i++
	}

	fmt.Printf("Raw: %s\n", metarLine)
	return nil
}

func decodeWindToken(tok string) string {
	// 19004KT, VRB03KT, 19012G18KT, 00000KT
	if !strings.HasSuffix(tok, "KT") {
		return tok
	}
	core := strings.TrimSuffix(tok, "KT")
	gust := ""
	if strings.Contains(core, "G") {
		parts := strings.SplitN(core, "G", 2)
		core = parts[0]
		gust = parts[1]
	}

	if strings.HasPrefix(core, "VRB") && len(core) >= 5 {
		spd := core[3:]
		if gust != "" {
			return fmt.Sprintf("Variable at %s kt gusting %s kt", spd, gust)
		}
		return fmt.Sprintf("Variable at %s kt", spd)
	}

	if len(core) < 5 {
		return tok
	}
	dir := core[:3]
	spd := core[3:]
	if dir == "000" && spd == "00" {
		return "Calm"
	}
	if gust != "" {
		return fmt.Sprintf("%s° at %s kt gusting %s kt", dir, spd, gust)
	}
	return fmt.Sprintf("%s° at %s kt", dir, spd)
}

func decodeVisibility(tokens []string, start int) (string, int) {
	if start >= len(tokens) {
		return "", 0
	}
	t0 := tokens[start]
	if strings.HasSuffix(t0, "SM") {
		return humanVisSM(t0), 1
	}
	// "1 1/2SM"
	if start+1 < len(tokens) && strings.HasSuffix(tokens[start+1], "SM") && looksNumeric(tokens[start]) {
		return strings.TrimSpace(tokens[start] + " " + humanVisSM(tokens[start+1])), 2
	}
	return "", 0
}

func humanVisSM(tok string) string {
	core := strings.TrimSuffix(tok, "SM")
	switch {
	case strings.HasPrefix(core, "P"):
		return fmt.Sprintf("Greater than %s statute miles", core[1:])
	case strings.HasPrefix(core, "M"):
		return fmt.Sprintf("Less than %s statute miles", core[1:])
	default:
		return fmt.Sprintf("%s statute miles", core)
	}
}

func looksNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isSkyToken(t string) bool {
	t = strings.ToUpper(t)
	if t == "SKC" || t == "CLR" || t == "NSC" || t == "NCD" {
		return true
	}
	// FEW050, SCT080, BKN250, OVC010, VV002
	if len(t) == 6 {
		p := t[:3]
		if p == "FEW" || p == "SCT" || p == "BKN" || p == "OVC" || p == "VV" {
			_, err := strconv.Atoi(t[3:])
			return err == nil
		}
	}
	return false
}

func decodeSkyToken(t string) string {
	t = strings.ToUpper(t)
	switch t {
	case "SKC":
		return "Sky clear"
	case "CLR":
		return "Clear below 12,000 ft"
	case "NSC":
		return "No significant clouds"
	case "NCD":
		return "No clouds detected"
	}
	if len(t) == 6 {
		cover := t[:3]
		hundreds, _ := strconv.Atoi(t[3:])
		ft := hundreds * 100
		return fmt.Sprintf("%s at %d ft AGL", decodeCloudCover(cover), ft)
	}
	return t
}

func isTempDewToken(t string) bool {
	if !strings.Contains(t, "/") {
		return false
	}
	parts := strings.SplitN(t, "/", 2)
	if len(parts) != 2 {
		return false
	}
	return looksSignedInt(parts[0]) && looksSignedInt(parts[1])
}

func looksSignedInt(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	s = strings.TrimPrefix(s, "M")

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func decodeTempDew(t string) (tempC string, dewC string) {
	parts := strings.SplitN(t, "/", 2)
	tempC = parseMInt(parts[0])
	dewC = parseMInt(parts[1])
	return tempC, dewC
}

func parseMInt(s string) string {
    s = strings.TrimSpace(s)

    neg := strings.HasPrefix(s, "M")
    s = strings.TrimPrefix(s, "M")

    if s == "" {
        return "?"
    }
    if neg {
        return "-" + s
    }
    return s
}

func isAltimeterToken(t string) bool {
	// A2969
	return len(t) == 5 && strings.HasPrefix(strings.ToUpper(t), "A") && looksNumeric(t[1:])
}

func decodeAltimeter(t string) string {
	t = strings.ToUpper(t)
	if !isAltimeterToken(t) {
		return t
	}
	v := t[1:]
	if len(v) != 4 {
		return t
	}
	return fmt.Sprintf("%s.%s inHg", v[:2], v[2:])
}
