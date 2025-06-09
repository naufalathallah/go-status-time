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

func DailyHourHandler(c *fiber.Ctx) error {
	var req JQLRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

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

	issueKeys := []string{}
	for _, issue := range searchResponse.Issues {
		issueKeys = append(issueKeys, issue.Key)
	}

	dailyTotals := make(map[string]int)
	for _, issueKey := range issueKeys {
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
			dateKey := startedTime.Format("2006-01-02")
			dailyTotals[dateKey] += wl.TimeSpentSeconds
		}
	}

	dailyHours := make(map[string]float64)
	for date, seconds := range dailyTotals {
		dailyHours[date] = float64(seconds) / 3600
	}

	return c.JSON(fiber.Map{
		"daily_hours": dailyHours,
	})
}