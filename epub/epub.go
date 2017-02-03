package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"log"
)

type Reader struct {
	Contents []Content
}

type ReadCloser struct {
	z *zip.ReadCloser
	Reader
}

type Chapter struct {
	Title string
	f     zip.File
}

type rootfile struct {
	FullPath string `xml:"full-path,attr"`
}

type container struct {
	XMLName   xml.Name   `xml:"container"`
	Rootfiles []rootfile `xml:"rootfiles>rootfile"`
}

type Content struct {
	XMLName xml.Name `xml:"package"`
	Metadata
	Manifest
	Spine
}

type Metadata struct {
}

type Manifest struct {
	Items []item `xml:"manifest>item"`
}

type item struct {
	ID   string `xml:"id,attr"`
	HREF string `xml:"href,attr"`
}

type Spine struct {
	ItemRefs []itemref `xml:"spine>itemref"`
}

type itemref struct {
	IDREF string `xml:"idref,attr"`
}

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
	files := make(map[string]*zip.File)
	for _, f := range z.File {
		files[f.Name] = f
	}

	r.setContents(files)
	log.Printf("%+v\n", r.Contents)

	return nil
}

func (r *Reader) rootfiles(files map[string]*zip.File) (rf []rootfile, err error) {
	f, err := files["META-INF/container.xml"].Open()
	if err != nil {
		return
	}

	var b bytes.Buffer
	_, err = io.Copy(&b, f)
	if err != nil {
		return
	}

	var c container
	err = xml.Unmarshal(b.Bytes(), &c)
	if err != nil {
		return
	}
	log.Printf("%+v\n", c)

	return c.Rootfiles, nil
}

func (r *Reader) setContents(files map[string]*zip.File) error {
	rfs, err := r.rootfiles(files)
	if err != nil {
		return err
	}

	for _, rf := range rfs {
		f, err := files[rf.FullPath].Open()
		if err != nil {
			return err
		}

		var b bytes.Buffer
		_, err = io.Copy(&b, f)
		if err != nil {
			return err
		}

		var c Content
		err = xml.Unmarshal(b.Bytes(), &c)
		if err != nil {
			return err
		}

		r.Contents = append(r.Contents, c)
	}

	return nil
}

func (rc *ReadCloser) Close() {
	rc.z.Close()
}

func (c *Chapter) Open() (io.ReadCloser, error) {
	// TODO pull out just the text
	return c.f.Open()
}
