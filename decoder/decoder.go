package decoder

import (
	"math"
	"strconv"

	"stupiddb/dberror"
)

//Delete offset spaces of []bytes provided to return clean result.
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

//Return casted string with 'trim()' from data provided.
//If something fails returns a 'DBError' with info message.
func String(data []byte) (interface{}, error) {
	return string(trim(data)), nil
}

//After check input data format return casted float from data provided.
//If something fails returns a 'DBError' with info message.
func Float(data []byte) (interface{}, error) {
	if len(data) > 20 {
		return 0.0, dberror.DBError{"Bad float data provided."}
	}

	var (
		encoded	[]byte
		u_int	uint64
		err		error
	)
	encoded = trim(data)
	u_int, err = strconv.ParseUint(string(encoded), 10, 64)
	if err != nil {
		return 0.0, dberror.DBError{"Error on float decoding."}
	}
	return math.Float64frombits(u_int), nil
}

//After check input data format return casted int from data provided.
//If something fails returns a 'DBError' with info message.
func Int(data []byte) (interface{}, error) {
	if len(data) > 4 {
		return 0, dberror.DBError{"Bad int data provided."}
	}

	var (
		encoded	[]byte
		res		int
		err		error
	)

	encoded = trim(data)
	if res, err = strconv.Atoi(string(encoded)); err != nil {
		return 0, dberror.DBError{"Error on float decoding."}
	}
	return res, nil
}

//After check input data format return casted bool from data provided.
//If something fails returns a 'DBError' with info message.
func Bool(data []byte) (interface{}, error) {
	if len(data) > 1 {
		return false, dberror.DBError{"Bad bool data provided."}
	}

	var res bool = false
	if data[0] == 1 {
		res = true
	}

	return res, nil
}
