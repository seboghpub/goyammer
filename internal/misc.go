package internal

import (
	"fmt"
	"os"
)

func ElipseMe(s string, length int, pad bool) string {
	sLength := len(s)
	if sLength == length {
		return s
	}
	if sLength < length {
		if pad {
			return fmt.Sprintf("%-*.*s", length, length, s)
		}
		return s
	}
	return fmt.Sprintf("%-*.*s…", length-1, length-1, s)
}

func FileExists(path string) bool {
	_, errStat := os.Stat(path)
	if os.IsNotExist(errStat) {
		return false
	}
	return true
}
