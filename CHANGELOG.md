# Changelog

## v0.0.5
#### New predicates
* domain
* name
* familiar-name
* path
* tag
* untagged
* digest
#### Removed predicates
* any (no implicit). Potentially breaking.

## v0.0.4

### New Commands

**list**: The list command behaves like the contains command but also prints matching image refernces.

## v0.0.3

### contains command
#### New predicate for contains command

**unpinned**: Match image references that are not pinned

## v0.0.2

### contains command
#### New predicate for contains command

**latest**: Match image references that are not pinned and untagged or tagged with "latest"

## v0.0.1

Initial release.

### New Commands

**contains**: The contains command returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions, non-null otherwise

### New Formats

**Dockerfile** The Format used by `docker build`
