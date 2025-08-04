package config

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// ValidateURL checks if the provided string is a valid URL.
func ValidateURL(s string) error {
	if s[:4] != "http" && s[:5] != "https" && s[:2] != "gs" {
		return errors.New("url must start with http(s):// or gs://")
	}
	if s[len(s)-1] == '/' {
		return errors.New("url must not end with a slash")
	}

	return nil
}

// ValidateImageExists checks if a docker image exists at the given URL.
func ValidateImageExists(s string) error {
	resp, err := http.Get(fmt.Sprintf("https://%s", s))
	if err != nil {
		return fmt.Errorf("failed to request image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("image not found, got status: %d", resp.StatusCode)
	}

	return nil
}

// ValidateSubdomain checks if the subdomain is valid and already exists.
func ValidateSubdomain(domainName *string) func(s string) error {
	return func(s string) error {
		currentDomainName := *domainName

		if s == "" {
			return errors.New("subdomain cannot be empty")
		}
		if len(s)+len(currentDomainName) > 255 {
			return errors.New("hostname too long (max 255 characters)")
		}

		validSubdomain := regexp.MustCompile(`^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$`)
		if !validSubdomain.MatchString(s) {
			return errors.New("invalid subdomain, only single level subdomains composed of lowercase letters, numbers, hyphens, and underscores are allowed")
		}

		fullDomain := fmt.Sprintf("http://%s.%s", s, currentDomainName)
		resp, err := http.Head(fullDomain)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode != http.StatusNotFound {
				return fmt.Errorf("host %s seems to already exist", fullDomain)
			}
		}
		return nil
	}
}

// ValidateDaysToLive checks if the number of days to live is valid.
func ValidateDaysToLive(subdomainName *string) func(s string) error {
	return func(s string) error {
		if s == "" {
			return errors.New("days to live cannot be empty")
		}

		d, err := strconv.ParseInt(s, 10, 64)
		if err != nil || d < 0 || d > CloudDeploymentMaxDaysToLive {
			return fmt.Errorf("days to live must be a number between 0 and %d", CloudDeploymentMaxDaysToLive)
		}

		currentSubdomainName := *subdomainName
		if s == "0" && strings.ToLower(currentSubdomainName) != "dev" {
			return errors.New("only 'dev' subdomain can live forever")
		}
		return nil
	}
}

// ValidateGCPResource checks if the GCP resource name is valid.
func ValidateGCPResource() func(s string) error {
	return func(s string) error {
		if s == "" || len(s) > 63 {
			return errors.New("gcp resource name must be between 1 and 63 characters long")
		}

		validChars := regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
		if !validChars.MatchString(s) {
			return errors.New("gcp resource name can only contain lowercase letters, numbers, and hyphens, and must start with a letter")
		}
		return nil
	}
}
