package aiops

import "os"

// osReadFile is a thin wrapper around os.ReadFile so burst_detector.go
// can call readFileBytes without importing os directly in its body.
func osReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
