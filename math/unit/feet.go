package unit

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseFeet(dirtyFeetText string) (float64, error) {
	trimmed := strings.TrimSpace(dirtyFeetText)

	feetMarkerIndex := strings.Index(trimmed, "'")
	inchesMarkerIndex := strings.Index(trimmed, "\"")

	if feetMarkerIndex == -1 && inchesMarkerIndex == -1 {
		parsed, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return parsed, fmt.Errorf("no feet or inches markings, can't parse %s: %w", trimmed, err)
		}
		return parsed, nil
	}

	var feet float64
	var inches float64
	var err error
	if feetMarkerIndex > -1 {
		text := trimmed[:feetMarkerIndex]
		feet, err = strconv.ParseFloat(text, 64)
		if err != nil {
			return 0, fmt.Errorf("unable to parse feet %q: %w", text, err)
		}
	}

	if inchesMarkerIndex > -1 {
		start := 0
		if feetMarkerIndex != -1 {
			start = feetMarkerIndex + 1
		}

		text := strings.TrimSpace(trimmed[start:inchesMarkerIndex])
		inches, err = strconv.ParseFloat(text, 64)
		if err != nil {
			return 0, fmt.Errorf("unable to parse inches %q: %w", text, err)
		}
	}

	return feet + (inches / 12), nil
}
