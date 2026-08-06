package s3

import "testing"

func TestSanitizeS3Folder(t *testing.T) {
	got, err := sanitizeS3Folder("uploads/docs")
	if err != nil || got != "uploads/docs" {
		t.Fatalf("got %q err %v", got, err)
	}
	_, err = sanitizeS3Folder("uploads/../secret")
	if err == nil {
		t.Fatal("expected error for traversal")
	}
}
