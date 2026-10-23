package dataset

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestParseCSV(t *testing.T) {
	got, err := Parse("ekspor.csv", []byte("\ufeffNama,Jumlah\r\nPadi,10\r\nJagung, 20 \r\n,\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Headers) != 2 || got.Headers[0] != "Nama" || len(got.Rows) != 2 || got.Rows[1]["Jumlah"] != "20" {
		t.Fatalf("hasil parse tak sesuai: %#v", got)
	}
}

func TestParseRejectsDuplicateHeaders(t *testing.T) {
	if _, err := Parse("x.csv", []byte("Nama,nama\na,b\n")); err == nil {
		t.Fatal("header duplikat semestinya ditolak")
	}
}

func TestParseRejectsUnsupportedAndOversized(t *testing.T) {
	if _, err := Parse("x.xls", []byte("x")); err != ErrUnsupported {
		t.Fatalf("error format = %v", err)
	}
	if _, err := Parse("x.csv", bytes.Repeat([]byte("a"), MaksBytes+1)); err == nil {
		t.Fatal("file oversized semestinya ditolak")
	}
}

func TestParseXLSX(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, body := range map[string]string{
		"[Content_Types].xml":        `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`,
		"_rels/.rels":                `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`,
		"xl/workbook.xml":            `<?xml version="1.0"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Data" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml":   `<?xml version="1.0"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>Nama</t></is></c><c r="B1" t="inlineStr"><is><t>Jumlah</t></is></c></row><row r="2"><c r="A2" t="inlineStr"><is><t>Padi</t></is></c><c r="B2"><v>12</v></c></row></sheetData></worksheet>`,
	} {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := Parse("data.xlsx", buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) != 1 || got.Rows[0]["Nama"] != "Padi" || got.Rows[0]["Jumlah"] != "12" {
		t.Fatalf("hasil XLSX tak sesuai: %#v", got)
	}
}
