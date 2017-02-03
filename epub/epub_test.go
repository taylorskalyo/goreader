package epub

import "testing"

func TestOpenReader(t *testing.T) {
	r, err := OpenReader("_test_files/alice.epub")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
}
