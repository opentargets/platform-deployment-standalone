package tools

import (
	"log"
	"os"
	"os/exec"
)

// ExtractArchive extracts a tar.gz archive to the specified destination directory.
func ExtractArchive(archive, dest string) error {

	if _, err := os.Stat(dest); err == nil {
		log.Printf("directory %s already exists, skipping extraction\n", dest)
		return nil
	}

	err := os.MkdirAll(dest, 0755)
	if err != nil {
		log.Fatalf("failed to create directory %s: %v", dest, err)
	}

	cmd := exec.Command("tar", "--use-compress-program=pigz", "-xf", archive, "-C", dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
