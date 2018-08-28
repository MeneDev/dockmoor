# dockfix [![CircleCI](https://circleci.com/gh/MeneDev/dockfix.svg?style=shield)](https://circleci.com/gh/MeneDev/dockfix) [![Coverage Status](https://coveralls.io/repos/github/MeneDev/dockfix/badge.svg)](https://coveralls.io/github/MeneDev/dockfix)
dockfix Version v0.0.1-rc12

Manage docker image references.

## Usage
> dockfix \[OPTIONS\] &lt;[find](#find-command)&gt; \[command-OPTIONS\]

## Application Options
**-l**, **--log-level**  
Sets the log-level (one of `NONE`, `ERROR`, `WARN`, `INFO`, `DEBUG`)

**--version**  
Show version and exit

**--manpage**  
Show man page and exit

**--markdown**  
Show usage as markdown and exit

## Help Options
**-h**, **--help**  
Show this help message

## Commands

 * [find](#find-command)

## find command
> dockfix \[OPTIONS\] find \[find-OPTIONS\] InputFile

The find command returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions, non-null otherwise

### Predicates
Specify which kind of image references should be selected. Exactly one must be specified

**--any**  
Find all images

### Filters
Optional additional filters. Specifying each kind of filter must be matched at least once

### Help Options
**-h**, **--help**  
Show this help message

