package main

import "fmt"

// FilterStrings takes a slice of strings and a predicate function.
// It returns a new slice containing only the strings for which the predicate returns true.
func FilterStrings(slice []string, predicate func(string) bool) []string {
	var result []string
	for _, str := range slice {
		if predicate(str) {
			result = append(result, str)
		}
	}
	return result
}

func main() {
	// Example usage:
	words := []string{"apple", "banana", "cherry", "date", "elderberry"}

	// Filter words that start with 'a'
	filteredWords := FilterStrings(words, func(s string) bool {
		return len(s) > 5
	})

	fmt.Println("Original words:", words)
	fmt.Println("Words longer than 5 characters:", filteredWords)

	// Another example: filter words containing 'e'
	wordsWithE := FilterStrings(words, func(s string) bool {
		for _, r := range s {
			if r == 'e' {
				return true
			}
		}
		return false
	})

	fmt.Println("Words containing 'e':", wordsWithE)
}
