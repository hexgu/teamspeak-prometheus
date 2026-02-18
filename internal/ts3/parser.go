package ts3

import (
	"strings"
)

// ParseResponse parses a TS3 ServerQuery response string into a list of maps.
func ParseResponse(response string) []map[string]string {
	if response == "" {
		return nil
	}

	var result []map[string]string

	// Items are separated by "|"
	items := strings.Split(response, "|")
	for _, itemStr := range items {
		itemMap := make(map[string]string)
		// Properties are separated by " "
		props := strings.Split(itemStr, " ")
		for _, prop := range props {
			if prop == "" {
				continue
			}
			// Key-value pairs separated by "="
			// Some keys might not have value? (e.g. flags)
			parts := strings.SplitN(prop, "=", 2)
			key := Unescape(parts[0])
			val := ""
			if len(parts) > 1 {
				val = Unescape(parts[1])
			}
			itemMap[key] = val
		}
		result = append(result, itemMap)
	}
	return result
}
