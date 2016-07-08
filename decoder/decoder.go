package decoder

import (
	"fmt"
	"math"
	"strconv"
)

type Type interface {
	Alias		string
	Constrain	string
	Size		int
	Content		interface{}
}

func Decoder(data *Type) ({}interface, bool) {
	var result {}interface
	var err bool = false

	switch data.Alias {
		case "string":
			result, err = stringDecoder(data)
		case "float":
			result, err = floatDecoder(data)
		case "int":
			result, err = intDecoder(data)
		case "bool":
			result, err = boolDecoder(data)
		default:
			err = true
	}

	return result, err
}

func stringDecoder(data *Type) (string, bool) {
	return string(data.Content), false
}

func floatDecoder(data *Type) (float64, bool) {
	uint, err := strconv.ParseUint(string(data.Content), 10, 64)
	if err != nil {
		return 0.0, err
	} else {
		return math.Float64frombits(uint), nil
	}

}

func intDecoder(data *Type) (int, bool) {
	var res int
	var err error

	if res, err = strconv.Atoi(string(data.Content)); err != nil {
		return 0, true
	}
	return res, false
}

func boolDecoder(data *Type) (bool, bool) {
	var res bool = false

	if data.Content[0] == 1 {
		res = true
	}

	return res, false
}
