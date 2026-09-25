package safermapp

import (
	"slices"
	"testing"
)

func TestBuildRemovalPlanMatchesNameAndMD5(t *testing.T) {
	key := fileKey{name: "photo.jpg", md5: "0123456789abcdef0123456789abcdef"}
	referenceRecords := []fileRecord{
		{path: "/reference-a/photo.jpg", manifestDir: "/reference-a"},
		{path: "/reference-b/photo.jpg", manifestDir: "/reference-b"},
	}
	targetRecord := fileRecord{
		path:        "/target/photo.jpg",
		manifestDir: "/target",
	}

	plan := buildRemovalPlan(
		fileIndex{key: referenceRecords},
		fileIndex{key: {targetRecord}},
	)

	if len(plan) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(plan))
	}

	got := plan[0]
	if got.key != key {
		t.Errorf("candidate key = %+v, want %+v", got.key, key)
	}
	if got.target != targetRecord {
		t.Errorf("candidate target = %+v, want %+v", got.target, targetRecord)
	}
	if !slices.Equal(got.references, referenceRecords) {
		t.Errorf("candidate references = %+v, want %+v", got.references, referenceRecords)
	}
}

func TestBuildRemovalPlanRequiresMatchingNameAndMD5(t *testing.T) {
	const sum = "0123456789abcdef0123456789abcdef"

	tests := []struct {
		name       string
		reference  fileKey
		target     fileKey
		wantLength int
	}{
		{
			name:       "both match",
			reference:  fileKey{name: "photo.jpg", md5: sum},
			target:     fileKey{name: "photo.jpg", md5: sum},
			wantLength: 1,
		},
		{
			name:       "different name",
			reference:  fileKey{name: "photo.jpg", md5: sum},
			target:     fileKey{name: "copy.jpg", md5: sum},
			wantLength: 0,
		},
		{
			name:       "different MD5",
			reference:  fileKey{name: "photo.jpg", md5: sum},
			target:     fileKey{name: "photo.jpg", md5: "abcdef0123456789abcdef0123456789"},
			wantLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			references := fileIndex{
				tt.reference: {{path: "/reference/file"}},
			}
			targets := fileIndex{
				tt.target: {{path: "/target/file"}},
			}

			plan := buildRemovalPlan(references, targets)
			if len(plan) != tt.wantLength {
				t.Fatalf("plan length = %d, want %d", len(plan), tt.wantLength)
			}
		})
	}
}

func TestBuildRemovalPlanCreatesOneCandidatePerTarget(t *testing.T) {
	key := fileKey{name: "photo.jpg", md5: "0123456789abcdef0123456789abcdef"}
	targets := []fileRecord{
		{path: "/target-a/photo.jpg"},
		{path: "/target-b/photo.jpg"},
	}

	plan := buildRemovalPlan(
		fileIndex{key: {{path: "/reference/photo.jpg"}}},
		fileIndex{key: targets},
	)

	if len(plan) != len(targets) {
		t.Fatalf("plan length = %d, want %d", len(plan), len(targets))
	}

	gotTargets := []fileRecord{plan[0].target, plan[1].target}
	if !slices.Equal(gotTargets, targets) {
		t.Errorf("candidate targets = %+v, want %+v", gotTargets, targets)
	}
}

func TestBuildRemovalPlanWithEmptyIndexes(t *testing.T) {
	if plan := buildRemovalPlan(fileIndex{}, fileIndex{}); len(plan) != 0 {
		t.Fatalf("plan length = %d, want 0", len(plan))
	}
}
