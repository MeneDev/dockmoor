package main

type ExitCode int

const (
	ExitSuccess ExitCode = iota
	ExitInvalidParams
	_ // reserved for ExitUnknownError
	ExitNotFound
	ExitInvalidFormat
	ExitCouldNotOpenFile
)
