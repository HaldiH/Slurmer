package utils

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

func FirstLine(r io.Reader) string {
	bio := bufio.NewReader(r)
	line, _ := bio.ReadBytes('\n')
	return strings.TrimRight(string(line), "\n")
}

func MapToJSONArray[K comparable, V any](m map[K]V) ([]byte, error) {
	jsonData := []byte{'['}

	first := true
	for _, val := range m {
		if first {
			first = false
		} else {
			jsonData = append(jsonData, ',')
		}
		jobJsonData, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, jobJsonData...)
	}

	jsonData = append(jsonData, ']')

	return jsonData, nil
}
