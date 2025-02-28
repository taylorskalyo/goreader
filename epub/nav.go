package epub

import (
	"encoding/xml"
	"io"
)

const idTOC = "toc"

// NavDoc represents an EPUB 3.0 compatible navigation document.
type NavDoc struct {
	Nav []Nav `xml:"body>nav"`
}

// Nav represents a list of navigable items.
type Nav struct {
	Items []ListItem `xml:"ol>li"`
}

// ListItem represents a location within the epub file that can be navigated
// to.
type ListItem struct {
	Link struct {
		Href string `xml:"href,attr"`
		Text string `xml:",chardata"`
	} `xml:"a"`
	SubItems *[]ListItem `xml:"ol>li"`
}

// Load EPUB 3.0 compatible navigation documents.
func (r *Reader) setTOC() error {
	for _, rf := range r.Container.Rootfiles {
		for _, item := range rf.Manifest.Items {
			if item.ID == idTOC {
				f, err := item.f.Open()
				if err != nil {
					return err
				}
				defer f.Close()

				data, err := io.ReadAll(f)
				if err != nil {
					return err
				}

				err = xml.Unmarshal(data, &rf.NavDoc)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// navItemName searches for the name of an item in a NavDoc document.
func (rf Rootfile) navItemName(href string) string {
	for _, nav := range rf.NavDoc.Nav {
		for _, item := range nav.Items {
			if label := item.lookupItemName(href); label != "" {
				return label
			}
		}
	}

	return ""
}

// lookupItemName traverses a ListItem looking for the name of an item.
func (li ListItem) lookupItemName(href string) string {
	if li.Link.Href == href {
		return li.Link.Text
	}

	if li.SubItems != nil {
		for _, item := range *li.SubItems {
			if label := item.lookupItemName(href); label != "" {
				return label
			}
		}
	}

	return ""
}
