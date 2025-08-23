package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

type WorklogResponse struct {
	Issues []Issue `json:"issues"`
}

type Issue struct {
	Key    string `json:"key"`
	Fields Fields `json:"fields"`
}

type Fields struct {
	Worklog WorklogData `json:"worklog"`
}

type WorklogData struct {
	Worklogs []WorklogEntry `json:"worklogs"`
}

type WorklogEntry struct {
	Started string  `json:"started"`
	Comment Comment `json:"comment"`
	Author  Author  `json:"author"`
}

type Author struct {
	AccountId string `json:"accountId"`
}

type Comment struct {
	Content []Content `json:"content"`
}

type Content struct {
	Content []TextContent `json:"content"`
}

type TextContent struct {
	Text string `json:"text"`
}

func TimesheetHandler(c *fiber.Ctx) error {
	// Get query parameters
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	authorId := c.Query("authorId")

	// Call worklog endpoint
	worklogData, err := getWorklogData(startDate, endDate, authorId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get worklog data",
			"details": err.Error(),
		})
	}

	// Process worklog data with author filtering
	timesheet := processWorklogDataFiltered(worklogData, authorId, startDate, endDate)

	// Print to console
	fmt.Println("Timesheet Data:")
	fmt.Println(timesheet)

	// Create Excel file
	excelFile, err := createTimesheetExcel(worklogData, authorId, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create Excel file",
			"details": err.Error(),
		})
	}

	// Set headers for Excel file download
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=timesheet_%s_to_%s.xlsx", startDate, endDate))

	// Write Excel file to response
	buffer, err := excelFile.WriteToBuffer()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to write Excel file",
			"details": err.Error(),
		})
	}

	return c.Send(buffer.Bytes())
}

func getWorklogData(startDate, endDate, authorId string) (*WorklogResponse, error) {
	// Build JQL query
	jql := fmt.Sprintf("worklogAuthor = %s AND worklogDate >= \"%s\" AND worklogDate <= \"%s\"", 
		authorId, startDate, endDate)

	// Create request URL
	baseURL := "https://lionparcel.atlassian.net/rest/api/3/search/jql"
	params := url.Values{}
	params.Add("jql", jql)
	params.Add("maxResults", "5000")
	params.Add("fields", "worklog")

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Create HTTP request
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	authToken := os.Getenv("JIRA_AUTH")
	req.Header.Add("Authorization", authToken)
	req.Header.Add("Accept", "application/json")

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var worklogResponse WorklogResponse
	if err := json.Unmarshal(body, &worklogResponse); err != nil {
		return nil, err
	}

	return &worklogResponse, nil
}

func processWorklogDataFiltered(worklogData *WorklogResponse, authorId, startDate, endDate string) string {
	// Parse date range for filtering
	startDateParsed, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Sprintf("Error parsing start date: %v", err)
	}
	endDateParsed, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Sprintf("Error parsing end date: %v", err)
	}
	// Set end date to end of day
	endDateParsedWithTime := endDateParsed.Add(24*time.Hour - time.Second)

	// Map to group by date with time.Time as key for proper sorting
	dateMap := make(map[time.Time][]string)

	// First, process all worklogs
	for _, issue := range worklogData.Issues {
		for _, worklog := range issue.Fields.Worklog.Worklogs {
			// Filter by author ID
			if worklog.Author.AccountId != authorId {
				continue
			}

			// Parse started date
			startedTime, err := time.Parse("2006-01-02T15:04:05.000-0700", worklog.Started)
			if err != nil {
				// Try alternative format
				startedTime, err = time.Parse("2006-01-02T15:04:05.000+0700", worklog.Started)
				if err != nil {
					continue
				}
			}

			// Filter by date range based on started time
			if startedTime.Before(startDateParsed) || startedTime.After(endDateParsedWithTime) {
				continue
			}

			// Use only date part (remove time) for grouping - use UTC to avoid timezone issues
			dateKey := time.Date(startedTime.Year(), startedTime.Month(), startedTime.Day(), 0, 0, 0, 0, time.UTC)

			// Extract comment tag
			commentText := extractCommentText(worklog.Comment)
			tag := extractTag(commentText)

			// Create entry
			entry := fmt.Sprintf("%s - %s", issue.Key, tag)

			// Add to date map
			if _, exists := dateMap[dateKey]; !exists {
				dateMap[dateKey] = []string{}
			}
			dateMap[dateKey] = append(dateMap[dateKey], entry)
		}
	}

	// Then, initialize all dates in range that don't have worklogs with empty slices
	for date := startDateParsed; !date.After(endDateParsed); date = date.AddDate(0, 0, 1) {
		// Use UTC to match the dateKey from worklogs
		dateKey := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		if _, exists := dateMap[dateKey]; !exists {
			dateMap[dateKey] = []string{}
		}
	}

	// Sort dates chronologically
	var dates []time.Time
	for date := range dateMap {
		dates = append(dates, date)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	var result []string
	for _, date := range dates {
		dateStr := formatIndonesianDate(date)
		if len(dateMap[date]) > 0 {
			entries := strings.Join(dateMap[date], ", ")
			result = append(result, fmt.Sprintf("%s: %s", dateStr, entries))
		} else {
			result = append(result, fmt.Sprintf("%s:", dateStr))
		}
	}

	return strings.Join(result, "\n")
}

func extractCommentText(comment Comment) string {
	var texts []string
	for _, content := range comment.Content {
		for _, textContent := range content.Content {
			if textContent.Text != "" {
				texts = append(texts, textContent.Text)
			}
		}
	}
	return strings.Join(texts, " ")
}

func extractTag(text string) string {
	// Regex to extract text within square brackets
	re := regexp.MustCompile(`\[([^\]]+)\]`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return "[" + matches[1] + "]"
	}
	// If no tag found, return the original text (truncated if too long)
	if len(text) > 50 {
		return text[:50] + "..."
	}
	return text
}

func formatIndonesianDate(t time.Time) string {
	months := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()], t.Year())
}

func createTimesheetExcel(worklogData *WorklogResponse, authorId, startDate, endDate string) (*excelize.File, error) {
	// Parse date range for filtering
	startDateParsed, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("error parsing start date: %v", err)
	}
	endDateParsed, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("error parsing end date: %v", err)
	}
	// Set end date to end of day
	endDateParsedWithTime := endDateParsed.Add(24*time.Hour - time.Second)

	// Map to group by date with time.Time as key for proper sorting
	dateMap := make(map[time.Time][]string)

	// First, process all worklogs
	for _, issue := range worklogData.Issues {
		for _, worklog := range issue.Fields.Worklog.Worklogs {
			// Filter by author ID
			if worklog.Author.AccountId != authorId {
				continue
			}

			// Parse started date
			startedTime, err := time.Parse("2006-01-02T15:04:05.000-0700", worklog.Started)
			if err != nil {
				// Try alternative format
				startedTime, err = time.Parse("2006-01-02T15:04:05.000+0700", worklog.Started)
				if err != nil {
					continue
				}
			}

			// Filter by date range based on started time
			if startedTime.Before(startDateParsed) || startedTime.After(endDateParsedWithTime) {
				continue
			}

			// Use only date part (remove time) for grouping - use UTC to avoid timezone issues
			dateKey := time.Date(startedTime.Year(), startedTime.Month(), startedTime.Day(), 0, 0, 0, 0, time.UTC)

			// Extract comment tag
			commentText := extractCommentText(worklog.Comment)
			tag := extractTag(commentText)

			// Create entry
			entry := fmt.Sprintf("%s - %s", issue.Key, tag)

			// Add to date map
			if _, exists := dateMap[dateKey]; !exists {
				dateMap[dateKey] = []string{}
			}
			dateMap[dateKey] = append(dateMap[dateKey], entry)
		}
	}

	// Then, initialize all dates in range that don't have worklogs with empty slices
	for date := startDateParsed; !date.After(endDateParsed); date = date.AddDate(0, 0, 1) {
		// Use UTC to match the dateKey from worklogs
		dateKey := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		if _, exists := dateMap[dateKey]; !exists {
			dateMap[dateKey] = []string{}
		}
	}

	// Sort dates chronologically
	var dates []time.Time
	for date := range dateMap {
		dates = append(dates, date)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	// Create Excel file
	excel := excelize.NewFile()
	sheetName := "Timesheet"
	excel.SetSheetName("Sheet1", sheetName)

	// Set headers
	excel.SetCellValue(sheetName, "A1", "Tanggal")
	excel.SetCellValue(sheetName, "B1", "Worklog")

	// Style headers
	headerStyle, _ := excel.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0E0E0"},
			Pattern: 1,
		},
	})
	excel.SetCellStyle(sheetName, "A1", "B1", headerStyle)

	// Fill data
	row := 2
	for _, date := range dates {
		dateStr := formatIndonesianDate(date)
		excel.SetCellValue(sheetName, fmt.Sprintf("A%d", row), dateStr)
		
		if len(dateMap[date]) > 0 {
			entries := strings.Join(dateMap[date], ", ")
			excel.SetCellValue(sheetName, fmt.Sprintf("B%d", row), entries)
		} else {
			excel.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "")
		}
		row++
	}

	// Auto-fit columns
	excel.SetColWidth(sheetName, "A", "A", 20)
	excel.SetColWidth(sheetName, "B", "B", 80)

	return excel, nil
}