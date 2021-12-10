package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/source"
	"github.com/xuri/excelize/v2"
)

var preferredParallelism = runtime.NumCPU() * 2

type UnicodeEscape string

func (ue UnicodeEscape) MarshalJSON() ([]byte, error) {
	return []byte(strconv.QuoteToASCII(string(ue))), nil
}

type JSONArrayWriter struct {
	w io.Writer
}

func (j JSONArrayWriter) Write(row interface{}) error {
	encoder := json.NewEncoder(j.w)
	err := encoder.Encode(row)
	if err != nil {
		return err
	}

	_, err = j.w.Write([]byte(","))
	return err
}

func withJSONArrayOutWriter(w *os.File, cb func(w JSONArrayWriter) error) error {
	_, err := w.Write([]byte("["))
	if err != nil {
		return err
	}

	err = cb(JSONArrayWriter{w})
	if err != nil {
		return err
	}

	// Find current offset
	lastChar, err := w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	if lastChar > 1 {
		// Overwrite the last comma
		lastChar = lastChar - 1
	}

	_, err = w.WriteAt([]byte("]"), lastChar)
	if err != nil {
		return err
	}

	return w.Truncate(lastChar + 1)
}

func withJSONArrayOutWriterFile(out string, cb func(w JSONArrayWriter) error) error {
	w, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer w.Close()

	return withJSONArrayOutWriter(w, cb)
}

func transformCSV(in io.Reader, out string) error {
	r := csv.NewReader(in)

	return withJSONArrayOutWriterFile(out, func(w JSONArrayWriter) error {
		isHeader := true
		var fields []string
		for {
			record, err := r.Read()
			if err == io.EOF {
				err = nil
				break
			}

			if err != nil {
				return err
			}

			if isHeader {
				for _, field := range record {
					fields = append(fields, field)
				}
				isHeader = false
				continue
			}

			row := map[string]UnicodeEscape{}
			for i, field := range fields {
				row[field] = UnicodeEscape(record[i])
			}

			err = w.Write(row)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func transformCSVFile(in, out string) error {
	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer f.Close()

	return transformCSV(f, out)
}

func transformJSON(in io.Reader, out string) error {
	w, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.ReadFrom(in)
	if err == io.EOF {
		err = nil
	}

	return err
}

func transformJSONFile(in, out string) error {
	r, err := os.Open(in)
	if err != nil {
		return err
	}
	defer r.Close()

	return transformJSON(r, out)
}

func transformParquet(in source.ParquetFile, out string) error {
	r, err := reader.NewParquetReader(in, nil, int64(preferredParallelism))
	if err != nil {
		return err
	}
	defer r.ReadStop()

	return withJSONArrayOutWriterFile(out, func(w JSONArrayWriter) error {
		size := 1000
		var offset int64 = 0
		for {
			err := r.SkipRows(offset)
			if err != nil {
				return err
			}

			rows, err := r.ReadByNumber(size)
			if err != nil {
				return err
			}

			for _, row := range rows {
				err := w.Write(row)
				if err != nil {
					return err
				}
			}

			offset += int64(size)

			if len(rows) < size {
				return nil
			}
		}
	})
}

func transformParquetFile(in, out string) error {
	r, err := local.NewLocalFileReader(in)
	if err != nil {
		return err
	}
	defer r.Close()

	return transformParquet(r, out)
}

func writeXLSXSheet(rows [][]string, w JSONArrayWriter) error {
	var header []string
	isHeader := true

	for _, r := range rows {
		if isHeader {
			header = r
			isHeader = false
			continue
		}

		row := map[string]interface{}{}
		for i, cell := range r {
			row[header[i]] = cell
		}

		err := w.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}

func transformXLSX(in *excelize.File, out string) error {
	sheets := in.GetSheetList()

	// Single sheet files get flattened into just an array, not a dict mapping sheet name to sheet contents
	if len(sheets) == 1 {
		return withJSONArrayOutWriterFile(out, func(w JSONArrayWriter) error {
			rows, err := in.GetRows(sheets[0])
			if err != nil {
				return err
			}

			return writeXLSXSheet(rows, w)
		})
	}

	w, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write([]byte("{"))
	if err != nil {
		return err
	}

	for _, sheet := range sheets {
		_, err = w.Write([]byte(`"` + strings.ReplaceAll(sheet, `"`, `\\"`) + `":`))
		if err != nil {
			return err
		}

		err = withJSONArrayOutWriter(w, func(w JSONArrayWriter) error {
			rows, err := in.GetRows(sheet)
			if err != nil {
				return err
			}
			return writeXLSXSheet(rows, w)
		})
		if err != nil {
			return err
		}

		_, err = w.Write([]byte(","))
		if err != nil {
			return err
		}
	}

	// Find current offset
	lastChar, err := w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	if lastChar > 1 {
		// Overwrite the last comma
		lastChar = lastChar - 1
	}

	_, err = w.WriteAt([]byte("}"), lastChar)
	if err != nil {
		return err
	}

	return w.Truncate(lastChar + 1)
}

func transformXLSXFile(in, out string) error {
	f, err := excelize.OpenFile(in)
	if err != nil {
		return err
	}

	return transformXLSX(f, out)
}

func getMimeType(fileName string, ct ContentTypeInfo) string {
	if ct.Type != "" {
		return ct.Type
	}

	switch filepath.Ext(fileName) {
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".parquet":
		return "parquet"
	}

	return ""
}

func evalFilePanel(project *ProjectState, pageIndex int, panel *PanelInfo) error {
	assumedType := getMimeType(panel.File.Name, panel.File.ContentTypeInfo)
	if assumedType == "" {
		return fmt.Errorf("Unknown type")
	}

	out := getPanelResultsFile(project.ProjectName, panel.Id)

	switch assumedType {
	case "application/json":
		return transformJSONFile(panel.File.Name, out)
	case "text/csv":
		return transformCSVFile(panel.File.Name, out)
	case "application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return transformXLSXFile(panel.File.Name, out)
	case "parquet":
		return transformParquetFile(panel.File.Name, out)
	}

	// TODO: Need to just copy it as a string instead of this
	return fmt.Errorf("Unsupported type " + assumedType)
}