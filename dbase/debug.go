package dbase

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
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
		debugLogger.Print(trace() + " - " + fmt.Sprintf(format, v...))
	}
}

func trace() string {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return ""
	}

	if f := runtime.FuncForPC(pc); f != nil {
		return fmt.Sprintf("%v:%v", f.Name(), line)
	}
	return fmt.Sprintf("%v:%v", file, line)
}
