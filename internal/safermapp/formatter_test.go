package safermapp

import (
	"errors"
	"strings"
	"testing"
)

func TestWriteRemovalPlan(t *testing.T) {
	const (
		sumA = "0123456789abcdef0123456789abcdef"
		sumB = "abcdef0123456789abcdef0123456789"
	)

	plan := []candidate{
		{
			key:    fileKey{name: "b.txt", md5: sumB},
			target: fileRecord{path: "/target/b.txt"},
			references: []fileRecord{
				{path: "/reference-z/b.txt"},
				{path: "/reference-a/b.txt"},
			},
		},
		{
			key:        fileKey{name: "a.txt", md5: sumA},
			target:     fileRecord{path: "/target/a.txt"},
			references: []fileRecord{{path: "/reference/a.txt"}},
		},
	}

	var output strings.Builder
	if err := writeRemovalPlan(&output, plan); err != nil {
		t.Fatalf("writeRemovalPlan returned an error: %v", err)
	}

	want := "Removal plan: 2 candidate(s)\n" +
		"\n[1] Target: \"/target/a.txt\"\n" +
		"    MD5: " + sumA + "\n" +
		"    References (1):\n" +
		"        - \"/reference/a.txt\"\n" +
		"\n[2] Target: \"/target/b.txt\"\n" +
		"    MD5: " + sumB + "\n" +
		"    References (2):\n" +
		"        - \"/reference-a/b.txt\"\n" +
		"        - \"/reference-z/b.txt\"\n"

	if output.String() != want {
		t.Errorf("output:\n%s\nwant:\n%s", output.String(), want)
	}

	if plan[0].target.path != "/target/b.txt" || plan[1].target.path != "/target/a.txt" {
		t.Errorf("writeRemovalPlan reordered the input plan: %+v", plan)
	}
	if plan[0].references[0].path != "/reference-z/b.txt" {
		t.Errorf("writeRemovalPlan reordered input references: %+v", plan[0].references)
	}
}

func TestWriteRemovalPlanEmpty(t *testing.T) {
	var output strings.Builder
	if err := writeRemovalPlan(&output, nil); err != nil {
		t.Fatalf("writeRemovalPlan returned an error: %v", err)
	}

	want := "Removal plan: 0 candidate(s)\nNo removal candidates found\n"
	if output.String() != want {
		t.Errorf("output = %q, want %q", output.String(), want)
	}
}

func TestWriteRemovalPlanReturnsWriterError(t *testing.T) {
	err := writeRemovalPlan(failingWriter{}, []candidate{
		{
			key:        fileKey{name: "photo.jpg", md5: testMD5},
			target:     fileRecord{path: "/target/photo.jpg"},
			references: []fileRecord{{path: "/reference/photo.jpg"}},
		},
	})

	if !errors.Is(err, errWriteFailed) {
		t.Fatalf("writeRemovalPlan error = %v, want %v", err, errWriteFailed)
	}
}

var errWriteFailed = errors.New("write failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}
