package ocpmirror

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/sebrandon1/go-toolbox/httpx"
)

const (
	MirrorsBaseURL = "https://mirror.openshift.com/pub/openshift-v4/clients/"
)

// httpClient is shared by all OCP mirror API calls. It bounds request duration,
// caps response sizes, and turns non-2xx responses into typed errors.
var httpClient = httpx.NewClient(httpx.Options{
	Timeout: 30 * time.Second,
})

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
		req, err := httpClient.NewRequest(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
		}
		body, err := httpx.ReadLimited(resp.Body, httpClient.MaxResponseBytes)
		closeErr := resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		if closeErr != nil {
			fmt.Printf("Error closing response body: %v\n", closeErr)
		}
		if err := httpx.CheckStatus(resp, body); err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
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
