package config

import (
	"fmt"
	"os"
	"reflect"
)

// WriteToFile writes the configuration settings to a file in a key-value format.
func (c *Config) WriteToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	// iterate all fields of the config struct and write them to the file
	v := reflect.ValueOf(c).Elem()

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Type().Name() == "Location" {
			continue
		}
		s := v.Field(i).Interface().(Setting)
		if !s.Secret {
			_, err = fmt.Fprintf(file, "%s=\"%s\"\n", s.Env, s.Value)
		}
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", filename, err)
		}
	}

	return nil
}

// WriteToStdout writes the configuration settings to standard output in a key-value format.
func (c *Config) WriteToStdout() error {
	// iterate all fields of the config struct and write them to stdout
	v := reflect.ValueOf(c).Elem()

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Type().Name() == "Location" {
			continue
		}
		s := v.Field(i).Interface().(Setting)
		if !s.Secret {
			fmt.Printf("%s=\"%s\"\n", s.Env, s.Value)
		}
	}

	return nil
}

// writeSecretToFile writes a secret setting to a file named after the environment variable.
func writeSecretToFile(secret Setting, filename string) error {
	if !secret.Secret {
		return fmt.Errorf("setting %s is not a secret", secret.Env)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	_, err = fmt.Fprint(file, secret.Value)
	if err != nil {
		return fmt.Errorf("failed to write secret to file %s: %w", filename, err)
	}

	return nil
}

// WriteSecrets writes secret settings to files.
func (c *Config) WriteSecrets(prefix string) error {
	if prefix == "" {
		prefix = "."
	}

	v := reflect.ValueOf(c).Elem()

	for i := 0; i < v.NumField(); i++ {
		s := v.Field(i).Interface().(Setting)
		if s.Secret {
			err := writeSecretToFile(s, prefix+"/"+s.SecretFile)
			if err != nil {
				return fmt.Errorf("failed to write secret %s: %w", s.Env, err)
			}
		}
	}
	return nil
}
