package pkg

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func CheckFileExists(filepath string) (string, error) {
	f, err := os.Open(filepath)
	// jattempt to open file, handle error, and defer close
	defer f.Close()
	if err != nil {
		return "", err
	}

	msg := fmt.Sprintf("File exists in %s filepath.", filepath)
	return msg, nil
}

func ReadTextFileContents(filepath string) (string, error) {
	f, err := os.Open(filepath)
	// Attempt to open file, handle error, and defer close
	defer f.Close()
	if err != nil {
		return "", err
	}

	// Scan first line as sample, handle error
	var sb strings.Builder
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", nil
	}

	return sb.String(), nil
}
