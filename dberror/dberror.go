package dberror

import (
	"fmt"
)

//Describe custom DBError
type DBError struct {
	Message string
}

//Required error function
func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}
