package main

import (
	"os"

	"github.com/op/go-logging"
)

var Logger = logging.MustGetLogger("petrel")

func GetLogger(level logging.Level) *logging.Logger {
	fmt_string := "\r%{color}[%{time:06-01-02 15:04:05}][%{level:.6s}]%{color:reset} %{message}"
	format := logging.MustStringFormatter(fmt_string)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendleveled := logging.AddModuleLevel(backend)
	backendleveled.SetLevel(level, "")
	logging.SetFormatter(format)
	logging.SetBackend(backendleveled)

	return Logger
}
