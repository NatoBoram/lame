package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func UnmarshalOpenAiCredentials(data []byte) (OpenAiCredentials, error) {
	var r OpenAiCredentials
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *OpenAiCredentials) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type OpenAiCredentials struct {
	Token string `json:"Token"`
}

func readOpenAiCredentials(configDir string) (OpenAiCredentials, error) {
	credsPath := openAiCredentialsPath(configDir)

	file, err := os.Open(credsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return OpenAiCredentials{}, fmt.Errorf("failed to open OpenAI credential file: %w", err)
		}

		err = createOpenaiCredentials(credsPath)
		if err != nil {
			return OpenAiCredentials{}, fmt.Errorf("failed to create OpenAI credentials: %w", err)
		}

		return OpenAiCredentials{}, fmt.Errorf("OpenAI credential file does not exist, created one at %s", credsPath)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return OpenAiCredentials{}, fmt.Errorf("failed to read OpenAI credential file: %w", err)
	}

	creds, err := UnmarshalOpenAiCredentials(data)
	if err != nil {
		return creds, fmt.Errorf("failed to unmarshal OpenAI credentials: %w", err)
	}

	err = verifyOpenAiCredentials(creds)
	return creds, err
}

func openAiCredentialsPath(configDir string) string {
	return filepath.Join(configDir, "openai_credentials.json")
}

func verifyOpenAiCredentials(creds OpenAiCredentials) error {
	if creds.Token == "" {
		return fmt.Errorf("OpenAI token is empty")
	}

	return nil
}

func createOpenaiCredentials(credsPath string) error {
	creds := OpenAiCredentials{
		Token: "",
	}

	bytes, err := creds.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal OpenAI credentials: %w", err)
	}

	file, err := os.Create(credsPath)
	if err != nil {
		return fmt.Errorf("failed to create OpenAI credential file: %w", err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("failed to write OpenAI credentials to file: %w", err)
	}

	return err
}
