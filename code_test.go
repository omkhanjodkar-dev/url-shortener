package main

import "testing"

func TestGenerateCode(t *testing.T) {
	code := generateCode()
	if len(code) != 6 {
		t.Errorf("expected code length 6, got %d", len(code))
	}
}