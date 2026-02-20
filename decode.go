package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func decodeFromStdin(in []byte) error {
	s := strings.TrimSpace(string(in))

	// Heuristic JSON detection
	if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
		// Try aviationweather JSON array
		var arr []awMetar
		if err := json.Unmarshal([]byte(s), &arr); err == nil && len(arr) > 0 {
			for i, m := range arr {
				if i > 0 {
					fmt.Println()
				}
				printHumanFromAWJSON(m)
			}
			return nil
		}

		// Try single object
		var obj awMetar
		if err := json.Unmarshal([]byte(s), &obj); err == nil && strings.TrimSpace(obj.RawOb) != "" {
			printHumanFromAWJSON(obj)
			return nil
		}

		// Fallback: pretty-print arbitrary JSON
		var v any
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			return fmt.Errorf("stdin looked like JSON but could not decode: %w", err)
		}
		out, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("re-encode JSON: %w", err)
		}
		fmt.Println(string(out))
		return nil
	}

	// Otherwise treat as raw METAR
	return decodeRawMETARToHuman(s)
}

// normalizeWFO accepts inputs like "mrx", "MRX", "kmrx" and returns "MRX".
func normalizeWFO(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	if strings.HasPrefix(s, "K") && len(s) == 4 {
		s = s[1:]
	}
	return s
}

// normalizeStation accepts "krdu" or "KRDU" and returns "KRDU".
func normalizeStation(s string) string {
	return strings.TrimSpace(strings.ToUpper(s))
}
