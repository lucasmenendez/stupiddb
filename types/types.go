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
	Constrain	string
	Size		int
	Content		interface{}
}

func Int(constrains ...string) Type {
	var constrain string
	if len(constrains) > 0 {
		if wrongConstrain(constrains[0]) {
			fmt.Println(DBError{"Wrong constrain."})
			return Type{}
		}
		constrain = constrains[0]
	}
	return Type{"int", constrain, 4, nil}
}

func Float(constrains ...string) Type {
	var constrain string
	if len(constrains) > 0 {
		if wrongConstrain(constrains[0]) {
			fmt.Println(DBError{"Wrong constrain."})
			return Type{}
		}
		constrain = constrains[0]
	}
	return Type{"float", constrain, 20, nil}
}

func Bool() Type {
	return Type{"bool", "", 1, nil}
}

func String(size int, constrains ...string) Type {
	var constrain string
	if len(constrains) > 0 {
		if wrongConstrain(constrains[0]) {
			fmt.Println(DBError{"Wrong constrain."})
			return Type{}
		}
		constrain = constrains[0]
	}
	return Type{"string", constrain, size, nil}
}

func wrongConstrain(constrain string) bool {
	return constrain != "primary" && constrain != "unique"
}
