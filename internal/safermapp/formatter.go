package safermapp

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"
)

func writeRemovalPlan(w io.Writer, plan []candidate) error {
	// do not update external plan
	ordered := slices.Clone(plan)
	slices.SortFunc(ordered, func(a, b candidate) int {
		return strings.Compare(a.target.path, b.target.path)
	})

	output := bufio.NewWriter(w)

	fmt.Fprintf(output, "Removal plan: %d candidate(s)\n", len(ordered))
	if len(ordered) == 0 {
		fmt.Fprintln(output, "No removal candidates found")
		return output.Flush()
	}

	for i, candidate := range ordered {
		references := slices.Clone(candidate.references)
		slices.SortFunc(references, func(a, b fileRecord) int {
			return strings.Compare(a.path, b.path)
		})

		fmt.Fprintf(output, "\n[%d] Target: %q\n", i+1, candidate.target.path)
		fmt.Fprintf(output, "    MD5: %s\n", candidate.key.md5)
		fmt.Fprintf(output, "    References (%d):\n", len(references))
		for _, ref := range references {
			fmt.Fprintf(output, "        - %q\n", ref.path)
		}
	}

	return output.Flush()
}
