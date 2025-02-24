/*
Package epub provides basic support for reading EPUB archives.
*/
package epub

import (
	"archive/zip"
	"errors"
	"io"
	"os"
)

var (
	// ErrNoRootfile occurs when there are no rootfile entries found in
	// container.xml.
	ErrNoRootfile = errors.New("epub: no rootfile found in container")

	// ErrBadRootfile occurs when container.xml references a rootfile that does
	// not exist in the zip.
	ErrBadRootfile = errors.New("epub: container references non-existent rootfile")

	// ErrNoItemref occurrs when a content.opf contains a spine without any
	// itemref entries.
	ErrNoItemref = errors.New("epub: no itemrefs found in spine")

	// ErrBadItemref occurs when an itemref entry in content.opf references an
	// item that does not exist in the manifest.
	ErrBadItemref = errors.New("epub: itemref references non-existent item")

	// ErrBadManifest occurs when a manifest in content.opf references an item
	// that does not exist in the zip.
	ErrBadManifest = errors.New("epub: manifest references non-existent item")
)

// Reader represents a readable epub file.
type Reader struct {
	Container
	files map[string]*zip.File
}

// ReadCloser represents a readable epub file that can be closed.
type ReadCloser struct {
	Reader
	f *os.File
}

// OpenReader will open the epub file specified by name and return a
// ReadCloser.
func OpenReader(name string) (*ReadCloser, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	rc := new(ReadCloser)
	rc.f = f

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	z, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return nil, err
	}

	if err = rc.init(z); err != nil {
		return nil, err
	}

	return rc, nil
}

// NewReader returns a new Reader reading from ra, which is assumed to have the
// given size in bytes.
func NewReader(ra io.ReaderAt, size int64) (*Reader, error) {
	z, err := zip.NewReader(ra, size)
	if err != nil {
		return nil, err
	}

	r := new(Reader)
	if err = r.init(z); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Reader) init(z *zip.Reader) error {
	// Create a file lookup table
	r.files = make(map[string]*zip.File)
	for _, f := range z.File {
		r.files[f.Name] = f
	}

	err := r.setContainer()
	if err != nil {
		return err
	}

	err = r.setPackages()
	if err != nil {
		return err
	}

	err = r.setItems()
	if err != nil {
		return err
	}

	err = r.setNCX()
	if err != nil {
		return err
	}

	err = r.setTOC()
	if err != nil {
		return err
	}

	return nil
}

// DefaultRendition selects the default rendition from a list of rootfiles of
// an epub container.
func (c *Container) DefaultRendition() *Rootfile {
	if len(c.Rootfiles) < 1 {
		return nil
	}

	// An epub file may contain multilpe renditions. In practice, there is often
	// just one. For simplicity, select the first available.
	return c.Rootfiles[0]
}
