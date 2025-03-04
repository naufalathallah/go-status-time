package utils

import (
	"strings"
	"time"
)

func GroupByColumnData(headers []string, records [][]string, startDate, endDate time.Time) map[string]map[string]interface{} {
	columnPairs := map[string]string{
		"Code Fixing-GQA":         "'->Code Fixing-GQA",
		"Code Review-GQA":         "'->Code Review-GQA",
		"Done-GQA":                "'->Done-GQA",
		"In Progress-GQA":         "'->In Progress-GQA",
		"on hold-GQA":             "'->on hold-GQA",
		"Reject Code Review-GQA":  "'->Reject Code Review-GQA",
		"To Do-GQA":               "'->To Do-GQA",
	}

	groupedData := make(map[string]map[string]interface{})
	indexMap := make(map[string]int)

	for key := range columnPairs {
		for i, header := range headers {
			if header == key || header == columnPairs[key] {
				indexMap[header] = i
			}
		}
	}

	extraColumns := []string{"Issue Type", "Summary", "Status", "Assignee", "Labels", "Story point estimate", "Created", "Bloked-GQA"}
	for _, col := range extraColumns {
		for i, header := range headers {
			if header == col {
				indexMap[col] = i
			}
		}
	}

	for _, record := range records {
		key := record[0]
		if _, exists := groupedData[key]; !exists {
			groupedData[key] = make(map[string]interface{})
		}

		for _, col := range extraColumns {
			if idx, ok := indexMap[col]; ok && idx < len(record) {
				groupedData[key][col] = strings.TrimSpace(record[idx])
			}
		}

		if _, ok := groupedData[key]["Details"]; !ok {
			groupedData[key]["Details"] = make(map[string][]string)
		}
		for mainCol, pairCol := range columnPairs {
			mainIdx, mainOk := indexMap[mainCol]
			pairIdx, pairOk := indexMap[pairCol]

			if mainOk && pairOk {
				mainValues := strings.Split(record[mainIdx], ",")
				pairValues := strings.Split(record[pairIdx], ",")

				for i := range mainValues {
					if i < len(pairValues) {
						date, err := time.Parse("2006-01-02 15:04", strings.TrimSpace(pairValues[i]))
						if err != nil {
							continue
						}

						if (date.After(startDate) || date.Equal(startDate)) && (date.Before(endDate) || date.Equal(endDate)) {
							groupedData[key]["Details"].(map[string][]string)[mainCol] = append(groupedData[key]["Details"].(map[string][]string)[mainCol], strings.TrimSpace(mainValues[i])+" ("+strings.TrimSpace(pairValues[i])+")")
						}
					}
				}
			}
		}
	}

	return groupedData
}
