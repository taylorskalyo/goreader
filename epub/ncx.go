package epub

import (
	"encoding/xml"
	"io"
)

const idNCX = "ncx"

type NCX struct {
	NavPoints []NavPoint `xml:"navMap>navPoint"`
}

type NavPoint struct {
	ID        string `xml:"id,attr"`
	PlayOrder string `xml:"playOrder,attr"`
	NavLabel  struct {
		Text string `xml:"text"`
	} `xml:"navLabel"`
	Content struct {
		Src string `xml:"src,attr"`
	} `xml:"content"`
	NavPoints []NavPoint `xml:"navPoint,omitempty"`
}

// Load EPUB 2.0 compatible navigation documents.
func (r *Reader) setNCX() error {
	for _, rf := range r.Container.Rootfiles {
		for _, item := range rf.Manifest.Items {
			if item.ID == idNCX {
				f, err := item.f.Open()
				if err != nil {
					return err
				}
				defer f.Close()

				data, err := io.ReadAll(f)
				if err != nil {
					return err
				}

				err = xml.Unmarshal(data, &rf.NCX)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (rf Rootfile) ncxItemName(href string) string {
	for _, point := range rf.NCX.NavPoints {
		if label := point.lookupItemName(href); label != "" {
			return label
		}
	}

	return ""
}

func (np NavPoint) lookupItemName(href string) string {
	if np.Content.Src == href {
		return np.NavLabel.Text
	}

	for _, point := range np.NavPoints {
		if label := point.lookupItemName(href); label != "" {
			return label
		}
	}

	return ""
}
