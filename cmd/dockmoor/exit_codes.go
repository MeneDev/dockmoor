package main

type ExitCode int

const (
	ExitSuccess ExitCode = iota
	ExitInvalidParams
	ExitUnknownError
	ExitNotFound
	ExitInvalidFormat
	ExitCouldNotOpenFile
)
