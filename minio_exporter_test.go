package main

import "testing"

func TestNewMinioExporter(t *testing.T) {
	cases := []struct {
		uri string
		ok  bool
	}{
		{uri: "", ok: false},
		{uri: "localhost:9000", ok: true},
		{uri: "https://localhost:9000", ok: true},
		{uri: "http://some.where:9000", ok: true},
		{uri: "//localhost:9000", ok: false},
		{uri: "wrong://localhost:9000", ok: false},
	}

	for _, test := range cases {
		_, err := NewMinioExporter(test.uri, "example_key", "example_secret", false)
		if test.ok && err != nil {
			t.Errorf("expected no error w/ %q, but got %q", test.uri, err)
		}
		if !test.ok && err == nil {
			t.Errorf("expected error w/ %q, but got %q", test.uri, err)
		}
	}
}
