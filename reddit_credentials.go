package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func UnmarshalRedditCredentials(data []byte) (RedditCredentials, error) {
	var r RedditCredentials
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *RedditCredentials) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type RedditCredentials struct {
	ID       string `json:"ID"`
	Secret   string `json:"Secret"`
	Username string `json:"Username"`
	Password string `json:"Password"`

	// Guide to this sub's explanatory comment rule.
	Guide string `json:"Guide"`
}

func readRedditCredentials(configDir string) (RedditCredentials, error) {
	credsPath := redditCredentialsPath(configDir)

	file, err := os.Open(credsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return RedditCredentials{}, fmt.Errorf("failed to open Reddit credential file: %w", err)
		}

		err = createRedditCredentials(credsPath)
		if err != nil {
			return RedditCredentials{}, fmt.Errorf("failed to create Reddit credentials: %w", err)
		}

		return RedditCredentials{}, fmt.Errorf("Reddit credential file does not exist, created one at %s", credsPath)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return RedditCredentials{}, fmt.Errorf("failed to read Reddit credential file: %w", err)
	}

	creds, err := UnmarshalRedditCredentials(data)
	if err != nil {
		return creds, fmt.Errorf("failed to unmarshal Reddit credentials: %w", err)
	}

	err = VerifyRedditCredentials(creds)
	return creds, err
}

func createRedditCredentials(credsPath string) error {
	creds := RedditCredentials{
		ID:       "",
		Secret:   "",
		Username: "",
		Password: "",
		Guide:    "lt8zlq",
	}

	data, err := creds.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal Reddit credentials: %w", err)
	}

	file, err := os.Create(credsPath)
	if err != nil {
		return fmt.Errorf("failed to create Reddit credentials file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write Reddit credentials file: %w", err)
	}

	return err
}

func VerifyRedditCredentials(creds RedditCredentials) error {
	var list []string

	if creds.ID == "" {
		list = append(list, "ID")
	}

	if creds.Secret == "" {
		list = append(list, "Secret")
	}

	if creds.Username == "" {
		list = append(list, "Username")
	}

	if creds.Password == "" {
		list = append(list, "Password")
	}

	if creds.Guide == "" {
		list = append(list, "Guide")
	}

	if len(list) > 0 {
		return fmt.Errorf("the following Reddit credentials are missing: %v", list)
	}

	return nil
}

func redditCredentialsPath(configDir string) string {
	return filepath.Join(configDir, "reddit_credentials.json")
}
