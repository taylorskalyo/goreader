package epub

import (
	"encoding/xml"
	"io"
)

const containerPath = "META-INF/container.xml"

// Rootfile contains the location of an epub .opf file.
type Rootfile struct {
	FullPath string `xml:"full-path,attr"`
	Package
	NCX
	NavDoc
}

// Container serves as a directory of Rootfiles.
type Container struct {
	Rootfiles []*Rootfile `xml:"rootfiles>rootfile"`
}

// setContainer unmarshals the epub's container.xml file.
func (r *Reader) setContainer() error {
	f, err := r.files[containerPath].Open()
	if err != nil {
		return err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(data, &r.Container)
	if err != nil {
		return err
	}

	if len(r.Container.Rootfiles) < 1 {
		return ErrNoRootfile
	}

	return nil
}

// ItemName attempts to find a name for the given item reference.
func (rf Rootfile) ItemName(href string) string {
	// EPUB 3.0 compatible.
	if label := rf.navItemName(href); label != "" {
		return label
	}

	// EPUB 2.0 compatible.
	return rf.ncxItemName(href)
}
