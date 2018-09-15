# prizmlang
[![Go Report Card](https://goreportcard.com/badge/github.com/comp500/prizmlang)](https://goreportcard.com/report/github.com/comp500/prizmlang) [![Build Status](https://travis-ci.org/comp500/prizmlang.svg?branch=master)](https://travis-ci.org/comp500/prizmlang)

A tool to convert CASIO Prizm g3l language files to and from JSON for editing.

## Usage
1. Download the tool from [the releases section](https://github.com/comp500/prizmlang/releases) and open a terminal (command prompt) in the folder where you downloaded it
1. Download an existing language file from CASIO's website or [Cemetech](https://www.cemetech.net/programs/index.php?mode=file&id=1434)
1. Decode the language file: `prizmlang decode English.g3l edit.json` (replace English.g3l with the file you downloaded)
1. Edit the created "edit.json" file with your translations
1. Ensure the FileName in edit.json is the name you want for the file
1. Re-encode the language file: `prizmlang encode edit.json NewLang.g3l` (replace NewLang.g3l with the name you want)
1. Copy the language file onto your calculator

Once you have copied it, it should be accessible in the language selection menu.

## Compile
If you are paranoid, you can compile it yourself.

1. `git clone https://github.com/comp500/prizmlang.git`
1. `go build`
