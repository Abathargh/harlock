package object

import "testing"

func TestStringHashKey(t *testing.T) {
	firstA := &String{"first string"}
	firstB := &String{"first string"}
	secondA := &String{"second string"}
	secondB := &String{"second string"}

	if firstA.HashKey() != firstB.HashKey() {
		t.Errorf("expected same strings to have same hashes")
	}
	if secondA.HashKey() != secondB.HashKey() {
		t.Errorf("expected same strings to have same hashes")
	}
	if firstA.HashKey() == secondA.HashKey() {
		t.Errorf("expected different strings to have different hashes")
	}
}
