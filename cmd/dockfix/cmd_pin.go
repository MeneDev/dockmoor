package main
//
//import (
//	"github.com/jessevdk/go-flags"
//	"os"
//	"github.com/MeneDev/dockfix/docker_repo"
//	"bytes"
//	"github.com/MeneDev/dockfix"
//	"github.com/MeneDev/dockfix/dockfmt"
//)
//
//type PinOptions struct {
//	Style    string         `required:"no" long:"style" description:"How image references should be pinned" choice:"tag" choice:"major" choice:"minor" choice:"patch"`
//	NoDigest bool           `required:"no" long:"no-digest" description:"Only pin by tag, don't add digest"`
//	InPlace  bool           `required:"no" long:"in-place" description:"Write pinned file to input file. Overwrites original file."`
//	Output   flags.Filename `description:"Input file" default:"-"`
//	DryRun   bool           `description:"Test run, don't modify any files.'"`
//
//	Positional OutputFile `positional-args:"yes" required:"yes"`
//}
//
//var pinOptions PinOptions
//
//func init() {
//	parser.AddCommand("pin",
//		"Replace image references with a pinned reference",
//		"Replace image references with a pinned reference",
//		&pinOptions)
//}
//
//func (pinOptions PinOptions) Execute(args []string) error {
//	filePathInput := string(pinOptions.Positional.Filename)
//	fpInput, err := os.Open(filePathInput)
//	defer fpInput.Close()
//
//	if err != nil {
//		log.Errorf("Error opening file: %s", err)
//		os.Exit(EXIT_FILE_ERROR)
//	}
//
//	var filePathOutput string = ""
//	if pinOptions.InPlace && pinOptions.Output != "" {
//		log.Errorf("Use either --in-place or --output")
//		os.Exit(EXIT_INVALID_PARAMS)
//	}
//
//	if !pinOptions.DryRun {
//		if !pinOptions.InPlace && pinOptions.Output == "" {
//			log.Errorf("One of --in-place, --output or --dry-run is required")
//			os.Exit(EXIT_INVALID_PARAMS)
//		}
//
//	}
//	if pinOptions.InPlace {
//		filePathInput = filePathOutput
//	} else if pinOptions.Output != "" && pinOptions.Output != "-" {
//		filePathOutput = string(pinOptions.Output)
//		rlInput, e := os.Readlink(filePathInput)
//		if e != nil {
//			log.Errorf("Error reading link for %s: %s", filePathInput, err)
//			os.Exit(EXIT_FILE_ERROR)
//		}
//		rlOutput, e := os.Readlink(filePathOutput)
//		if e != nil {
//			log.Errorf("Error reading link for %s: %s", filePathOutput, err)
//			os.Exit(EXIT_FILE_ERROR)
//		}
//
//		if rlInput == rlOutput {
//			log.Errorf("Input and Output are the same file. Use --in-place instead.")
//			os.Exit(EXIT_FILE_ERROR)
//		}
//	} else if pinOptions.Output == "-" {
//		filePathOutput = string(pinOptions.Output)
//	}
//
//	repo, err := docker_repo.DockerRepoNew()
//	if err != nil {
//		log.Errorf("Error communicating with docker: %s", err)
//		os.Exit(EXIT_DOCKER_ERROR)
//	}
//
//	var buffer bytes.Buffer
//	err = dockfix.Pin(log, formatProvider, repo, fpInput, filePathInput, &buffer)
//	if err != nil {
//		log.Error(err)
//		if _, ok := err.(dockfmt.UnknownFormatError); ok {
//			os.Exit(EXIT_UNKNOWN_FORMAT)
//		} else {
//			os.Exit(EXIT_UNKNOWN_ERROR)
//		}
//	}
//
//	var fpOutput *os.File = nil
//
//	if filePathOutput == "-" {
//		fpOutput = os.Stdout
//	} else if filePathOutput != "" {
//
//		fpOutput, err = os.Open(filePathOutput)
//		defer fpInput.Close()
//
//		if err != nil {
//			log.Errorf("Error opening file: %s", err)
//			os.Exit(EXIT_FILE_ERROR)
//		}
//	} else if pinOptions.InPlace {
//		fpOutput = fpInput
//		fpOutput.Truncate(0)
//	}
//
//	if fpInput != nil {
//		buffer.WriteTo(fpOutput)
//	}
//
//	return nil
//}
