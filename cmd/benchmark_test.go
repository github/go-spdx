package main

import (
	"testing"

	"github.com/github/go-spdx/v2/spdxexp"
)
func BenchmarkValidateLicensesMIT(b *testing.B) {
	b.ReportAllocs()

	licenses := []string{"MIT"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valid, invalid := spdxexp.ValidateLicenses(licenses)
		if !valid || len(invalid) != 0 {
			b.Fatalf("expected MIT to be valid; valid=%v invalid=%v", valid, invalid)
		}
	}
}

func BenchmarkStringEqualityMIT(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if "MIT" != "MIT" {
			b.Fatal("unexpected string inequality")
		}
	}
}
