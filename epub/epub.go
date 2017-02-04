package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"log"
)

const ContainerPath = "META-INF/container.xml"

type Reader struct {
	Container
	files map[string]*zip.File
}

type ReadCloser struct {
	Reader
	z *zip.ReadCloser
}

// Rootfile contains the location of a content.opf package file.
type Rootfile struct {
	FullPath string `xml:"full-path,attr"`
	Package
}

// Container serves as a directory of Rootfiles.
type Container struct {
	Rootfiles []*Rootfile `xml:"rootfiles>rootfile"`
}

// Package represents an epub content.opf file.
type Package struct {
	Metadata Metadata `xml:"metadata"`
	Manifest Manifest `xml:"manifest"`
	Spine    Spine    `xml:"spine"`
}

// Metadata contains publishing information about the epub.
type Metadata struct {
	Title       string `xml:"title"`
	Language    string `xml:"language"`
	Identifier  string `xml:"idenifier"`
	Creator     string `xml:"creator"`
	Contributor string `xml:"contributor"`
	Publisher   string `xml:"publisher"`
	Subject     string `xml:"subject"`
	Description string `xml:"description"`
	Event       []struct {
		Name string `xml:"event,attr"`
		Date string `xml:",innerxml"`
	} `xml:"date"`
	Type     string `xml:"type"`
	Format   string `xml:"format"`
	Source   string `xml:"source"`
	Relation string `xml:"relation"`
	Coverage string `xml:"coverage"`
	Rights   string `xml:"rights"`
}

// Manifest lists every file that is part of the epub.
type Manifest struct {
	Items []Item `xml:"item"`
}

// Item represents a file stored in the epub.
type Item struct {
	ID   string `xml:"id,attr"`
	HREF string `xml:"href,attr"`
	f    *zip.File
}

// Spine defines the reading order of the epub documents.
type Spine struct {
	Itemrefs []Itemref `xml:"itemref"`
}

// Itemref points to an Item.
type Itemref struct {
	IDREF string `xml:"idref,attr"`
	Item  *Item
}

// OpenReader will open the epub file specified by name and return a
// ReadCloser.
func OpenReader(name string) (*ReadCloser, error) {
	z, err := zip.OpenReader(name)
	if err != nil {
		return nil, err
	}

	rc := new(ReadCloser)
	rc.z = z
	if err = rc.init(&z.Reader); err != nil {
		return nil, err
	}
	return rc, nil
}

// NewReader returns a new Reader reading from r, which is assumed to have the
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

	r.setContainer()
	r.setPackages()
	r.setItems()
	for _, rf := range r.Container.Rootfiles {
		log.Printf("%+v\n", rf)
	}

	return nil
}

// setContainer unmarshals the epub's container.xml file.
func (r *Reader) setContainer() error {
	f, err := r.files[ContainerPath].Open()
	if err != nil {
		return err
	}

	var b bytes.Buffer
	_, err = io.Copy(&b, f)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(b.Bytes(), &r.Container)
	if err != nil {
		return err
	}

	return nil
}

// setPackages unmarshal's each of the epub's container.opf files.
func (r *Reader) setPackages() error {
	for _, rf := range r.Container.Rootfiles {
		f, err := r.files[rf.FullPath].Open()
		if err != nil {
			return err
		}

		var b bytes.Buffer
		_, err = io.Copy(&b, f)
		if err != nil {
			return err
		}

		err = xml.Unmarshal(b.Bytes(), &rf.Package)
		if err != nil {
			return err
		}
	}

	return nil
}

// setItems associates Itemrefs with their respective Item and Items with
// their zip.File.
func (r *Reader) setItems() {
	for _, rf := range r.Container.Rootfiles {
		itemMap := make(map[string]*Item)
		for i, item := range rf.Manifest.Items {
			rf.Manifest.Items[i].f = r.files[item.HREF]
			itemMap[item.ID] = &rf.Manifest.Items[i]
		}

		for i, itemref := range rf.Spine.Itemrefs {
			rf.Spine.Itemrefs[i].Item = itemMap[itemref.IDREF]
		}
	}
}

// Close closes the epub file, rendering it unusable for I/O.
func (rc *ReadCloser) Close() {
	rc.z.Close()
}
