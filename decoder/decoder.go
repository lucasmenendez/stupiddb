package decoder

import (
	"fmt"
	"math"
	"strconv"
)

type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}

func String(data []byte) (string, error) {
	return string(data), nil
}

func Float(data []byte) (float64, error) {
	res, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0.0, DBError{"Error on float decoding."}
	}
	return res, nil
}

func Int(data []byte) (int, error) {
	var res int
	var err error

	if res, err = strconv.Atoi(string(data)); err != nil {
		return 0, DBError{"Error on float decoding."}
	}
	return res, nil
}

func Bool(data []byte) (bool, erro) {
	var res bool = false

	if data[0] == 1 {
		res = true
	}

	return res, nil
}
