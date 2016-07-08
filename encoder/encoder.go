package encoder

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

func Encoder(data *Type) ([]byte, bool) {
	var result []byte
	var err bool = false

	switch data.Alias {
		case "string":
			result, err = stringEncoder(data)
		case "float":
			result, err = floatEncoder(data)
		case "int":
			result, err = intEncoder(data)
		case "bool":
			result, err = boolEncoder(data)
		default:
			err = true
	}

	return result, err
}

func stringEncoder(data *Type) ([]byte, bool) {
	l := len(data.Content)
	if l >= n {
		return make([]byte, n), true
	}

	res := make([]byte, data.Size-l)
	bff := bytes.NewBuffer(res)
	if _, err := bff.Write([]byte(data.Content)); err != nil {
		fmt.Println(err)
		return make([]byte, data.Size), true
	}

	return bff.Bytes(), false
}

func floatEncoder(data *Type) ([]byte, bool) {
	uint := math.Float64bits(data.Content)
	return []byte(fmt.Sprint(uint)), false
}

func intEncoder(data *Type) ([]byte, bool) {
	return []byte(strconv.Itoa(data.Content)), false
}

func boolEncoder(data *Type) ([]byte, bool) {
	var result []byte
	var err bool = false

	if data.Content == true {
		result = []byte{1}
	} else if data.Content == false {
		result = []byte{0}
	} else {
		err = true
	}

	return result, err
}
