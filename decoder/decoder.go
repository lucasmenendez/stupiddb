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

func trim(data []byte) []byte {
	var trimmed int
	for _, value := range data {
		if value != 32 {
			break
		}
		trimmed += 1
	}
	return data[trimmed:]
}

func String(data []byte) (interface{}, error) {
	return string(trim(data)), nil
}

func Float(data []byte) (interface{}, error) {
	if len(data) > 20 {
		return 0.0, DBError{"Bad float data provided."}
	}

	var encoded []byte = trim(data)

	u_int, err := strconv.ParseUint(string(encoded), 10, 64)
	if err != nil {
		return 0.0, DBError{"Error on float decoding."}
	}
	return math.Float64frombits(u_int), nil
}

func Int(data []byte) (interface{}, error) {
	if len(data) > 4 {
		return 0, DBError{"Bad int data provided."}
	}

	var res int
	var err error

	var encoded []byte = trim(data)
	if res, err = strconv.Atoi(string(encoded)); err != nil {
		return 0, DBError{"Error on float decoding."}
	}
	return res, nil
}

func Bool(data []byte) (interface{}, error) {
	if len(data) > 1 {
		return false, DBError{"Bad bool data provided."}
	}

	var res bool = false
	if data[0] == 1 {
		res = true
	}

	return res, nil
}
