package pkg

import "fmt"

func HandleErrorFormat(desc string, err error) error {
	return fmt.Errorf("\n%s\nError Details: %w", desc, err)
}
