package utils

import (
	"bufio"
	"io"
	"strings"
)

func FirstLine(r io.Reader) string {
	bio := bufio.NewReader(r)
	line, _ := bio.ReadBytes('\n')
	return strings.TrimRight(string(line), "\n")
}
