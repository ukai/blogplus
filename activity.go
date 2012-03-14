package blogplus

import (
	"bytes"
	"encoding/gob"
	"html/template"
)

// https://developers.google.com/+/api/latest/activities/list#response
type ActivityFeed struct {
	ETag          string     `json:"etag"`
	NextPageToken string     `json:"nextPageToken"`
	Items         []Activity `json:"items"`
}

// https://developers.google.com/+/api/latest/activities#resource
type Activity struct {
	ETag      string `json:"etag"`
	Title     string `json:"title"`
	Published string `json:"published"` // RFC3339 timestamp
	Updated   string `json:"updated"`   // RFC3339 timestamp
	Id        string `json:"id"`
	Url       string `json:"url"`
	Verb      string `json:"verb"`
	Object    Object `json:"object"`

	// used in blogplus
	FormedAttachment string
	Permalink        string
}

func (a Activity) HTMLFormedAttachment() template.HTML {
	return template.HTML(a.FormedAttachment)
}

type Object struct {
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments"`
	Replies     Counter      `json:"replies"`
	PlusOners   Counter      `json:"plusoners"`
	Resharers   Counter      `json:"resharers"`

	// used in blogplus
	Subject string
}

func (o Object) HTMLContent() template.HTML {
	return template.HTML(o.Content)
}

type Attachment struct {
	ObjectType  string `json:"objectType"`
	DisplayName string `json:"displayName"`
	Id          string `json:"id"`
	Content     string `json:"content"`
	Url         string `json:"url"`
	Image       Image  `json:"image"`
}

func (a Attachment) HTMLDisplayName() template.HTML {
	return template.HTML(a.DisplayName)
}

type Image struct {
	Url string `json:"url"`
}

type Counter struct {
	TotalItems int `json:"totalItems"`
}

func init() {
	var post Activity
	gob.Register(post)
}

func DecodeActivity(data []byte) (post Activity, err error) {
	buf := bytes.NewBuffer(data)
	d := gob.NewDecoder(buf)
	err = d.Decode(&post)
	return post, err
}

func EncodeActivity(post Activity) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	e := gob.NewEncoder(buf)
	err := e.Encode(post)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
