// Package config provides functionality for managing deployment configurations.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/opentargets/platform-deployment-standalone/internal/tools"
)

// DeploymentConfig is an interface that represents the configuration for a deployment.
type DeploymentConfig interface {
	// GetDeploymentDir returns the directory where the deployment files are stored.
	GetDeploymentDir() string
	// Validate validates all settings in the deployment configuration.
	Validate() error
	// ReplaceFromEnv replaces the values of the deployment configuration from environment variables.
	ReplaceFromEnv()
	// ToString returns a string representation of the deployment configuration.
	ToString() string
	// GetSecretFields returns a slice of Settings that are secrets.
	GetSecretFields() []Setting
}

// Setting represents a configuration setting.
type Setting struct {
	Title          string
	Description    string
	Env            string
	Value          string
	Secret         bool
	SecretFilename string
	Validator      func(value string) error
	ValidatedValue string
}

// Validate checks the value of the Setting using the provided validator function.
func (s *Setting) Validate() error {
	if s.Validator == nil {
		return nil
	}
	err := s.ValidateWithSpinner()(s.Value)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", strings.ToLower(s.Title), err)
	}
	return nil
}

// ReplaceFromEnv sets the value of the Setting from the environment.
func (s *Setting) ReplaceFromEnv() {
	if newValue, exists := os.LookupEnv(s.Env); exists {
		s.Value = newValue
	}
}

// ToString returns a string representation of the Setting.
func (s *Setting) ToString() string {
	if s.Secret {
		return fmt.Sprintf("# %s is a secret located at ./%s\n", s.Env, s.SecretFilename)
	}
	return fmt.Sprintf("%s=\"%s\"\n", s.Env, s.Value)
}

// ValidateWithSpinner returns a validation function that uses a spinner to indicate progress.
func (s *Setting) ValidateWithSpinner() func(v string) error {
	return func(v string) error {
		if s.ValidatedValue == v || s.Validator == nil {
			return nil
		}
		var err error
		a := func() {
			err = s.Validator(v)
			if err == nil {
				s.ValidatedValue = v
			}
		}
		tools.RunWithSpinner(fmt.Sprintf("checking %s", strings.ToLower(s.Title)), a)
		if err != nil {
			return fmt.Errorf("invalid %s: %w", strings.ToLower(s.Title), err)
		}
		return nil
	}
}

// Input creates a new input for the Setting using the huh package.
func (s *Setting) Input() *huh.Input {
	return huh.NewInput().
		Title(s.Title).
		Description(s.Description).
		Value(&s.Value).
		Validate(s.ValidateWithSpinner())
}
