package safermapp

import (
	"fmt"
	"slices"
	"strings"
)

func printRemovalPlan(plan []candidate) {
	slices.SortFunc(plan, func(a, b candidate) int {
		return strings.Compare(a.target.path, b.target.path)
	})

	for _, candidate := range plan {
		printCandidate(candidate)
	}
}

func printCandidate(candidate candidate) {
	fmt.Printf("%s: %s\n", candidate.target.path, candidate.key.md5)
}
