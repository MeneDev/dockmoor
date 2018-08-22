# dockfix [![CircleCI](https://circleci.com/gh/MeneDev/dockfix.svg?style=svg)](https://circleci.com/gh/MeneDev/dockfix) [![Coverage Status](https://coveralls.io/repos/github/MeneDev/dockfix/badge.svg)](https://coveralls.io/github/MeneDev/dockfix)
dockfix Version not a release

Manage docker image references.

## Usage
> dockfix \[OPTIONS\] &lt;[find](#find-command) | [update](#update-command)&gt; \[command-OPTIONS\]

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
 * [update](#update-command)

## find command
> dockfix \[OPTIONS\] find \[find-OPTIONS\] InputFile

The find command returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions, non-null otherwise

### Predicates
Specify which kind of image references should be selected. Exactly one must be specified

**--any**  
Find all images

**--latest**  
Using latest tag

**--unpinned**  
Using unpinned images

**--outdated**  
Find all images with newer versions available

### Filters
Optional additional filters. Specifying each kind of filter must be matched at least once

**--name**  
Find all images matching one of the specified names

**--domain**  
Find all images matching one of the specified domains

### Help Options
**-h**, **--help**  
Show this help message

## update command
> dockfix \[OPTIONS\] update \[update-OPTIONS\] Filename

Replace image references with a latest reference from repository

**--version**  
The version to update to (one of `latest`, `tag`, `major`, `minor`, `patch`)

### Help Options
**-h**, **--help**  
Show this help message

