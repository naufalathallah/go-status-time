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

func WeeklyHourHandler(c *fiber.Ctx) error {
	var req JQLRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Parse tanggal dari input
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid startDate format")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid endDate format")
	}

	// Bangun JQL
	projectList := strings.Join(req.Project, ", ")
	jql := fmt.Sprintf(
		`project IN (%s) AND assignee = %s AND updated >= "%s" AND updated <= "%s" ORDER BY created DESC`,
		projectList, req.Assignee, req.StartDate, req.EndDate,
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

	// Kumpulkan worklog per tanggal
	dailyTotals := make(map[string]int)
	totalSeconds := 0

	for _, issue := range searchResponse.Issues {
		issueKey := issue.Key
		worklogURL := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog", jiraDomain, issueKey)

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

			// Cek apakah worklog dalam range tanggal
			if startedTime.Before(startDate) || startedTime.After(endDate) {
				continue
			}

			dateKey := startedTime.Format("2006-01-02")
			dailyTotals[dateKey] += wl.TimeSpentSeconds
			totalSeconds += wl.TimeSpentSeconds
		}
	}

	// Konversi hasil daily ke jam
	dailyHours := make(map[string]float64)
	for date, seconds := range dailyTotals {
		dailyHours[date] = float64(seconds) / 3600
	}

	return c.JSON(fiber.Map{
		"start_date":            req.StartDate,
		"end_date":              req.EndDate,
		"total_time_seconds":    totalSeconds,
		"total_time_hours":      fmt.Sprintf("%.2f", float64(totalSeconds)/3600),
		"daily_time_hours":      dailyHours,
	})
}