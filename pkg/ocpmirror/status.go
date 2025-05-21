package ocpmirror

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ReleaseStatus holds the status info for a release from the status API.
type ReleaseStatus struct {
	Phase             string `json:"phase"`
	Age               string `json:"age"`
	Created           string // from changeLogJson.to.created
	KubernetesVersion string // from components
	RHCOSVersion      string // from components
	RHCOSFrom         string // from components (from)
}

type ReleaseStatusDetail struct {
	PullSpec string `json:"pullSpec"`
}

// FetchReleaseStatus fetches the phase, age, and created date for a given release version (e.g., "4.20.0-ec.0")
func FetchReleaseStatus(version string) (*ReleaseStatus, error) {
	paths := []string{"4-dev-preview", "4-stable"}
	for _, path := range paths {
		url := fmt.Sprintf("https://openshift-release.apps.ci.l2s4.p1.openshiftapps.com/api/v1/releasestream/%s/release/%s", path, version)
		resp, err := http.Get(url)
		if err != nil {
			continue // try next path
		}
		defer resp.Body.Close()
		var data struct {
			Phase         string `json:"phase"`
			Age           string `json:"age"`
			ChangeLogJson struct {
				To struct {
					Created string `json:"created"`
				} `json:"to"`
				Components []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
					From    string `json:"from"`
				} `json:"components"`
			} `json:"changeLogJson"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			continue // try next path
		}
		var kubeVer, rhcosVer, rhcosFrom string
		for _, c := range data.ChangeLogJson.Components {
			if c.Name == "Kubernetes" {
				kubeVer = c.Version
			}
			if c.Name == "Red Hat Enterprise Linux CoreOS" {
				rhcosVer = c.Version
				rhcosFrom = c.From
			}
		}
		return &ReleaseStatus{
			Phase:             data.Phase,
			Age:               data.Age,
			Created:           data.ChangeLogJson.To.Created,
			KubernetesVersion: kubeVer,
			RHCOSVersion:      rhcosVer,
			RHCOSFrom:         rhcosFrom,
		}, nil
	}
	return nil, fmt.Errorf("release status not found in 4-dev-preview or 4-stable for %s", version)
}

// FetchReleaseDetail fetches the pullSpec for a given release version, trying both 4-dev-preview and 4-stable
func FetchReleaseDetail(version string) (*ReleaseStatusDetail, error) {
	paths := []string{"4-dev-preview", "4-stable"}
	for _, path := range paths {
		url := fmt.Sprintf("https://openshift-release.apps.ci.l2s4.p1.openshiftapps.com/api/v1/releasestream/%s/release/%s", path, version)
		resp, err := http.Get(url)
		if err != nil {
			continue // try next path
		}
		defer resp.Body.Close()
		var data struct {
			PullSpec string `json:"pullSpec"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			continue // try next path
		}
		return &ReleaseStatusDetail{PullSpec: data.PullSpec}, nil
	}
	return nil, fmt.Errorf("release detail not found in 4-dev-preview or 4-stable for %s", version)
}
