package types

import (
	"fmt"
	"reflect"
	"stupiddb/encoder"
	"stupiddb/decoder"
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

func (data *Type) Encoder() error {
	var err error
	var value reflect.Value = reflect.ValueOf(data.Content)

	switch data.Alias {
		case "string":
			data.Content, err = encoder.String(value.String(), data.Size)
		case "float":
			data.Content, err = encoder.Float(value.Float())
		case "int":
			data.Content, err = encoder.Int(value.Int())
		case "bool":
			data.Content, err = encoder.Bool(value.Bool())
		default:
			err = DBError{"Unknown data type."}
	}

	return err
}

func (data *Type) Decoder() error {
	var err error
	var value []byte = []byte(reflect.ValueOf(data.Content).String())

	switch data.Alias {
		case "string":
			data.Content, err = decoder.String(value)
		case "float":
			data.Content, err = decoder.Float(value)
		case "int":
			data.Content, err = decoder.Int(value)
		case "bool":
			data.Content, err = decoder.Bool(value)
		default:
			err = DBError{"Unknown data type."}
	}

	return err
}
