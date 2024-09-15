# Leopards Ate My Explanation

[![Go](https://github.com/NatoBoram/lame/actions/workflows/go.yaml/badge.svg)](https://github.com/NatoBoram/lame/actions/workflows/go.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/NatoBoram/lame)](https://goreportcard.com/report/github.com/NatoBoram/lame) [![Go Reference](https://pkg.go.dev/badge/github.com/NatoBoram/lame.svg)](https://pkg.go.dev/github.com/NatoBoram/lame)

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
Enter a Reddit post url: https://www.reddit.com/r/LeopardsAteMyFace/comments/1f81eyd/men_who_argued_that_anyone_involved_in_abortion/

Title: Men who argued that "anyone involved in abortion were sinners" ... and now in areas that banned abortions ... are realizing that they messed up when their wife's health is threatened and can't get abortion health care.
Body:
URL: https://www.washingtonpost.com/nation/2024/09/03/abortion-bans-pregnancy-miscarriage-men/

Found comment by u/AutoModerator
Found explanatory comment by u/Lighting
Body: 1. People who wanted to remove Medical Power of Attorney without due process (e.g. ban abortion) succeeded in many red states.

2. They succeeded and now they live in states where abortion is banned.

3. Now "red states" means the rivers of blood from the increased maternal mortality/morbidity from their policies and finding that death/disease/infertility is threatening the lives of women they know personally. Oops.

From the article:

&gt; Thomas Stovall grew up in a strict Baptist family in Mississippi and always believed that anyone involved with abortion was destined for hell.
&gt;
&gt; But his lifelong conviction crumbled when his wife, Chelsea, was 20 weeks pregnant with their third child. Tests showed a severely malformed and underdeveloped fetus, one that was sure to be stillborn if carried to term. There was other devastating news, too. Continuing with the pregnancy could threaten Chelsea’s health and future fertility, doctors warned.
&gt;
&gt; The couple live in Arkansas, which has a near-total ban on abortion and is surrounded by states with their own highly restrictive laws. So they drove 400 miles to reach a clinic in Illinois where they could end the pregnancy. As they did, Stovall says he’d decided he was “dead wrong about abortion being a sin.”

Recommendation: Approve
Someone: Thomas Stovall
Something: anti-abortion policies
Consequences: the increased maternal mortality/morbidity resulting from anti-abortion policies in red states
Explanation: **Thomas Stovall** voted for, supported or wanted to impose **anti-abortion policies** on **other people**. **Anti-abortion policies** have the consequences of **the increased maternal mortality/morbidity resulting from these policies**. As a consequence of **anti-abortion policies**, **the increased maternal mortality/morbidity resulting from these policies** happened to **Thomas Stovall**.

You can "approve", "remove" (no removal reason) or skip (default): a
Approved!
```
