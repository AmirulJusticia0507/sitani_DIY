package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitStatements(t *testing.T) {
	input := `-- komentar baris dipindai keluar
CREATE TABLE a (nama VARCHAR(10) NOT NULL); -- komentar di belakang
/* komentar blok; dengan titik koma */
INSERT INTO a (nama) VALUES ('bei; kolom');;
`
	got, err := splitStatements(input)
	if err != nil {
		t.Fatalf("splitStatements error: %v", err)
	}
	want := []string{
		"CREATE TABLE a (nama VARCHAR(10) NOT NULL)",
		"INSERT INTO a (nama) VALUES ('bei; kolom')",
	}
	if len(got) != len(want) {
		t.Fatalf("jumlah pernyataan = %d, mau %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if strings.TrimSpace(got[i]) != want[i] {
			t.Errorf("pernyataan %d = %q, mau %q", i, got[i], want[i])
		}
	}
}

func TestSplitStatementsStringEscapes(t *testing.T) {
	got, err := splitStatements(`INSERT INTO t (k) VALUES ('dia'';s ja'); SELECT 1;`)
	if err != nil {
		t.Fatalf("splitStatements error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("jumlah pernyataan = %d, mau 2: %#v", len(got), got)
	}
	if !strings.Contains(got[0], "dia'';s ja") {
		t.Errorf("literal boleh mengandung titik koma, hasil %q", got[0])
	}
}

func TestSplitStatementsUnclosedString(t *testing.T) {
	if _, err := splitStatements(`INSERT INTO t (k) VALUES ('belum tutup`); err == nil {
		t.Fatal("string literal yang tidak ditutup harus mengembalikan error")
	}
}

// Migrasi proyek harus selalu dapat dipecah tanpa error dan menghasilkan
// pernyataan, supaya runner tidak gagal di tengah pemasangan skema.
func TestSplitStatementsProjectMigrations(t *testing.T) {
	dir := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("baca migrations: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("baca %s: %v", entry.Name(), err)
		}
		stmts, err := splitStatements(string(content))
		if err != nil {
			t.Errorf("splitStatements(%s) error: %v", entry.Name(), err)
			continue
		}
		if len(stmts) == 0 {
			t.Errorf("%s menghasilkan 0 pernyataan", entry.Name())
		}
	}
}
