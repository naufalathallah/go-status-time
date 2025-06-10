package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/naufalathallah/go-status-time/utils"
)

type JQLRequest struct {
	Project   []string `json:"project"`
	Assignee  string   `json:"assignee"`
	StartDate string   `json:"startDate"`
	EndDate   string   `json:"endDate"`
}

type JiraIssue struct {
	Key string `json:"key"`
	Fields JiraIssueField `json:"fields"`
}

type JiraSearchResponse struct {
	Issues []JiraIssue `json:"issues"`
}

type CommentText struct {
	Text string `json:"text"`
}

type CommentContent struct {
	Content []CommentText `json:"content"`
}

type Comment struct {
	Content []CommentContent `json:"content"`
}

type WorklogEntry struct {
	Updated string  `json:"updated"`
	Started string  `json:"started"` 
	Comment Comment `json:"comment"`
	TimeSpentSeconds int     `json:"timeSpentSeconds"`
}

type WorklogResponse struct {
	Worklogs []WorklogEntry `json:"worklogs"`
}

func TimesheetWorklogHandler(c *fiber.Ctx) error {
	var req JQLRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Build JQL
	projectList := strings.Join(req.Project, ", ")
	jql := fmt.Sprintf(
		`project IN (%s) AND assignee = %s AND updated >= "%s" AND updated <= "%s" ORDER BY created DESC`,
		projectList, req.Assignee, req.StartDate, req.EndDate,
	)

	// Encode JQL
	jqlEncoded := strings.ReplaceAll(jql, " ", "%20")
	jqlEncoded = strings.ReplaceAll(jqlEncoded, `"`, "%22")
	jqlEncoded = strings.ReplaceAll(jqlEncoded, ":", "%3A")

	// Jira API config
	jiraDomain := "https://lionparcel.atlassian.net"
	auth := os.Getenv("JIRA_AUTH")

	// Initial URL
	url := fmt.Sprintf("%s/rest/api/3/search?jql=%s&maxResults=100&startAt=0", jiraDomain, jqlEncoded)

	// Make the request
	reqClient, _ := http.NewRequest("GET", url, nil)
	reqClient.Header.Add("Authorization", auth)
	reqClient.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(reqClient)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to request Jira")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var searchResponse JiraSearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse Jira response")
	}

	// Kumpulkan semua key
	var issueKeys []string
	for _, issue := range searchResponse.Issues {
		issueKeys = append(issueKeys, issue.Key)
	}
	fmt.Println("Issue keys:", issueKeys)

	var worklogData []map[string]interface{}
	for _, issueKey := range issueKeys {
		worklogURL := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog", jiraDomain, issueKey)
		fmt.Println("Fetching worklog for issue:", issueKey)

		reqW, _ := http.NewRequest("GET", worklogURL, nil)
		reqW.Header.Add("Authorization", auth)
		reqW.Header.Add("Accept", "application/json")

		respW, err := client.Do(reqW)
		if err != nil {
			fmt.Println("❌ Failed to get worklog for", issueKey)
			continue
		}
		defer respW.Body.Close()

		bodyW, _ := io.ReadAll(respW.Body)

		var wlResp WorklogResponse
		if err := json.Unmarshal(bodyW, &wlResp); err != nil {
			fmt.Println("❌ Failed to parse worklog for", issueKey)
			continue
		}

		for _, wl := range wlResp.Worklogs {
			var text string
			if len(wl.Comment.Content) > 0 && len(wl.Comment.Content[0].Content) > 0 {
				text = wl.Comment.Content[0].Content[0].Text
			}

			worklogData = append(worklogData, map[string]interface{}{
				"issue_key": issueKey,
				"updated":   wl.Updated,
				"started":   wl.Started,
				"comment":   text,
			})
		}
	}

	// Parse start & end date dari request
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid startDate format")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid endDate format")
	}

	// Format worklogs ke bentuk timesheet
	timesheetText := FormatWorklogTimesheet(worklogData, startDate, endDate)

	fmt.Println("=== Timesheet Data ===")
	fmt.Println(timesheetText)

	// Create and export Excel file
	excelFile, err := utils.ExportTimesheetWorklog(timesheetText, startDate, endDate)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create Excel file")
	}

	// Save Excel to buffer
	buffer, err := excelFile.WriteToBuffer()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save Excel file")
	}

	// Set headers and send Excel file
	filename := fmt.Sprintf("worklog-timesheet-%s-%s.xlsx", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Return both JSON and Excel file options
	return c.Send(buffer.Bytes())
}

func FormatWorklogTimesheet(worklogs []map[string]interface{}, startDate, endDate time.Time) string {
	grouped := make(map[string][]string)

	for _, entry := range worklogs {
		updatedRaw, ok := entry["started"].(string)
		if !ok {
			continue
		}
		updatedTime, err := time.Parse("2006-01-02T15:04:05.000-0700", updatedRaw)
		if err != nil {
			continue
		}

		if updatedTime.Before(startDate) || updatedTime.After(endDate) {
			continue
		}

		comment, _ := entry["comment"].(string)
		issueKey, _ := entry["issue_key"].(string)

		// Format tanggal ke 2 June 2025
		dateKey := updatedTime.Format("2 January 2006")
		text := fmt.Sprintf("%s - %s", comment, issueKey)

		grouped[dateKey] = append(grouped[dateKey], text)
	}

	// Buat output final
	var output strings.Builder
	currentDate := startDate
	for !currentDate.After(endDate) {
		dateKey := currentDate.Format("2 January 2006")
		output.WriteString(dateKey + ":\n")

		if activities, ok := grouped[dateKey]; ok {
			// hapus duplikat
			unique := make(map[string]bool)
			var filtered []string
			for _, act := range activities {
				if !unique[act] {
					unique[act] = true
					filtered = append(filtered, act)
				}
			}
			if len(filtered) > 0 {
				output.WriteString(strings.Join(filtered, ", ") + "\n")
			}
		}
		currentDate = currentDate.Add(24 * time.Hour)
	}

	return output.String()
}