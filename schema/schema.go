package schema

import (
	//"bytes"
	//"encoding/binary"
	"fmt"
	//"math"
)

//Handle error
func Error(err string) {
	fmt.Println("Error: %v", err)
	return
}

type DataType struct {
	alias     string
	constrain string
	size      int
}

func checkConstrain(constrain string) bool {
	return constrain == "primary" || constrain == "unique"
}

func Int(constrains []string) *DataType {
	var constrain string
	if len(constrains) > 0 {
		if checkConstrain(constrains[0]) {
			Error("Wrong constain.")
			return nil
		}
		constrain = constrains[0]
	}
	return &DataType{"int", constrain, 4}
}

func Float(constrains []string) *DataType {
	var constrain string
	if len(constrains) > 0 {
		if checkConstrain(constrains[0]) {
			Error("Wrong constain.")
			return nil
		}
		constrain = constrains[0]
	}
	return &DataType{"float", constrain, 20}
}

func Bool() *DataType {
	return &DataType{"bool", "", 1}
}

func String(size int, constrains []string) *DataType {
	var constrain string
	if len(constrains) > 0 {
		if checkConstrain(constrains[0]) {
			Error("Wrong constain.")
			return nil
		}
		constrain = constrains[0]
	}
	return &DataType{"string", constrain, size}
}

//Create table
func CreateTable(dblocation string, table string, fields map[string]*DataType) error {
	/*fd, err := os.Create(dblocation + table)
	if err != nil {
		return Error{"Error creating database file."}
	}
	defer fd.Close()
	*/
	var header string
	for column, datatype := range fields {
		header += fmt.Sprintf("%s(%d)%s;", datatype.alias, datatype.size, column)
	}
	fmt.Println("%d\n%s", len(header), header)

	/*	if _, err = fd.Write(); err != nil {
			return Error{"Error creating database struct."}
		}
	*/
	return nil
}
