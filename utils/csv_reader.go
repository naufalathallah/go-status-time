package utils

import (
	"encoding/csv"
	"io"
)

func ParseCSV(file io.Reader) ([]string, [][]string, error) {
	reader := csv.NewReader(file)

	headers, err := reader.Read()
	if err != nil {
		return nil, nil, err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	return headers, records, nil
}