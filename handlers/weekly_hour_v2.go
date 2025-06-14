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
)

func WeeklyHourV2Handler(c *fiber.Ctx) error {
    var req JQLRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
    }

    // Initialize date range
    startDate := req.StartDate
    endDate := req.EndDate

    // If dates are not provided, default to today
    if startDate == "" || endDate == "" {
        today := time.Now().Format("2006-01-02")
        startDate = today
        endDate = today
    }

    // Validate date format
    _, err := time.Parse("2006-01-02", startDate)
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid start date format. Use YYYY-MM-DD")
    }
    
    _, err = time.Parse("2006-01-02", endDate)
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid end date format. Use YYYY-MM-DD")
    }

    // JQL with date range
    projectList := strings.Join(req.Project, ", ")
    jql := fmt.Sprintf(
        `project IN (%s) AND assignee = %s AND worklogDate >= "%s" AND worklogDate <= "%s" ORDER BY created DESC`,
        projectList, req.Assignee, startDate, endDate,
    )

    jqlEncoded := strings.ReplaceAll(jql, " ", "%20")
    jqlEncoded = strings.ReplaceAll(jqlEncoded, `"`, "%22")
    jqlEncoded = strings.ReplaceAll(jqlEncoded, ":", "%3A")

    jiraDomain := "https://lionparcel.atlassian.net"
    auth := os.Getenv("JIRA_AUTH")
    url := fmt.Sprintf("%s/rest/api/3/search?jql=%s&maxResults=100&startAt=0", jiraDomain, jqlEncoded)

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

    var worklogs []map[string]interface{}
    var totalSeconds int

    // Map to track daily totals
    dailyTotals := make(map[string]int)

    for _, issue := range searchResponse.Issues {
        issueKey := issue.Key
        worklogURL := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog", jiraDomain, issueKey)
        fmt.Println("Fetching worklog for issue:", issueKey)

        reqW, _ := http.NewRequest("GET", worklogURL, nil)
        reqW.Header.Add("Authorization", auth)
        reqW.Header.Add("Accept", "application/json")

        respW, err := client.Do(reqW)
        if err != nil {
            continue
        }
        defer respW.Body.Close()
        bodyW, _ := io.ReadAll(respW.Body)

        var wlResp WorklogResponse
        if err := json.Unmarshal(bodyW, &wlResp); err != nil {
            continue
        }

        for _, wl := range wlResp.Worklogs {
            startedTime, err := time.Parse("2006-01-02T15:04:05.000-0700", wl.Started)
            if err != nil {
                continue
            }

            worklogDate := startedTime.Format("2006-01-02")
            
            // Check if the worklog date is within the requested range
            if worklogDate < startDate || worklogDate > endDate {
                continue
            }

            comment := ""
            if len(wl.Comment.Content) > 0 && len(wl.Comment.Content[0].Content) > 0 {
                comment = wl.Comment.Content[0].Content[0].Text
            }

            worklogs = append(worklogs, map[string]interface{}{
                "issue_key":         issueKey,
                "started":           startedTime.Format("2006-01-02 15:04"),
                "date":              worklogDate,
                "time_spent_hours":  float64(wl.TimeSpentSeconds) / 3600,
                "raw_time_seconds":  wl.TimeSpentSeconds,
                "comment":           comment,
            })

            // Update total seconds
            totalSeconds += wl.TimeSpentSeconds
            
            // Update daily totals
            dailyTotals[worklogDate] += wl.TimeSpentSeconds
        }
    }

    // Convert daily totals to a list for the response
    dailySummary := make([]map[string]interface{}, 0, len(dailyTotals))
    for date, seconds := range dailyTotals {
        dailySummary = append(dailySummary, map[string]interface{}{
            "date":           date,
            "hours":          fmt.Sprintf("%.2f", float64(seconds)/3600),
            "total_seconds":  seconds,
        })
    }

    return c.JSON(fiber.Map{
        "total_time_hours":   fmt.Sprintf("%.2f", float64(totalSeconds)/3600),
        "daily_summary":      dailySummary,
        "worklogs":           worklogs,
    })
}