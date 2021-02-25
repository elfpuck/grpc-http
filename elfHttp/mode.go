package elfHttp

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type ElfMode uint32

var elfMode = DebugMode
var DefaultWriter io.Writer = os.Stdout

const (
	DebugMode ElfMode = iota
	ReleaseMode
	TestMode
)

func SetMode(elMode ElfMode) {
	elfMode = elMode
}

func IsDebugging() bool {
	return elfMode == DebugMode
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
