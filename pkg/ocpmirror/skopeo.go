package ocpmirror

// import (
// 	"fmt"
// 	"os/exec"
// )

// ImageCheckResult holds the result of a Skopeo image existence check.
type ImageCheckResult struct {
	Image  string
	Exists bool
	Error  error
}

// CheckImagesExist checks if the given images exist in the remote registry using skopeo inspect.
// The images parameter should be a map of logical name to image reference (with version already interpolated).
// func CheckImagesExist(images map[string]string) map[string]ImageCheckResult {
// 	results := make(map[string]ImageCheckResult)
// 	for name, ref := range images {
// 		cmd := exec.Command("skopeo", "inspect", "--override-arch=amd64", "--override-os=linux", "--authfile=/auth.json", fmt.Sprintf("docker://%s", ref))
// 		output, err := cmd.CombinedOutput()
// 		if err != nil {
// 			fmt.Printf("[Skopeo] %s: NOT FOUND\nOutput: %s\nError: %v\n", ref, string(output), err)
// 			results[name] = ImageCheckResult{Image: ref, Exists: false, Error: err}
// 		} else {
// 			fmt.Printf("[Skopeo] %s: FOUND\nOutput: %s\n", ref, string(output))
// 			results[name] = ImageCheckResult{Image: ref, Exists: true, Error: nil}
// 		}
// 	}
// 	return results
// }
