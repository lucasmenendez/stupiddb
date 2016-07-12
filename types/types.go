package types

import (
	"fmt"
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

	switch data.Alias {
		case "string":
			data.Content, err = encoder.String(data.Content.(string), data.Size)
		case "float":
			data.Content, err = encoder.Float(data.Content.(float64))
		case "int":
			data.Content, err = encoder.Int(data.Content.(int))
		case "bool":
			data.Content, err = encoder.Bool(data.Content.(bool))
		default:
			err = DBError{"Unknown data type."}
	}

	return err
}

func (data *Type) Decoder() error {
	var err error

	switch data.Alias {
		case "string":
			data.Content, err = decoder.String(data.Content)
		case "float":
			data.Content, err = decoder.Int(data.Content)
		case "int":
			data.Content, err = decoder.Float(data.Content)
		case "bool":
			data.Content, err = decoder.Bool(data.Content)
		default:
			err = DBError{"Unknown data type."}
	}

	return err
}
