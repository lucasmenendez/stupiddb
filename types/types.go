package types

import (
	"fmt"
)


type Type struct {
	Alias     string
	Constrain string
	Size      int
}

func Int(constrains ...string) Type {
	var constrain string
	if len(constrains) > 0 {
		if wrongConstrain(constrains[0]) {
			fmt.Println("Wrong constrain.")
			return Type{"", "", 0}
		}
		constrain = constrains[0]
	}
	return Type{"int", constrain, 4}
}

func Float(constrains ...string) Type {
	var constrain string
	if len(constrains) > 0 {
		if wrongConstrain(constrains[0]) {
			fmt.Println("Wrong constrain.")
			return Type{"", "", 0}
		}
		constrain = constrains[0]
	}
	return Type{"float", constrain, 20}
}

func Bool() Type {
	return Type{"bool", "", 1}
}

func String(size int, constrains ...string) Type {
	var constrain string
	if len(constrains) > 0 {
		if wrongConstrain(constrains[0]) {
			fmt.Println("Wrong constrain.")
			return Type{"", "", 0}
		}
		constrain = constrains[0]
	}
	return Type{"string", constrain, size}
}

func wrongConstrain(constrain string) bool {
	return constrain != "primary" && constrain != "unique"
}
