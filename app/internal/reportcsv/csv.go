package reportcsv

import (
	"encoding/csv"
	"main/internal/history"
	"os"
	"strconv"
)

var (
	CSVDir      = "./csv_report_storage/"
	FilePostfix = ".csv"
)

// CreateReport creates a csv report and writes the data to a file.
func CreateReport(story []history.HistoryDto, fileName string) error {
	file, err := os.Create(CSVDir + fileName + FilePostfix)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"user", "segment", "operation", "date"}
	if err = writer.Write(headers); err != nil {
		return err
	}

	for _, row := range story {
		csvRow := []string{
			strconv.Itoa(row.UserId),
			row.SegmentSlug,
			row.Operation,
			row.Date,
		}
		if err = writer.Write(csvRow); err != nil {
			return err
		}
	}

	return nil
}
