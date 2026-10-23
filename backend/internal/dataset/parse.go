// Package dataset memvalidasi dan membaca file CSV/XLSX menjadi baris
// bertajuk kolom untuk repository data lokal.
package dataset

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/xuri/excelize/v2"
)

const (
	MaksBytes            = 10 << 20
	MaksBaris            = 10000
	MaksKolom            = 100
	MaksHeader           = 128
	MaksXLSXUncompressed = 64 << 20
)

var ErrUnsupported = errors.New("format file harus CSV atau XLSX")

type Result struct {
	Headers []string            `json:"headers"`
	Rows    []map[string]string `json:"rows"`
}

// Parse menerima CSV UTF-8 atau XLSX. XLSX hanya membaca sheet pertama,
// baris pertama adalah header, dan file tidak boleh melebihi limit.
func Parse(filename string, raw []byte) (Result, error) {
	if len(raw) == 0 {
		return Result{}, errors.New("file kosong")
	}
	if len(raw) > MaksBytes {
		return Result{}, fmt.Errorf("file melebihi batas %d MiB", MaksBytes>>20)
	}
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".csv"):
		return parseCSV(raw)
	case strings.HasSuffix(lower, ".xlsx"):
		return parseXLSX(raw)
	default:
		return Result{}, ErrUnsupported
	}
}

func parseCSV(raw []byte) (Result, error) {
	r := csv.NewReader(bytes.NewReader(raw))
	r.FieldsPerRecord = -1
	r.ReuseRecord = false
	var records [][]string
	for {
		row, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return Result{}, fmt.Errorf("CSV: %w", err)
		}
		records = append(records, row)
		if len(records) > MaksBaris+1 {
			return Result{}, fmt.Errorf("jumlah baris melebihi batas %d", MaksBaris)
		}
	}
	return recordsToResult(records)
}

func parseXLSX(raw []byte) (Result, error) {
	if len(raw) < 4 || !bytes.Equal(raw[:4], []byte{'P', 'K', 3, 4}) {
		return Result{}, errors.New("file XLSX tidak valid")
	}
	f, err := excelize.OpenReader(bytes.NewReader(raw), excelize.Options{
		UnzipSizeLimit:    MaksXLSXUncompressed,
		UnzipXMLSizeLimit: MaksXLSXUncompressed,
	})
	if err != nil {
		return Result{}, fmt.Errorf("baca XLSX: %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return Result{}, errors.New("workbook tidak memiliki sheet")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return Result{}, fmt.Errorf("baca sheet pertama: %w", err)
	}
	if len(rows) > MaksBaris+1 {
		return Result{}, fmt.Errorf("jumlah baris melebihi batas %d", MaksBaris)
	}
	return recordsToResult(rows)
}

func recordsToResult(records [][]string) (Result, error) {
	if len(records) == 0 {
		return Result{}, errors.New("file tidak memiliki header")
	}
	if len(records[0]) == 0 || len(records[0]) > MaksKolom {
		return Result{}, fmt.Errorf("jumlah kolom header harus 1 sampai %d", MaksKolom)
	}
	headers := make([]string, len(records[0]))
	seen := make(map[string]struct{}, len(headers))
	for i, raw := range records[0] {
		header := strings.TrimSpace(strings.TrimPrefix(raw, "\ufeff"))
		if header == "" || len([]rune(header)) > MaksHeader {
			return Result{}, fmt.Errorf("header kolom %d kosong atau melebihi %d karakter", i+1, MaksHeader)
		}
		key := strings.ToLower(header)
		if _, ok := seen[key]; ok {
			return Result{}, fmt.Errorf("header kolom duplikat: %q", header)
		}
		for _, r := range header {
			if unicode.IsControl(r) {
				return Result{}, fmt.Errorf("header kolom %d mengandung karakter kontrol", i+1)
			}
		}
		seen[key] = struct{}{}
		headers[i] = header
	}

	result := Result{Headers: headers, Rows: make([]map[string]string, 0, len(records)-1)}
	for i, values := range records[1:] {
		if len(values) > len(headers) {
			return Result{}, fmt.Errorf("baris %d memiliki lebih banyak kolom daripada header", i+2)
		}
		row := make(map[string]string, len(headers))
		nonEmpty := false
		for col, header := range headers {
			value := ""
			if col < len(values) {
				value = strings.TrimSpace(values[col])
			}
			if value != "" {
				nonEmpty = true
			}
			row[header] = value
		}
		if nonEmpty {
			result.Rows = append(result.Rows, row)
		}
	}
	return result, nil
}
