package logger

import (
	"github.com/op/go-logging"
	"os"
)

var Logger = logging.MustGetLogger("petrel")

func Get() *logging.Logger {
	fmt_string := "\r%{color}[%{time:06-01-02 15:04:05}][%{level:.6s}]%{color:reset} %{message}"
	format := logging.MustStringFormatter(fmt_string)
	logging.SetFormatter(format)
	logging.SetBackend(logging.NewLogBackend(os.Stdout, "", 0))

	return Logger
}
