package engine

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"stupiddb/query"
)

type DBError struct {
	Message string
}

func (err DBError) Error() string {
	return fmt.Sprintf("%v", err.Message)
}

type Engine struct {
	database []os.FileInfo
	location string
}

func CreateInstance(name string) error {
	user, err := user.Current()
	if err != nil {
		return err
	}

	path := user.HomeDir + "/.godb/"
	if _, err = os.Stat(path); err != nil {
		if err = os.Mkdir(path, os.ModePerm); err != nil {
			return err
		}
	}

	database := base64.StdEncoding.EncodeToString([]byte(name))

	if _, err = os.Stat(path + database); err != nil {
		if err = os.Mkdir(path+database, os.ModePerm); err != nil {
			return err
		}
		return nil
	} else {
		return err
	}
}

func Instance(schema string) (*Engine, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	path := user.HomeDir + "/.godb/"

	database_name := base64.StdEncoding.EncodeToString([]byte(schema))

	if _, err := os.Stat(path + database_name); err != nil {
		return nil, err
	}

	location := path + database_name

	var database []os.FileInfo
	if database, err = ioutil.ReadDir(location); err != nil {
		return nil, err
	}

	return &Engine{database, location + "/"}, nil
}

func (db *Engine) CreateTable(name string, fields []string) error {
	table := base64.StdEncoding.EncodeToString([]byte(name))

	fd, err := os.Create(db.location + table)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.Write([]byte(strings.Join(fields, ";") + ";\n"))
	return err
}

func Unique(data string) string {
	return "U%" + data
}

func (db *Engine) Add(query *query.Query) error {
	table := base64.StdEncoding.EncodeToString([]byte(query.T))

	fd, err := os.OpenFile(db.location+table, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return DBError{"Table not found."}
	}

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return DBError{"Table not found."}
	}

	header := strings.Split(scanner.Text(), ";")
	headers := header[:len(header)-1]

	defer fd.Close()

	if len(headers) == len(query.D) {
		for scanner.Scan() {
			fields := strings.Split(scanner.Text(), ";")
			for i, hd := range headers {
				if hd[:2] == "U%" && query.D[hd[2:]] == fields[i] {
					return DBError{"Unique row '" + hd + "' restriction violated."}
				}
			}
		}

		var data string
		for _, hd := range headers {
			isField := false
			for column, value := range query.D {
				if hd[:2] == "U%" {
					hd = hd[2:]
				}

				if column == hd {
					data += value + ";"
					isField = true
				} else {
					isField = isField || false
				}
			}
			if isField != true {
				return DBError{"Table struct and data attributes doesn't match."}
			}

		}
		_, err = fd.Write([]byte(data + "\n"))
		return err
	}

	return DBError{"Attribute length mismatch."}
}

func (db *Engine) Get(query *query.Query) ([]map[string]string, error) {
	table := base64.StdEncoding.EncodeToString([]byte(query.T))

	fd, err := os.OpenFile(db.location+table, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, DBError{"Table not found."}
	}

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return nil, DBError{"Table not found."}
	}

	header := strings.Split(scanner.Text(), ";")
	headers := header[:len(header)-1]

	defer fd.Close()

	filters := make(map[int]string)
	for i, hd := range headers {
		if hd[:2] == "U%" {
			hd = hd[2:]
		}

		for filter, value := range query.F {
			if hd == filter {
				filters[i] = value
			}
		}
	}

	var result []map[string]string
	for scanner.Scan() {
		content := strings.Split(scanner.Text(), ";")
		match := true
		for pos, value := range filters {
			match = value == content[pos]
			if !match {
				break
			}
		}
		if match {
			item := make(map[string]string)
			for i, hd := range headers {
				if hd[:2] == "U%" {
					hd = hd[2:]
				}
				item[hd] = content[i]
			}
			result = append(result, item)
		}
	}

	err = DBError{"No result."}
	if len(result) > 0 {
		err = nil
	}
	return result, err

}

func (db *Engine) GetOne(query *query.Query) (map[string]string, error) {
	table := base64.StdEncoding.EncodeToString([]byte(query.T))

	fd, err := os.OpenFile(db.location+table, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, DBError{"Table not found."}
	}

	scanner := bufio.NewScanner(fd)
	if !scanner.Scan() {
		return nil, DBError{"Table not found."}
	}

	header := strings.Split(scanner.Text(), ";")
	headers := header[:len(header)-1]

	defer fd.Close()

	filters := make(map[int]string)
	for i, hd := range headers {
		if hd[:2] == "U%" {
			hd = hd[2:]
		}

		for filter, value := range query.F {
			if hd == filter {
				filters[i] = value
			}
		}
	}

	for scanner.Scan() {
		content := strings.Split(scanner.Text(), ";")
		match := true
		for pos, value := range filters {
			match = value == content[pos]
			if !match {
				break
			}
		}
		if match {
			item := make(map[string]string)
			for i, hd := range headers {
				if hd[:2] == "U%" {
					hd = hd[2:]
				}
				item[hd] = content[i]
			}
			return item, nil
		}
	}

	return nil, DBError{"No result."}
}
