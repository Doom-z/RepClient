package fileutil

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// mode append | overwrite
// SaveData saves data to a file based on the provided format (json, txt, csv).
func SaveData(data any, outputFile string, mode string) error {
	ext := strings.ToLower(filepath.Ext(outputFile))
	switch ext {
	case ".ndjson":
		if mode == "append" {
			return appendAsNDJSON(data, outputFile)
		}
		return saveAsNDJSON(data, outputFile)
	case ".json":
		if mode == "append" {
			return appendAsJSON(data, outputFile)
		}
		return saveAsJSON(data, outputFile) // JSON array (overwrites)
	case ".txt":
		if mode == "append" {
			return appendAsTxt(data, outputFile)
		}

		return saveAsTxt(data, outputFile)
	case ".csv":
		return saveAsCSV(data, outputFile)
	default:
		return fmt.Errorf("unsupported file format: %s in file %s", ext, outputFile)
	}
}

func saveAsNDJSON(data any, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	switch v := data.(type) {
	case []any:
		for _, item := range v {
			if err := encoder.Encode(item); err != nil {
				return err
			}
		}
	case []map[string]any:
		for _, item := range v {
			if err := encoder.Encode(item); err != nil {
				return err
			}
		}
	default:
		return encoder.Encode(data)
	}

	return nil
}

func appendAsNDJSON(data any, path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	switch v := data.(type) {
	case []any:
		for _, item := range v {
			if err := encoder.Encode(item); err != nil {
				return err
			}
		}
	case []map[string]any:
		for _, item := range v {
			if err := encoder.Encode(item); err != nil {
				return err
			}
		}
	default:
		return encoder.Encode(data)
	}

	return nil
}

// not optimal because need read the file first, if you want a high throughput use ndjson instead.
func appendAsJSON(data any, path string) error {
	var existing []any

	// Read existing data if file exists
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&existing); err != nil {
			// fallback to empty if file is not a valid array
			existing = []any{}
		}
	}

	// Normalize new data into []any
	var newItems []any
	switch v := data.(type) {
	case []any:
		newItems = v
	case []map[string]any:
		for _, item := range v {
			newItems = append(newItems, item)
		}
	default:
		newItems = append(existing, v)
		existing = nil // keep only the new item in list
	}

	existing = append(existing, newItems...)

	// Write combined array back to file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	return encoder.Encode(existing)
}

func saveAsJSON(data any, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func appendAsTxt(data any, path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	switch v := data.(type) {
	case string:
		_, _ = fmt.Fprintln(file, v)
	case []string:
		for _, line := range v {
			_, _ = fmt.Fprintln(file, line)
		}
	case []fmt.Stringer:
		for _, item := range v {
			_, _ = fmt.Fprintln(file, item.String())
		}
	default:
		return errors.New("unsupported type for .txt append; expected string, []string, or []fmt.Stringer")
	}
	return nil
}

func saveAsTxt(data any, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch v := data.(type) {
	case string:
		_, _ = fmt.Fprintln(file, v+"\n")
	case []string:
		for _, line := range v {
			_, _ = fmt.Fprintln(file, line)
		}
	case []fmt.Stringer:
		for _, item := range v {
			_, _ = fmt.Fprintln(file, item.String())
		}
	default:
		return errors.New("unsupported type for .txt export; expected []string or []fmt.Stringer")
	}
	return nil
}

func saveAsCSV(data any, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	switch records := data.(type) {
	case [][]string:
		return writer.WriteAll(records)
	case []map[string]string:
		if len(records) == 0 {
			return nil
		}
		// write header
		var header []string
		for k := range records[0] {
			header = append(header, k)
		}
		writer.Write(header)

		// write rows
		for _, row := range records {
			var line []string
			for _, h := range header {
				line = append(line, row[h])
			}
			writer.Write(line)
		}
	default:
		return errors.New("unsupported type for .csv export; expected [][]string or []map[string]string")
	}

	return nil
}
