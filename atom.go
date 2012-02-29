package blogplus

import (
	"encoding/xml"
)

type AtomFeed struct {
	XMLName    xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	Id         string   `xml:"id"`
	Title      string   `xml:"title"`
	Link       []AtomLink
	Updated    string `xml:"updated"`
	AuthorName string `xml:"author>name"`
	AuthorUri  string `xml:"author>uri"`
	Logo       string `xml:"logo"`
	Entries    []AtomEntry
}

type AtomLink struct {
	XMLName xml.Name `xml:"link"`
	Href    string   `xml:"href,attr"`
	Rel     string   `xml:"rel,attr,omitempty"`
}

type AtomEntry struct {
	XMLName   xml.Name `xml:"entry"`
	Id        string   `xml:"id"`
	Link      AtomLink
	Title     string `xml:"title"`
	Content   AtomContent
	Summary   string `xml:"summary"`
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
}

type AtomContent struct {
	XMLName xml.Name `xml:"content"`
	Content string   `xml:",chardata"`
	Type    string   `xml:"type,attr"`
}
