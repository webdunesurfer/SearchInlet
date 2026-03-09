package optimizer

import (
	"testing"
)

func TestTruncator_CountTokens(t *testing.T) {
	trunc, err := NewTruncator("gpt-4")
	if err != nil {
		t.Fatalf("Failed to create truncator: %v", err)
	}

	text := "Hello, world!"
	count := trunc.CountTokens(text)
	if count <= 0 {
		t.Errorf("Expected positive token count, got %d", count)
	}
}

func TestTruncator_Truncate(t *testing.T) {
	trunc, err := NewTruncator("gpt-4")
	if err != nil {
		t.Fatalf("Failed to create truncator: %v", err)
	}

	text := "This is a longer sentence that we will truncate to a smaller token budget."
	totalTokens := trunc.CountTokens(text)
	
	maxTokens := totalTokens / 2
	truncated := trunc.Truncate(text, maxTokens)
	
	newCount := trunc.CountTokens(truncated)
	if newCount > maxTokens {
		t.Errorf("Expected at most %d tokens, got %d", maxTokens, newCount)
	}
}

func TestTruncator_TruncateResults(t *testing.T) {
	trunc, err := NewTruncator("gpt-4")
	if err != nil {
		t.Fatalf("Failed to create truncator: %v", err)
	}

	snippets := []string{
		"Short snippet 1",
		"This is a much longer snippet that might exceed the budget on its own.",
		"Snippet 3",
	}

	// Calculate tokens of first snippet to set a tight budget
	s1Tokens := trunc.CountTokens(snippets[0])
	
	// Budget that only allows first snippet and part of second
	budget := s1Tokens + 5
	
	results := trunc.TruncateResults(snippets, budget)
	
	if len(results) != 2 {
		t.Errorf("Expected 2 result snippets, got %d", len(results))
	}
	
	totalUsed := 0
	for _, r := range results {
		totalUsed += trunc.CountTokens(r)
	}
	
	// Total used should be around the budget (might be slightly over due to "..." suffix)
	if totalUsed > budget + 5 { // +5 for leeway with "..."
		t.Errorf("Total used tokens %d exceeded budget %d significantly", totalUsed, budget)
	}
}
