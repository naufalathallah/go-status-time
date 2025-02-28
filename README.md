# Go Status Time

A Go web service for processing and exporting timesheet data to Excel format.

## Setup

### Prerequisites

- Go 1.23.3 or higher
- Git

### Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/naufalathallah/go-status-time.git
   cd go-status-time
   ```

2. Install dependencies:
   ```sh
   go mod download
   ```

## Running the Application

Start the server:

```sh
go run main.go
```

The server will run at `http://localhost:8000`

## API Endpoints

### 1. Weekly Report

**Endpoint:** `POST /weekly`

Generate weekly report in Excel format from CSV data.

**Form Parameters:**

- `file`: CSV file containing timesheet data
- `startDate`: Start date in format `YYYY/MM/DD`
- `endDate`: End date in format `YYYY/MM/DD`

**Example Request:**
`sh
    curl -X POST \
      -F "file=@data.csv" \
      -F "startDate=2024/03/01" \
      -F "endDate=2024/03/07" \
      http://localhost:8000/weekly
    `

### 2. Timesheet Export

**Endpoint:** `POST /timesheet`

Generate formatted timesheet in Excel format from CSV data.

**Form Parameters:**

- `file`: CSV file containing timesheet data
- `startDate`: Start date in format `YYYY/MM/DD`
- `endDate`: End date in format `YYYY/MM/DD`

**Example Request:**
`sh
    curl -X POST \
      -F "file=@data.csv" \
      -F "startDate=2024/03/01" \
      -F "endDate=2024/03/07" \
      http://localhost:8000/timesheet
    `

## Project Structure

```
├── handlers/
│   ├── timesheet.go    # Timesheet endpoint handler
│   └── weekly.go       # Weekly report endpoint handler
├── utils/
│   ├── csv_reader.go           # CSV parsing utilities
│   ├── excel.go               # Excel file generation
│   ├── export_timesheet.go    # Timesheet export logic
│   ├── format_number.go       # Number formatting utilities
│   └── group_by_column_data.go # Data grouping logic
├── go.mod
├── go.sum
└── main.go            # Application entry point
```

## Response Format

Both endpoints return an Excel file (.xlsx) containing the processed data. The file will be named using the pattern:

- Weekly report: `weekly-YYYY-MM-DD-YYYY-MM-DD.xlsx`
- Timesheet: `timesheet-YYYY-MM-DD-YYYY-MM-DD.xlsx`

## Dependencies

- [fiber/v2](https://github.com/gofiber/fiber) - Web framework
- [excelize/v2](https://github.com/xuri/excelize) - Excel file handling

## Error Handling

The API returns appropriate HTTP status codes:

- 400 Bad Request: Invalid input parameters
- 405 Method Not Allowed: Wrong HTTP method
- 500 Internal Server Error: Server-side processing errors

All error responses include a descriptive error message in the response body.
