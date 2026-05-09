package device

import (
	"encoding/csv"
	"os"
)

func LoadDevicesCSV(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}

	deviceIDs := make([]string, 0, len(records))

	for i, record := range records {
		if i == 0 {
			continue // skip header
		}
		if len(record) == 0 || record[0] == "" {
			continue
		}

		deviceIDs = append(deviceIDs, record[0])
	}

	return deviceIDs, nil
}
