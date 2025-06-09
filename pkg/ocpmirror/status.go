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
	SucceededJobs     int    // total succeeded jobs
	FailedJobs        int    // total failed jobs
}

// ReleaseStatusAPIResponse represents the JSON structure returned by the OCP release status API.
type ReleaseStatusAPIResponse struct {
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
	Results struct {
		InformingJobs map[string]struct {
			State string `json:"state"`
		} `json:"informingJobs"`
	} `json:"results"`
}

// ReleaseStatusDetail holds the detailed information for a release, such as the pullSpec.
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
		defer func() {
			_ = resp.Body.Close()
		}()
		var data ReleaseStatusAPIResponse
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
		// Count succeeded and failed jobs
		succeeded, failed := 0, 0
		for _, job := range data.Results.InformingJobs {
			switch job.State {
			case "Succeeded":
				succeeded++
			case "Failed":
				failed++
			}
		}
		return &ReleaseStatus{
			Phase:             data.Phase,
			Age:               data.Age,
			Created:           data.ChangeLogJson.To.Created,
			KubernetesVersion: kubeVer,
			RHCOSVersion:      rhcosVer,
			RHCOSFrom:         rhcosFrom,
			SucceededJobs:     succeeded,
			FailedJobs:        failed,
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
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				fmt.Printf("Error closing response body: %v\n", err)
			}
		}()
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
