package elfHttp

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var elfMode = debugMode
var ElfMode = DebugMode
var DefaultWriter io.Writer = os.Stdout

const (
	DebugMode = "debug"
	ReleaseMode = "release"
	TestMode = "test"
)

const (
	debugMode  = iota
	releaseMode
	testMode
)

func SetMode(value string) {
	switch value {
	case DebugMode:
		elfMode = debugMode
	case ReleaseMode:
		elfMode = releaseMode
	case TestMode:
		elfMode = testMode
	default:
		panic("elf mode unknown:" + value)
	}
	ElfMode = value
}

func IsDebugging() bool {
	return elfMode == debugMode
}

func debugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		if strings.HasPrefix(format, "\n") {
			format = strings.TrimPrefix(format, "\n")
			fmt.Fprintf(DefaultWriter, "\n")
		}
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(DefaultWriter, "[ELF] "+format, values...)
	}
}
