# Contributing

Thanks for your interest, you're very welcome to contribute. Don't take anything in here as a hard rule, and don't feel intimidated. I'll look at any PR you make and I'll try to answer any question you have!

## Prerequisites for building

* Setup a recent go, the CI currently uses Version 1.11.2
* Clone develop branch, make sure to clone to your `GOPATH`, e.g. `/home/mene/go/github.com/MeneDev/dockmoor`
```bash
cd $GOPATH
mkdir -p src/github.com/MeneDev
git clone https://github.com/MeneDev/dockmoor.git src/github.com/MeneDev/dockmoor
```
* Install [`dep` v0.5.0](https://github.com/golang/dep/releases/tag/v0.5.0)
  * If you use Homebrew you should be able to `brew install dep`
* Run `dep ensure` inside project folder


## Prerequisites for quality control
These are optional to get the project running, but are greatly appreciated before creating Pull Requests.

* gometalinter
  * Easy install by running `curl -L https://git.io/vp6lP | sh`
  * The CI currently uses v2.0.11, but newer should be fine.

## Working with the go cli tools

### Build dockmoor

    cd cmd/dockmoor
    go build
    
### Run

    cd cmd/dockmoor
    ./dockmoor
    
## Working with Visual Studio Code

Follow the official [Go in Visual Studio Code](https://code.visualstudio.com/docs/languages/go) Guide

The project contains an example `.vscode/launch.json` that should allow you to debug the project.

Note that you may need to change the arguments (also in `.vscode/launch.json`) to your needs.

Edit the run configuration to change the arguments that are passed to the program.

## Working with Goland

You can Run, Test and Debug out-of-the-box.
Right click on the `cmd/dockmoor` **folder** and choose the desired action.

## Communicate your ideas
There is a rough idea where the project is going in `ROADMAP.md`, but nothing's set in stone.
Whenever you want to take on any task, be it an issue that is already open,
something else from the Roadmap or a new idea, it's best to talk about it before you implement anything.

This just helps to avoid work that doesn't fit the project's scope or is better archived in a different manner.

If you have a use-case in mind, try to be as specific as you can. Examples help.

## Committing

Please run `gofmt -s -w .` before committing to fix simple code-style problems.

If you like include at least one emoji in the first line of your commit message.


| Emoji | Raw Emoji Code | Description |
|:---:|---|---|
| :art: | `:art:` | when improving the **format**/structure of the code |
| :books: | `:books:` | when writing **docs** |
| :ambulance: | `:ambulance:` | when fixing a **bug** |
| :fire: | `:fire:` | when **removing code** or files |
| :white_check_mark: | `:white_check_mark:` | when adding/fixing/improving **tests** |
| :construction_worker: | `:construction_worker:` | when modifying the **CI** build |
| :heavy_plus_sign: | `:heavy_plus_sign:` | when adding **dependencies** |
| :heavy_minus_sign: | `:heavy_minus_sign:` | when removing **dependencies** |
| :arrow_up: | `:arrow_up:` | when upgrading **dependencies** |
| :arrow_down: | `:arrow_down:` | when downgrading **dependencies** |
| :shirt: | `:shirt:` | when removing **linter**/strict/deprecation warnings |
| :construction: | `:construction:` | **WIP**(Work In Progress) Commits |
| :gem: | `:gem:` | New **Release** |
| :speaker: | `:speaker:` | when Adding **Logging** |
| :mute: | `:mute:` | when Reducing **Logging** |
| :sparkles: | `:sparkles:` | when introducing **New** Features |
| :zap: | `:zap:` | when introducing **Backward-Incompatible** Features |
| :octopus:  | `:octopus:` | **GIT** related stuff |
| :whale2:  | `:whale2:` | **docker** related stuff |

## Documenting

Documentation currently resides in `README.adoc` and is **generated** from `cmd/dockmoor/doc` and `cmd/dockmoor/end-to-end`.
So please **do not edit README.adoc in the root folder**.

Instead you can edit `cmd/dockmoor/doc/_readme.adoc` or one of the referenced files.

The file `cmd/dockmoor/doc/dockmoor.adoc` is also automatically generated.
It is the output of running `dockmoor --asciidoc`.
This will generate output similar to using the built-in `--help` flag from go-flats and should help to keep code and docs in-sync.

To generate the updated documentation, run `generate.sh` inside `cmd/dockmoor/doc`.

The documentations heavily references parts of the end-to-end tests in `cmd/docmoor/end-to-end/test.sh`.
This makes sure that claims made by the documentation are actually true.
Adding additional test-cases there and referencing it in the documentation is highly appreciated,
just make sure to change the used AsciiDoc-Tags to something unique.

## Pull Requests

Don't hesitate to create a pull request before satisfying all suggestions (I also do this). We can always look at the code together.

Please do not merge against the master branch, use develop instead.

Please run the `guess-quality.sh` script in the root folder and look at the output. Ideally there is none.

Please take a look at the test coverage.

Please make sure there are no merge conflicts.

## Releasing

Releases are created by tagging the develop branch, please ask when you think a release is due and I haven't step up to my duty.
