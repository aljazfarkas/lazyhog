package utils

import "strings"

// FuzzyMatch performs a simple fuzzy match between a query and a target string
// Returns true if all characters in query appear in order in target (case-insensitive)
func FuzzyMatch(query, target string) bool {
	query = strings.ToLower(query)
	target = strings.ToLower(target)

	if query == "" {
		return true
	}

	queryIdx := 0
	for i := 0; i < len(target) && queryIdx < len(query); i++ {
		if target[i] == query[queryIdx] {
			queryIdx++
		}
	}

	return queryIdx == len(query)
}

// FuzzyScore calculates a fuzzy match score (higher is better)
// Returns 0 if no match, otherwise returns a score based on match quality
func FuzzyScore(query, target string) int {
	query = strings.ToLower(query)
	target = strings.ToLower(target)

	if query == "" {
		return 1
	}

	if !FuzzyMatch(query, target) {
		return 0
	}

	// Exact match gets highest score
	if query == target {
		return 1000
	}

	// Prefix match gets high score
	if strings.HasPrefix(target, query) {
		return 500
	}

	// Contains gets medium score
	if strings.Contains(target, query) {
		return 100
	}

	// Fuzzy match gets base score
	return 10
}
