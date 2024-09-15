# Leopards Ate My Explanation

[![Go](https://github.com/NatoBoram/lame/actions/workflows/go.yml/badge.svg)](https://github.com/NatoBoram/lame/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/NatoBoram/lame)](https://goreportcard.com/report/github.com/NatoBoram/lame)

A LLM-powered verification tool for explanatory comments in [r/LeopardsAteMyFace](https://www.reddit.com/r/LeopardsAteMyFace).

## Installation

```sh
go install github.com/NatoBoram/lame@latest
```

## Usage

Run `lame` and it'll create configurations files at `~/.config/lame` that you'll be able to fill out.

You'll need to create a _personal use script_ at <https://ssl.reddit.com/prefs/apps> and put its information in `~/.config/lame/reddit_credentials.json` then create a _User API key_ at <https://platform.openai.com/settings/profile?tab=api-keys> and put it in `~/.config/lame/openai_credentials.json`.

Once the information is filled out, run `lame` again and it'll prompt for Reddit post links.

Currently, only links in the form of `https://www.reddit.com/r/LeopardsAteMyFace/comments/lt8zlq` are supported.

## Example

```log
Enter a Reddit post url: https://www.reddit.com/r/LeopardsAteMyFace/comments/1fdx25e/minnesota_restaurant_owner_confronts_don_jr/

Title: Minnesota Restaurant Owner Confronts Don Jr. Because He Lost Half His Customers Over Trump Support
Body:
URL: https://www.yahoo.com/news/don-jr-confronted-podcast-restaurant-222601407.html

Found comment by u/AutoModerator
Found explanatory comment by u/Humble_Novice
Body: 1. Restaurant owner Allen Brenycz had purchased a digital billboard expressing his support for Donald Trump.
2. But because of this billboard, he ended up losing over half of his customers.
3. The restaurant owner then confronts Don Jr. on his podcast out of frustration over the whole thing.

Recommendation: Approve
Explanation: Allen Brenycz voted for, supported or wanted to impose support for Donald Trump on other people. Support for Donald Trump has the consequences of losing over half of his customers. As a consequence of supporting Donald Trump, losing over half of his customers happened to Allen Brenycz.
```
