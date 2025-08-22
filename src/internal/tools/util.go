package tools

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/charmbracelet/huh/spinner"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
)

// RandomString generates a random string of the specified length using the characters 'a' to 'h'.
func RandomString(length int) string {
	letters := []rune("abcdefgh")
	randomString := make([]rune, length)

	for i := range randomString {
		randomString[i] = letters[rand.Intn(len(letters))]
	}

	return string(randomString)
}

// Either returns the first non-empty string from either or.
func Either(either string, or string) string {
	if either != "" {
		return either
	}
	return or
}

// AppendIfErr appends an error to a slice if it is not nil.
func AppendIfErr(errs *[]error, err error) {
	if err != nil {
		*errs = append(*errs, err)
	}
}

// ListFilesInGCSPrefix lists files in a Google Cloud Storage bucket with the specified prefix.
func ListFilesInGCSPrefix(prefix string) ([]string, error) {
	ctx := context.Background()

	parts := strings.SplitN(strings.TrimPrefix(prefix, "gs://"), "/", 2)
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid gcs uri: %s", prefix)
	}

	bucketName := parts[0]
	var blobPrefix string
	if len(parts) == 2 {
		blobPrefix = parts[1]
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	it := client.Bucket(bucketName).Objects(ctx, &storage.Query{Prefix: blobPrefix})
	var files []string
	for {
		attr, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		basename := strings.TrimPrefix(attr.Name, blobPrefix+"/")
		files = append(files, basename)
	}

	return files, nil
}

// ReadFileFromGCS reads a file from Google Cloud Storage and returns its content as a string.
func ReadFileFromGCS(uri string) (string, error) {
	ctx := context.Background()

	parts := strings.SplitN(strings.TrimPrefix(uri, "gs://"), "/", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid gcs uri: %s", uri)
	}

	bucketName := parts[0]
	blobName := parts[1]

	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	reader, err := client.Bucket(bucketName).Object(blobName).NewReader(ctx)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	return string(content), err
}

// WriteFileToGCS writes a string to a file in Google Cloud Storage.
func WriteFileToGCS(uri string, content string) error {
	ctx := context.Background()

	parts := strings.SplitN(strings.TrimPrefix(uri, "gs://"), "/", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid gcs uri: %s", uri)
	}

	bucketName := parts[0]
	blobName := parts[1]

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	writer := client.Bucket(bucketName).Object(blobName).NewWriter(ctx)
	defer writer.Close()

	_, err = writer.Write([]byte(content))
	return err

}

// LoadEnvFromFile reads environment variables from a specified file and returns them as a map.
func LoadEnvFromFile(configFilePath string) (map[string]string, error) {
	if strings.HasPrefix(configFilePath, "gs://") {
		config, err := ReadFileFromGCS(configFilePath)
		if err != nil {
			return nil, err
		}

		return godotenv.Unmarshal(config)
	}

	return godotenv.Read(configFilePath)
}

// RunWithSpinner runs a function with a spinner and a title.
func RunWithSpinner(title string, action func()) error {
	spinnerType := spinner.Points
	return spinner.New().Type(spinnerType).Title(" " + title).Action(action).Run()
}
