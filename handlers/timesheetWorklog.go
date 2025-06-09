package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type JQLRequest struct {
	Project   []string `json:"project"`
	Assignee  string   `json:"assignee"`
	StartDate string   `json:"startDate"`
	EndDate   string   `json:"endDate"`
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

	// You can unmarshal and filter or just return raw
	return c.Type("json").Send(body)
}