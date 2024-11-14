# Leopards Ate My Explanation

[![Go](https://github.com/NatoBoram/lame/actions/workflows/go.yaml/badge.svg)](https://github.com/NatoBoram/lame/actions/workflows/go.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/NatoBoram/lame)](https://goreportcard.com/report/github.com/NatoBoram/lame) [![Go Reference](https://pkg.go.dev/badge/github.com/NatoBoram/lame.svg)](https://pkg.go.dev/github.com/NatoBoram/lame)

A LLM-powered verification tool for explanatory comments in [r/LeopardsAteMyFace](https://www.reddit.com/r/LeopardsAteMyFace).

> [!IMPORTANT]
>
> Large language models are biaised towards positivity. They are probably going to refuse to remove anything.

## Installation

```sh
go install github.com/NatoBoram/lame@latest
```

## Usage

Run `lame` and it'll create configurations files at `~/.config/lame` that you'll be able to fill out.

You'll need to create a _personal use script_ at <https://ssl.reddit.com/prefs/apps> and put its information in `~/.config/lame/reddit_credentials.json`

```json
{
	"ID": "",
	"Secret": "",
	"Username": "",
	"Password": "",
	"Guide": "lt8zlq"
}
```

To use OpenAI, create a _User API key_ at <https://platform.openai.com/settings/profile?tab=api-keys> and put it in `~/.config/lame/openai_credentials.json`.

To use Ollama, host it somewhere and fill the fields in `~/.config/lame/openai_credentials.json`.

```json
{
	"Token": "ollama",
	"BaseURL": "http://localhost:11434/v1",
	"Model": "llama3.1"
}
```

Once the information is filled out, run `lame` again and it'll prompt for Reddit post links.

Only links in the form of `https://www.reddit.com/r/LeopardsAteMyFace/comments/lt8zlq` are supported.

## Example

![image](https://github.com/user-attachments/assets/ecd6b2e8-2cc5-44a1-b867-8d1c38dba3c8)
