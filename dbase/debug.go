package dbase

import (
	"io"
	"log"
	"os"
)

var (
	debug       = false
	debugLogger = log.New(os.Stdout, "[dbase] [DEBUG] ", log.LstdFlags)
	errorLogger = log.New(os.Stdout, "[dbase] [ERROR] ", log.LstdFlags)
)

// Debug enables or disables debug logging for the dbase package.
// If debug is true, debug messages will be printed to the specified io.Writer (default: os.Stdout).
// If out is nil, the output destination is not changed.
func Debug(enabled bool, out io.Writer) {
	if out != nil {
		debugLogger.SetOutput(out)
		errorLogger.SetOutput(out)
	}
	debug = enabled
}

func debugf(format string, v ...interface{}) {
	if debug {
		debugLogger.Printf(format, v...)
	}
}
