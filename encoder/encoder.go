package encoder

import (
	"fmt"
	"math"
	"bytes"
	"strconv"
)

type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}

func String(data string, size int) (interface{}, error) {
	l := len(data)
	if l >= size {
		return make([]byte, size), DBError{"Data exceed column size limit."}
	}

	res := make([]byte, size-l)
	bff := bytes.NewBuffer(res)
	if _, err := bff.Write([]byte(data)); err != nil {
		return make([]byte, size), DBError{"Error on string encoding."}
	}

	return bff.Bytes(), nil
}

func Float(data float64) (interface{}, error) {
	uint := math.Float64bits(data)
	return []byte(fmt.Sprint(uint)), nil
}

func Int(data int64) (interface{}, error) {
	l := len(data)
	if l > 4 {
		return make([]bytem 4), DBError{"Data exceed int size."}
	} else if l == 4 {
		return []byte(strconv.Itoa(int(data))), nil
	}

	res := make([]byte, 4-l)
	bff := bytes.NewBuffer(res)
	content := []byte(strconv.Itoa(int(data)))

	if _, err := bff.Write(content); err != nil {
		return make([]byte, size), DBError{"Error on int encoding."}
	}

	return bff.Bytes(), nil
}

func Bool(data bool) (interface{}, error) {
	var result []byte

	if data {
		result = []byte{1}
	} else {
		result = []byte{0}
	}

	return result, nil
}
