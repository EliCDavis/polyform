package edit

import (
	"fmt"
	"strings"
)

// pathSuffixAfterMarker returns the URL path following the first occurrence of marker.
// A leading slash on the suffix is stripped.
func pathSuffixAfterMarker(url, marker string) (string, error) {
	idx := strings.Index(url, marker)
	if idx == -1 {
		return "", fmt.Errorf("url should contain %q", marker)
	}

	suffix := url[idx+len(marker):]
	if len(suffix) > 0 && suffix[0] == '/' {
		suffix = suffix[1:]
	}
	return suffix, nil
}
