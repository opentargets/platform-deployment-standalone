// Package tools provides utility functions.
package tools

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

// DownloadFile downloads a file from the specified URL to the given path.
func DownloadFile(url, path string) error {
	if _, err := os.Stat(path); err == nil {
		log.Printf("file %s already exists, skipping download\n", path)
		return nil
	}

	var cmd *exec.Cmd
	if strings.HasPrefix(url, "gs://") {
		cmd = exec.Command("gsutil", "cp", url, path)
	} else {
		cmd = exec.Command("wget", url, "-O", path)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
