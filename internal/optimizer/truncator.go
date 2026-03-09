package optimizer

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

// Truncator handles token-aware truncation of text
type Truncator struct {
	tkm *tiktoken.Tiktoken
}

// NewTruncator creates a new truncator for a specific model (e.g., "gpt-4")
func NewTruncator(model string) (*Truncator, error) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		// Fallback to cl100k_base if model not found
		tkm, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return nil, fmt.Errorf("failed to get encoding: %w", err)
		}
	}
	return &Truncator{tkm: tkm}, nil
}

// CountTokens returns the number of tokens in a string
func (t *Truncator) CountTokens(text string) int {
	return len(t.tkm.Encode(text, nil, nil))
}

// Truncate ensures text fits within the maxTokens budget
func (t *Truncator) Truncate(text string, maxTokens int) string {
	tokens := t.tkm.Encode(text, nil, nil)
	if len(tokens) <= maxTokens {
		return text
	}

	// Truncate to maxTokens
	truncatedTokens := tokens[:maxTokens]
	return t.tkm.Decode(truncatedTokens)
}

// TruncateResults ensures a list of result snippets fits within a total budget
func (t *Truncator) TruncateResults(snippets []string, totalBudget int) []string {
	var results []string
	currentBudget := totalBudget

	for _, s := range snippets {
		if currentBudget <= 0 {
			break
		}

		sTokens := t.CountTokens(s)
		if sTokens <= currentBudget {
			results = append(results, s)
			currentBudget -= sTokens
		} else {
			// If a single snippet is too large, truncate it and then stop
			truncated := t.Truncate(s, currentBudget)
			results = append(results, truncated+"...")
			break
		}
	}

	return results
}
