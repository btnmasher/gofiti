# gofiti
Github vandalizer written in Go. Inspired by [gelstudios/gitfiti](https://github.com/gelstudios/gitfiti)

## Usage

Gofiti is a command-line application, with the minimum required parameters, it can be piped to a file to generate a shellscript that will initialize and run the needed commits to stylize the github histogram.

##### Setup

Make a new repo on github, and run the command with the correct parameters.

##### Commandline Flags

- `--user={string}` - `required`
  - The github username to pull the correct date information from and to generate the correct remotes in the local git repo.

- `--repo={string}` - `required`
  - The name of the repo as it will appears on github.

- `--text={string}` - `required`
  - The message to render in the histogram.
    - Message size is limited to 51 pixels width.
  - Most basic keyboard-printable symbols have a pre-made bitmap representation that is non-monospaced.
    - Backslash escapes are required for the following symbols:
    - **Backslash**: `\\` **Double Quotes**: `\"`
    - TODO: Image bitmaps (there's some placeholders in the sourcecode if you want to tinker).

- `--ssh` - `optional`
  - Enables generating the script to use git over SSH instead of HTTPS for the push to github.

- `--debug` - `optional`
  - Turns on debug logging (This will screw up the generated shellscript.)

- `--dryrun` - `optional`
  - Disabled generated shellscript output, only shows the preview histogram (Good when used with `--debug` for.. well.. debugging!)

##### Example:

```BASH
$ gofiti --user=btnmasher --repo=vandalized --text="I <3 Golang!" > vandalize.sh
```

Output:

```BASH
#!/bin/bash
git init vandalized
cd vandalized
touch README.md
git add README.md

# ----- Histogram Preview -----

#    -@@@------@--@@-----@@@--@@--@----@@--@--@--@@@-@---
#    --@------@--@--@---@----@--@-@---@--@-@--@-@----@---
#    --@-----@------@---@----@--@-@---@--@-@@-@-@----@---
#    --@----@-----@@----@-@@-@--@-@---@@@@-@-@@-@-@@-@---
#    --@-----@------@---@--@-@--@-@---@--@-@--@-@--@-@---
#    --@------@--@--@---@--@-@--@-@---@--@-@--@-@--@-----
#    -@@@------@--@@-----@@@--@@--@@@-@--@-@--@--@@@-@---

#
#..... generated commit commands would follow here .....
#
```

Then

```BASH
$ bash vandalize.sh
```

## Todo:

- Lowercase letters
- Text images
- Load from bitmap file
- Output to specified file instead of stdout
- Direct action
  - Call git commands directly instead of generating a shellscript
  - Automatically push to github
  - Easy one-shot update of github histogram from existing

## Contribute

Send me PRs, you know what to do <3