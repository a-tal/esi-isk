package db

import (
	"database/sql"
	"testing"
)

func TestGetPattern(t *testing.T) {
	d := "the default"

	s1 := sql.NullString{
		Valid:  false,
		String: "something",
	}
	p1 := getPattern(s1, d)
	if p1 != d {
		t.Errorf("invalid pattern. received %q, expected %q", p1, d)
	}

	s2 := sql.NullString{
		Valid:  true,
		String: "",
	}
	p2 := getPattern(s2, d)
	if p2 != d {
		t.Errorf("invalid pattern. received %q, expected %q", p2, d)
	}

	s3 := sql.NullString{
		Valid: true,
		String: "		",
	}
	p3 := getPattern(s3, d)
	if p3 != d {
		t.Errorf("invalid pattern. received %q, expected %q", p3, d)
	}

	s4 := sql.NullString{
		Valid:  true,
		String: " ",
	}
	p4 := getPattern(s4, d)
	if p4 != d {
		t.Errorf("invalid pattern. received %q, expected %q", p4, d)
	}

	s5 := sql.NullString{
		Valid: true,
		String: "		words    ",
	}
	p5 := getPattern(s5, d)
	if p5 != s5.String {
		t.Errorf("invalid pattern. received %q, expected %q", p4, s5.String)
	}
}
