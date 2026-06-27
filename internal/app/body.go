package app

import (
	"io"
	"os"
	"strings"
)

func ReadBody(input string, stdin io.Reader) (string, error) {
	if input == "" {
		return "", NewError("Missing body input", "Pass --body <text>, --body @path, or --body -.")
	}
	if input == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", WrapError("Could not read body from stdin", "Check the piped input and try again.", err)
		}
		return string(data), nil
	}
	if strings.HasPrefix(input, "@") {
		path := strings.TrimPrefix(input, "@")
		data, err := os.ReadFile(path)
		if err != nil {
			return "", WrapError("Could not read body file", "Check the path after @ and try again.", err)
		}
		return string(data), nil
	}
	return input, nil
}
