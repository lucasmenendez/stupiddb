package tables

import (
	"os"
	"fmt"
	"bufio"
	"regexp"
	"strconv"
	"stupiddb/types"
)

type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("DBError: %v", err.Message)
}

type Table struct {
	Header Header
	Name string
	LineSize int
	FileDescriptor *os.File
}

type Header struct {
	Size int
	Columns map[string]types.Type
}

func Use(name, location string, fd *os.File, sizes_rgx, header_rgx *regexp.Regexp) (*Table, error) {
	var err error

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return nil, DBError{"Bad formated table."}
	}

	var line_size, header_size int

	sizes_line := scanner.Text()
	res := sizes_rgx.FindStringSubmatch(sizes_line)
	if line_size, err = strconv.Atoi(res[1]); err != nil {
		return nil, err
	}
	if header_size, err = strconv.Atoi(res[2]); err != nil {
		return nil, err
	}

	if !scanner.Scan() {
		return nil, DBError{"Bad formated table."}
	}

	var columns map[string]types.Type = make(map[string]types.Type)
	var header_line string = scanner.Text()
	var header_items [][]string = header_rgx.FindAllStringSubmatch(header_line, -1)

	for _, column := range header_items {
		var column_size int
		if column_size, err = strconv.Atoi(column[2]); err != nil {
			return nil, err
		}
		columns[column[3]] = types.Type{column[1], "", column_size, nil}
	}

	return &Table{Header{header_size, columns}, name, line_size, fd}, nil
}

//func (db *engine) Add(query *Query) error {
//	fd, err := os.OpenFile(db.location+query.Table, os.O_RDWR|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		return DBError{"Table not found."}
//	}
//
//	scanner := bufio.NewScanner(fd)
//	if !scanner.Scan() {
//		return DBError{"Table not found."}
//	}
//
//	header := strings.Split(scanner.Text(), "<;;>")
//	headers := header[:len(header)-1]
//
//	defer fd.Close()
//
//	if len(headers) == len(query.Data) {
//		for scanner.Scan() {
//			fields := strings.Split(scanner.Text(), "<;;>")
//			for i, hd := range headers {
//				if hd[:2] == "U%" && query.Data[hd[2:]] == fields[i] {
//					return DBError{"Unique row '" + hd + "' restriction violated."}
//				}
//			}
//		}
//
//		var data string
//		for _, hd := range headers {
//			isField := false
//			for column, value := range query.Data {
//				if hd[:2] == "U%"c{
//					hd = hd[2:]
//				}
//
//				if column == hd {
//					data += value + "<;;>"
//					isField = true
//				} else {
//					isField = isField || false
//				}
//			}
//			if isField != true {
//				return DBError{"Table struct and data attributes doesn't match."}
//			}
//
//		}
//		if _, err = fd.Write([]byte(data + "\n")); err != nil {
//			return DBError{"Error writting table."}
//		}
//		return nil
//	}
//
//	return DBError{"Attribute length mismatch."}
//	return nil
//}
//
//func (db *engine) Edit(q *query.Query) error {
//	fd, err := os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		return DBError{"Table not found."}
//	}
//
//	scanner := bufio.NewScanner(fd)
//	if !scanner.Scan() {
//		return DBError{"Table not found."}
//	}
//
//	header := strings.Split(scanner.Text(), "<;;>")
//	headers := header[:len(header)-1]
//
//	filters := make(map[int]string)
//	for i, hd := range headers {
//		header := hd
//		if header[:2] == "U%" {
//			header = header[2:]
//		}
//
//		for filter, value := range q.f {
//			if header == filter {
//				filters[i] = value
//			}
//		}
//	}
//
//	var content []string
//	for scanner.Scan() {
//		content = strings.Split(scanner.Text(), "<;;>")
//		match := true
//		for pos, value := range filters {
//			match = value == content[pos]
//			if !match {
//				break
//			}
//		}
//		if match {
//			break
//		}
//	}
//
//	fd.Close()
//
//	old := strings.Join(content, "<;;>")
//	for i, hd := range headers {
//		for key, value := range q.d {
//			if hd[:2] == "U%" && hd[2:] == key {
//				return DBError{"Unique column cannot be modificated."}
//			}
//			if hd == key {
//				content[i] = value
//			}
//		}
//	}
//
//	new := []byte(strings.Join(content, "<;;>"))
//
//	fd, err = os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		return DBError{"Error opening table."}
//	}
//
//	fileinfo, _ := fd.Stat()
//	data := make([]byte, fileinfo.Size())
//	if _, err = fd.Read(data); err != nil {
//		return DBError{"Error reading table."}
//	}
//
//	data = bytes.Replace(data, []byte(old), new, -1)
//	if err = fd.Truncate(0); err != nil {
//		return DBError{"Error truncating table."}
//	}
//
//	defer fd.Close()
//
//	if _, err = fd.Write(data); err != nil {
//		return DBError{"Error writting table."}
//	}
//	return nil
//}
//
//func (db *engine) Delete(q *query.Query) error {
//	fd, err := os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		return DBError{"Table not found."}
//	}
//
//	scanner := bufio.NewScanner(fd)
//	if !scanner.Scan() {
//		return DBError{"Table not found."}
//	}
//
//	header := strings.Split(scanner.Text(), "<;;>")
//	headers := header[:len(header)-1]
//
//	filters := make(map[int]string)
//	for i, hd := range headers {
//		if hd[:2] == "U%" {
//			hd = hd[2:]
//		}
//
//		for filter, value := range q.f {
//			if hd == filter {
//				filters[i] = value
//			}
//		}
//	}
//
//	var content []string
//	for scanner.Scan() {
//		content = strings.Split(scanner.Text(), "<;;>")
//		match := true
//		for pos, value := range filters {
//			match = value == content[pos]
//			if !match {
//				break
//			}
//		}
//		if match {
//			break
//		}
//	}
//
//	fd.Close()
//
//	old := strings.Join(content, "<;;>") + "\n"
//	for i, hd := range headers {
//		for key, value := range q.d {
//			if hd == key {
//				content[i] = value
//			}
//		}
//	}
//
//	fd, err = os.OpenFile(db.location+q.t, os.O_RDWR|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		return DBError{"Error opening table."}
//	}
//
//	fileinfo, _ := fd.Stat()
//	data := make([]byte, fileinfo.Size())
//	if _, err = fd.Read(data); err != nil {
//		return DBError{"Error reading table."}
//	}
//
//	data = bytes.Replace(data, []byte(old), []byte(""), -1)
//	if err = fd.Truncate(0); err != nil {
//		return DBError{"Error truncating table."}
//	}
//
//	defer fd.Close()
//
//	if _, err = fd.Write(data); err != nil {
//		return DBError{"Error writting table."}
//	}
//	return nil
//}
//
//func (db *engine) Get(q *query.Query) ([]map[string]string, error) {
//	fd, err := os.OpenFile(db.location+q.t, os.O_RDONLY, os.ModePerm)
//	if err != nil {
//		return nil, DBError{"Table not found."}
//	}
//
//	scanner := bufio.NewScanner(fd)
//	if !scanner.Scan() {
//		return nil, DBError{"Table not found."}
//	}
//
//	header := strings.Split(scanner.Text(), "<;;>")
//	headers := header[:len(header)-1]
//
//	defer fd.Close()
//
//	filters := make(map[int]string)
//	for i, hd := range headers {
//		if hd[:2] == "U%" {
//			hd = hd[2:]
//		}
//
//		for filter, value := range q.f {
//			if hd == filter {
//				filters[i] = value
//			}
//		}
//	}
//
//	var result []map[string]string
//	for scanner.Scan() {
//		content := strings.Split(scanner.Text(), "<;;>")
//		match := true
//		for pos, value := range filters {
//			match = value == content[pos]
//			if !match {
//				break
//			}
//		}
//		if match {
//			item := make(map[string]string)
//			for i, hd := range headers {
//				if hd[:2] == "U%" {
//					hd = hd[2:]
//				}
//				item[hd] = content[i]
//			}
//			result = append(result, item)
//		}
//	}
//
//	err = DBError{"No result."}
//	if len(result) > 0 {
//		err = nil
//	}
//	return result, err
//
//}
//
//func (db *engine) GetOne(q *query.Query) (map[string]string, error) {
//	fd, err := os.OpenFile(db.location+q.t, os.O_RDONLY, os.ModePerm)
//	if err != nil {
//		return nil, DBError{"Table not found."}
//	}
//
//	scanner := bufio.NewScanner(fd)
//	if !scanner.Scan() {
//		return nil, DBError{"Table not found."}
//	}
//
//	header := strings.Split(scanner.Text(), "<;;>")
//	headers := header[:len(header)-1]
//
//	defer fd.Close()
//
//	filters := make(map[int]string)
//	for i, hd := range headers {
//		if hd[:2] == "U%" {
//			hd = hd[2:]
//		}
//
//		for filter, value := range q.f {
//			if hd == filter {
//				filters[i] = value
//			}
//		}
//	}
//
//	for scanner.Scan() {
//		content := strings.Split(scanner.Text(), "<;;>")
//		match := true
//		for pos, value := range filters {
//			match = value == content[pos]
//			if !match {
//				break
//			}
//		}
//		if match {
//			item := make(map[string]string)
//			for i, hd := range headers {
//				if hd[:2] == "U%" {
//					hd = hd[2:]
//				}
//				item[hd] = content[i]
//			}
//			return item, nil
//		}
//	}
//
//	return nil, DBError{"No result."}
//}
