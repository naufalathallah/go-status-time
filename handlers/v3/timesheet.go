package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
)

type WorklogSearchResponse struct {
	Issues []JiraIssue `json:"issues"`
}

type JiraIssue struct {
	Key    string          `json:"key"`
	Fields JiraIssueFields `json:"fields"`
}

type JiraIssueFields struct {
	Worklog Worklog `json:"worklog"`
}

type Worklog struct {
	Worklogs []WorklogEntry `json:"worklogs"`
}

type WorklogEntry struct {
	Updated          string  `json:"updated"`
	Started          string  `json:"started"`
	TimeSpentSeconds int     `json:"timeSpentSeconds"`
	Comment          Comment `json:"comment"`
	Author           Author  `json:"author"`
}

type Comment struct {
	Content []CommentContent `json:"content"`
}

type CommentContent struct {
	Content []CommentText `json:"content"`
}

type CommentText struct {
	Text string `json:"text"`
}

type Author struct {
	AccountId   string `json:"accountId"`
	DisplayName string `json:"displayName"`
}

func TimesheetHandler(c *fiber.Ctx) error {
	// Get query parameters
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	authorId := c.Query("authorId")

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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create request",
			"details": err.Error(),
		})
	}

	// Add headers
	authToken := os.Getenv("JIRA_AUTH")

	req.Header.Add("Authorization", authToken)
	req.Header.Add("Accept", "application/json")

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to make request",
			"details": err.Error(),
		})
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read response",
			"details": err.Error(),
		})
	}

	// Check status code and return response directly
	if resp.StatusCode != http.StatusOK {
		return c.Status(resp.StatusCode).JSON(fiber.Map{
			"error": "Atlassian API error",
			"status": resp.StatusCode,
			"response": string(body),
		})
	}

	// Return raw response from Atlassian API
	c.Set("Content-Type", "application/json")
	return c.Send(body)
}