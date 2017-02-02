PACKAGE DOCUMENTATION

package dberror
    import "./dberror/"


TYPES

type DBError struct {
    Message string
}
    Describe custom DBError

func (err DBError) Error() string
    Required error function

