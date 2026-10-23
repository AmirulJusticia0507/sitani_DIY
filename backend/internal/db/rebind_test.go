package db

import "testing"

func TestRebind(t *testing.T) {
	tests := []struct{ input, want string }{
		{`SELECT * FROM t WHERE a=? AND b='?' AND c="?"`, `SELECT * FROM t WHERE a=$1 AND b='?' AND c="?"`},
		{"SELECT ? -- ignore ?\n, ? /* ignore ? */", "SELECT $1 -- ignore ?\n, $2 /* ignore ? */"},
		{`INSERT INTO t VALUES ('it''s ?', ?)`, `INSERT INTO t VALUES ('it''s ?', $1)`},
	}
	for _, tt := range tests {
		if got := Rebind(tt.input); got != tt.want {
			t.Errorf("Rebind(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
