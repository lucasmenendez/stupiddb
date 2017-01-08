package types

import (
	"reflect"

	"stupiddb/dberror"
	"stupiddb/encoder"
	"stupiddb/decoder"
)

//Type contains column attributes that defines data type, length and provided
//content access.
type Type struct {
	Alias		string
	Indexable	bool
	Size		int
	Content		interface{}
}

//Returns empty Int type struct.
func Int(indexable bool) Type {
	return Type{"int", indexable, 4, nil}
}

//Returns empty Float type struct.
func Float() Type {
	return Type{"float", false, 20, nil}
}

//Returns empty Bool type struct.
func Bool() Type {
	return Type{"bool", false, 1, nil}
}

//Returns empty String type struct with length provided.
func String(size int, indexable bool) Type {
	return Type{"string", indexable, size, nil}
}

//Returns if current type is empty
func (data *Type) Empty() bool {
	return data.Alias == "" && data.Content == nil
}

//Fill Type content with encoded typed representation data content according
//to associate data structure. If something was wrong returns a DBError,
//else nil.
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
			err = dberror.DBError{"Unknown data type."}
	}

	return err
}

//Fill Type content typed according with associated data structure. If
//something was wrong returns a DBError, else nil.
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
			err = dberror.DBError{"Unknown data type."}
	}

	return err
}
