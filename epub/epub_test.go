package epub

import (
	"os"
	"testing"
)

const expFormat = "Expected: %v, but got: %v\n"

type containerTest struct {
	*testing.T
	c Container
}

func TestOpenReader(t *testing.T) {
	r, err := OpenReader("_test_files/alice.epub")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	t.Run("ReadCloser", func(t *testing.T) {
		tt := containerTest{t, r.Container}
		tt.TestContainer()
	})
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
	r, err := NewReader(rc, fi.Size())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Reader", func(t *testing.T) {
		tt := containerTest{t, r.Container}
		tt.TestContainer()
	})
}

func (ct *containerTest) TestContainer() {
	ct.Run("Container", func(t *testing.T) {
		tt := containerTest{t, ct.c}
		tt.TestMetadata()
		tt.TestSpine()
		tt.TestManifest()
	})
}

func (ct *containerTest) TestMetadata() {
	meta := ct.c.Rootfiles[0].Metadata

	exp := "Alice's Adventures in Wonderland / Illustrated by Arthur Rackham. With a Proem by Austin Dobson"
	if meta.Title != exp {
		ct.Errorf(expFormat, exp, meta.Title)
	}

	exp = "Lewis Carroll"
	if meta.Creator != exp {
		ct.Errorf(expFormat, exp, meta.Creator)
	}
}

func (ct *containerTest) TestSpine() {
	testCases := []struct {
		itemrefIndex int
		expIDREF     string
	}{
		{0, "coverpage-wrapper"},
		{1, "item41"},
	}

	spine := ct.c.Rootfiles[0].Spine
	for _, tc := range testCases {
		ct.Run("Item", func(t *testing.T) {
			itemref := spine.Itemrefs[tc.itemrefIndex]
			if itemref.IDREF != tc.expIDREF {
				t.Errorf(expFormat, tc.expIDREF, itemref.IDREF)
			}

			if itemref.Item == nil {
				t.Errorf(expFormat, "not nil", "nil")
			} else if itemref.Item.ID != tc.expIDREF {
				t.Errorf(expFormat, tc.expIDREF, itemref.Item.ID)
			}
		})
	}
}

func (ct *containerTest) TestManifest() {
	testCases := []struct {
		itemIndex int
		expID     string
		expHREF   string
	}{
		{
			40,
			"item41",
			"@public@vhost@g@gutenberg@html@files@28885@28885-h@28885-h-0.htm.html",
		},
		{
			0,
			"item1",
			"@public@vhost@g@gutenberg@html@files@28885@28885-h@images@cover.jpg",
		},
	}

	manifest := ct.c.Rootfiles[0].Manifest
	for _, tc := range testCases {
		ct.Run("Item", func(t *testing.T) {
			item := manifest.Items[tc.itemIndex]

			if item.ID != tc.expID {
				t.Errorf(expFormat, tc.expID, item.ID)
			}

			if item.HREF != tc.expHREF {
				t.Errorf(expFormat, tc.expHREF, item.HREF)
			}
		})
	}
}
