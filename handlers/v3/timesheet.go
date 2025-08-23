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

	// Process worklog data
	timesheet := processWorklogData(worklogData)

	// Return as plain text response
	return c.SendString(timesheet)
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

func processWorklogData(worklogData *WorklogResponse) string {
	// Map to group by date
	dateMap := make(map[string][]string)

	for _, issue := range worklogData.Issues {
		for _, worklog := range issue.Fields.Worklog.Worklogs {
			// Parse started date
			startedTime, err := time.Parse("2006-01-02T15:04:05.000-0700", worklog.Started)
			if err != nil {
				// Try alternative format
				startedTime, err = time.Parse("2006-01-02T15:04:05.000+0700", worklog.Started)
				if err != nil {
					continue
				}
			}

			// Format date as "1 Januari 2025"
			dateKey := formatIndonesianDate(startedTime)

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

	// Sort dates and build result
	var dates []string
	for date := range dateMap {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	var result []string
	for _, date := range dates {
		entries := strings.Join(dateMap[date], ", ")
		result = append(result, fmt.Sprintf("%s: %s", date, entries))
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