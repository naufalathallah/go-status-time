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

## Generate Timesheet Excel

Untuk generate timesheet dalam format Excel, ikuti langkah berikut:

1. **Sesuaikan Start Date dan End Date**

   - Ubah parameter `startDate` dan `endDate` sesuai periode yang diinginkan
   - Format: `YYYY-MM-DD`

2. **Sesuaikan Author ID**

   - Ganti parameter `authorId` dengan Jira User ID Anda
   - Format: `712020:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

3. **Kirim Request**

   ```sh
   curl --location 'http://localhost:8000/v3/timesheet?startDate=2025-10-15&endDate=2025-11-15&authorId=712020%3Ab0f4204d-cf46-477e-b09e-056718804914' \
   --header 'Accept: application/json'
   ```

4. **Save Response**
   - Setelah request berhasil, save response sebagai file Excel (.xlsx)
   - File yang dihasilkan berisi timesheet dalam format Excel yang sudah terformat

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
