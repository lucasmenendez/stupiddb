package encoder

import (
	"fmt"
	"math"
	"bytes"
	"strconv"

	"stupiddb/dberror"
)

func String(data string, size int) (interface{}, error) {
	var length		= len(data)
	var offset int	= (size - length)
	if offset < 0 {
		return []byte{}, dberror.DBError{"Data exceed column size limit."}
	}

	res := make([]byte, offset)
	for i := 0; i < offset; i++ {
		res[i] = 32
	}

	bff := bytes.NewBuffer(res)
	if writted, err := bff.Write([]byte(data)); err != nil || writted != length {
		return []byte{}, dberror.DBError{"Error on string encoding."}
	}

	return bff.Bytes(), nil
}

func Float(data float64) (interface{}, error) {
	var content []byte	= []byte(fmt.Sprint(math.Float64bits(data)))
	var length int		= len(content)
	var offset int		= 20 - length

	if offset < 0 {
		return []byte{}, dberror.DBError{"Data exceed float size."}
	}

	res := make([]byte, offset)
	for i := 0; i < offset; i++ {
		res[i] = 32
	}

	bff := bytes.NewBuffer(res)
	if writted, err := bff.Write(content); err != nil || writted != length {
		return []byte{}, dberror.DBError{"Error on float encoding"}
	}

	return bff.Bytes(), nil
}

func Int(data int64) (interface{}, error) {
	var content []byte	= []byte(strconv.Itoa(int(data)))
	var length int		= len(content)
	var offset int		= 4 - length

	if offset < 0 {
		return []byte{}, dberror.DBError{"Data exceed int size."}
	} else if offset == 0 {
		return content, nil
	}

	res := make([]byte, offset)
	for i := 0; i < offset; i++ {
		res[i] = 32
	}

	bff := bytes.NewBuffer(res)
	if writted, err := bff.Write(content); err != nil || writted != length {
		return []byte{}, dberror.DBError{"Error on int encoding."}
	}

	return bff.Bytes(), nil
}

func Bool(data bool) (interface{}, error) {
	var result []byte

	if data {
		result = []byte("1")
	} else {
		result = []byte("0")
	}

	return result, nil
}
