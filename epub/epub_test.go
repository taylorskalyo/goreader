package epub

import (
	"os"
	"testing"
)

func TestOpenReader(t *testing.T) {
	r, err := OpenReader("_test_files/alice.epub")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
}

func TestNewReader(t *testing.T) {
	rc, err := os.Open("_test_files/alice.epub")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	fi, err := rc.Stat()
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewReader(rc, fi.Size())
	if err != nil {
		t.Fatal(err)
	}
}

func TestMetadata(t *testing.T) {
	r, err := OpenReader("_test_files/alice.epub")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	meta := r.Container.Rootfiles[0].Metadata
	if meta.Title != "Alice's Adventures in Wonderland / Illustrated by Arthur Rackham. With a Proem by Austin Dobson" {
		t.Fatalf("Unexpected title: %s\n", meta.Title)
	}
	if meta.Creator != "Lewis Carroll" {
		t.Fatalf("Unexpected creator: %s\n", meta.Creator)
	}
}

func TestSpine(t *testing.T) {
	r, err := OpenReader("_test_files/alice.epub")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	spine := r.Container.Rootfiles[0].Spine
	if spine.Itemrefs[0].IDREF != "coverpage-wrapper" {
		t.Fatalf("Unexpected IDREF: %s\n", spine.Itemrefs[0].IDREF)
	}

	if spine.Itemrefs[0].Item.ID != "coverpage-wrapper" {
		t.Fatalf("Unexpected ID: %s\n", spine.Itemrefs[0].Item.ID)
	}
}

func TestManifest(t *testing.T) {
	r, err := OpenReader("_test_files/alice.epub")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	manifest := r.Container.Rootfiles[0].Manifest
	if manifest.Items[0].ID != "item1" {
		t.Fatalf("Unexpected ID: %s\n", manifest.Items[0].ID)
	}

	if manifest.Items[0].HREF != "@public@vhost@g@gutenberg@html@files@28885@28885-h@images@cover.jpg" {
		t.Fatalf("Unexpected HREF: %s\n", manifest.Items[0].HREF)
	}
}
