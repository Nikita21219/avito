package reportcsv

import (
	"encoding/csv"
	"main/internal/history"
	"os"
)

var (
	CSVDir      = "./csv_report_storage/"
	FilePostfix = ".csv"
)

func CreateReport(story []history.HistoryDto, fileName string) error {
	// TODO read data from story

	file, err := os.Create(CSVDir + fileName + FilePostfix)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	headers := []string{"user", "segment", "operation", "date"}
	data := [][]string{
		{"2", "1", "deleted", "2023-08-30 08:31:15.000000"},
		{"2", "1", "added", "2023-08-30 08:31:36.000000"},
		{"2", "1", "deleted", "2023-08-30 08:32:03.000000"},
	}
	err = writer.Write(headers)
	if err != nil {
		return err
	}
	for _, row := range data {
		err = writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}
