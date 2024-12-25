package utils

import (
	"strings"
)

func GroupByColumnData(headers []string, records [][]string) map[string]map[string][]string {
	columnPairs := map[string]string{
		"Code Fixing-GQA":         "'->Code Fixing-GQA",
		"Code Review-GQA":         "'->Code Review-GQA",
		"Done-GQA":                "'->Done-GQA",
		"In Progress-GQA":         "'->In Progress-GQA",
		"on hold-GQA":             "'->on hold-GQA",
		"Reject Code Review-GQA":  "'->Reject Code Review-GQA",
		"To Do-GQA":               "'->To Do-GQA",
	}

	groupedData := make(map[string]map[string][]string)

	indexMap := make(map[string]int)
	for key := range columnPairs {
		for i, header := range headers {
			if header == key || header == columnPairs[key] {
				indexMap[header] = i
			}
		}
	}

	for _, record := range records {
		// Ambil key (kolom "Key")
		key := record[0]

		if _, exists := groupedData[key]; !exists {
			groupedData[key] = make(map[string][]string)
		}

		for mainCol, pairCol := range columnPairs {
			mainIdx, mainOk := indexMap[mainCol]
			pairIdx, pairOk := indexMap[pairCol]

			if mainOk && pairOk {
				mainValues := strings.Split(record[mainIdx], "\n")
				pairValues := strings.Split(record[pairIdx], "\n")

				for i := range mainValues {
					if i < len(pairValues) {
						groupedData[key][mainCol] = append(groupedData[key][mainCol], mainValues[i]+" ("+pairValues[i]+")")
					}
				}
			}
		}
	}

	return groupedData
}
