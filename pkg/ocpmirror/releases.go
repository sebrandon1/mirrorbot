package ocpmirror

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
)

const (
	MirrorsBaseURL = "https://mirror.openshift.com/pub/openshift-v4/clients/"
)

// ReleaseInfo holds information about a found OCP release.
type ReleaseInfo struct {
	Version string
	Folder  string // "ocp" or "ocp-dev-preview"
	URL     string
}

// ListReleases fetches and parses available releases for a given major.minor version (e.g., "4.20")
func ListReleases(version string) ([]ReleaseInfo, error) {
	folders := []string{"ocp", "ocp-dev-preview"}
	var releases []ReleaseInfo
	for _, folder := range folders {
		url := MirrorsBaseURL + folder + "/"
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				fmt.Printf("Error closing response body: %v\n", err)
			}
		}()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		// Find all subfolders matching the version (e.g., 4.20.0-0.nightly-*)
		// This regex will match folder names like 4.20.0-0.nightly-2024-05-20-123456
		re := regexp.MustCompile(fmt.Sprintf(`href="(%s[^"]+)/"`, regexp.QuoteMeta(version)))
		matches := re.FindAllStringSubmatch(string(body), -1)
		for _, m := range matches {
			ver := m[1]
			releases = append(releases, ReleaseInfo{
				Version: ver,
				Folder:  folder,
				URL:     url + ver + "/",
			})
		}
	}
	// Sort releases by version string descending (latest first)
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Version > releases[j].Version
	})
	return releases, nil
}
