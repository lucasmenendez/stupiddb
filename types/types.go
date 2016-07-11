package types

import (
	"fmt"
)

type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}

type Type struct {
	Alias		string
	Indexable	bool
	Size		int
	Content		interface{}
}

func Int(indexable bool) Type {
	return Type{"int", indexable, 4, nil}
}

func Float() Type {
	return Type{"float", false, 20, nil}
}

func Bool() Type {
	return Type{"bool", false, 1, nil}
}

func String(size int, indexable bool) Type {
	return Type{"string", indexable, size, nil}
}

func wrongConstrain(constrain string) bool {
	return constrain != "primary" && constrain != "unique"
}
