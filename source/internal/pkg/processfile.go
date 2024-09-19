package pkg

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

func CheckTextFileExists(filepath string) (string, error) {
	f, err := os.Open(filepath)
	// jattempt to open file, handle error, and defer close
	defer f.Close()
	if err != nil {
		return "", err
	}

	msg := fmt.Sprintf("File exists in %s filepath.", filepath)
	return msg, nil
}

func CheckFileExists(filepath string) (string, error) {
	if _, err := os.Stat(filepath); err == nil {
		// Exists
		msg := fmt.Sprintf("File exists in %s filepath.", filepath)
		return msg, nil
	} else if errors.Is(err, fs.ErrNotExist) {
		// Does *not* exist
		return "", err
	} else {
		// File may or may not exist. See err for details
		return "", err
	}
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
