package blogplus

import (
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	text_template "text/template"
)

var (
	BaseTempl       = template.New("base")
	HeaderTempl     = BaseTempl.New("header")
	headerTempl     = ""
	EntryTempl      = BaseTempl.New("entry")
	SidebarTempl    = BaseTempl.New("sidebar")
	ArchiveTempl    = SidebarTempl.New("archive")
	ArchivesJsTempl = text_template.New("archives.js")

	ImageAttachmentTempl = template.New("image_attachment")
	TextAttachmentTempl  = template.New("text_attachment")
)

const (
	baseTempl = `<!DOCTYPE html>
<html>
 <head>
  <title>{{.Blogplus.Title}}{{.Title}}</title>
  <link rel="me" type="text/html" href="{{.Blogplus.AuthorUri}}"/>
  <link rel="alternate" type="application/atom+xml" title="RSS" href="{{.Blogplus.Prefix}}` + atomFeedPath + `"/>
  {{template "header" .}}
  <script type="text/javascript" src="{{.Blogplus.Prefix}}` + archivesJsPath + `"></script>
  <script type="text/javascript" src="https://apis.google.com/js/plusone.js"></script>
 </head>
 <body>
  <div id="content">
   <h1><a itemprop="name" href="{{.Blogplus.Prefix}}` + mainPath + `">{{.Blogplus.Title}}</a></h1>
   <div id="main">
   {{if .Posts }}
    {{range .Posts}}{{template "entry" .}}{{end}} 
   {{else}}
    {{template "entry" .Post}}
   {{end}}
   </div>
   <div id="sidebar">
     {{template "sidebar" .}}
   </div>
  </div>
 </body>
</html>
`

	entryTempl      = `
 <div class="post">
  <div class="content">{{.Object.HTMLContent}}</div>
  {{if .FormedAttachment }}
  <div class="attachments">
   <hr class="attachment">
   {{.HTMLFormedAttachment}}
  </div>
  {{end}}
  <div class="meta">
   <span class="date">{{.Published}}</span>
   <span class="permalink"><a href="{{.Permalink}}">permalink</a></span>
   <span class="original_post"><a target="_blank" href="{{.Url}}">original post</a></span> |
    {{if .Object.PlusOners.TotalItems }}
      <span class="plusones"><a href="{{.Url}}">+{{.Object.PlusOners.TotalItems}}</a></span>
    {{end}}
    {{if .Object.Resharers.TotalItems }}
      <span class="reshares"><a href="{{.Url}}">{{.Object.Resharers.TotalItems}} reshares</a></span>
    {{end}}
    {{if .Object.Replies.TotalItems }}
      <span class="replies"><a href="{{.Url}}">{{.Object.Replies.TotalItems}} replies</a></span>
    {{end}}
    <g:plusone size="small" href="{{.Permalink}}"></g:plusone>
  </div>
  <hr class="entry" />
 </div>
`
	sidebarTempl    = `{{template "archive" .}}`
	archiveTempl    = `<h2>Archives</h2>
<div class="content">
 <select id="select_archive">
   <option value=""></option>
   {{range .ArchiveItems}}
   <option value="{{.Datespec}}">{{.Datespec}} ({{.Count}})</option>
   {{end}}
 </select>
</div>
`
	archivesJsTempl = `
(function() {
function redirectToArchive(e) {
  if (!e.target.value)
    return;
  window.location.href = '{{.ServerRootURL}}` + archivePath + `' + e.target.value;
}
window.addEventListener('load', function() {
  var selector = document.getElementById('select_archive');
  selector.addEventListener('change', redirectToArchive);
});
})()
`

	imageAttachmentTempl = `<a href="{{.Url}}"><img src="{{.Image.Url}}"></a>
{{if .DisplayName}}<a href="{{.Url}}">{{.DisplayName}}</a>{{end}}`
	textAttachmentTempl  = `{{if .VisualAttachments}}
<div>
{{range .VisualAttachments}}
<a href="{{.Url}}"><img src="{{.Image.Url}}"></a><br>
{{end}}
</div>
{{end}}
<div>
{{range .TextAttachments}}
<div>
<a href="{{.Url}}">{{.HTMLDisplayName}}</a><br>{{.Content}}
</div>
{{end}}
</div>`

	styleCss = `
/* style */
h1 {
    font-family: sans-serif;
}

h1 a {
    color: 'black';
    width: 100%;
}

#sidebar h2 {
    font-family: serif;
    font-size: medium;
    text-decoration:none;
    border-bottom:solid black 1px;
}

h1 a:hover {
    text-decoration:none;
}

div.post {
    width: 97%;
}

hr {
    border-color:green;
}

hr.entry {
    margin-bottom: 7ex;
}

div.attachments {
    margin-left:10%;
}

div.meta {
    font-family: serif;
    background-color:#EEE;
}

#main div.content {
    word-wrap: break-word;
}

#sidebar div.content {
    margin-left: 5px;
}

#content {
    margin-left: 4ex;
}

#main {
   max-width: 640px;
   border-right: 1px solid #BBB;
   float:left;
   line-height: 1.7em;
}

#sidebar {
    max-width: 320px;
    float:left;
}
`
)

func init() {
	_, err := BaseTempl.Parse(baseTempl)
	if err != nil {
		panic(err)
	}

	_, err = HeaderTempl.Parse(headerTempl)
	if err != nil {
		panic(err)
	}

	_, err = EntryTempl.Parse(entryTempl)
	if err != nil {
		panic(err)
	}

	_, err = SidebarTempl.Parse(sidebarTempl)
	if err != nil {
		panic(err)
	}

	_, err = ArchiveTempl.Parse(archiveTempl)
	if err != nil {
		panic(err)
	}

	_, err = ArchivesJsTempl.Parse(archivesJsTempl)
	if err != nil {
		panic(err)
	}

	_, err = ImageAttachmentTempl.Parse(imageAttachmentTempl)
	if err != nil {
		panic(err)
	}

	_, err = TextAttachmentTempl.Parse(textAttachmentTempl)
	if err != nil {
		panic(err)
	}
}

func createTempl(dir, path, templ string) {
	err := ioutil.WriteFile(filepath.Join(dir, path),
		[]byte(templ), os.FileMode(0644))
	if err != nil {
		panic(err)
	}
}

func loadTempl(dir, path string, templ *template.Template) {
	data, err := ioutil.ReadFile(filepath.Join(dir, path))
	if err != nil {
		panic(err)
	}
	_, err = templ.Parse(string(data))
	if err != nil {
		panic(err)
	}
}

func loadTextTempl(dir, path string, templ *text_template.Template) {
	data, err := ioutil.ReadFile(filepath.Join(dir, path))
	if err != nil {
		panic(err)
	}
	_, err = templ.Parse(string(data))
	if err != nil {
		panic(err)
	}
}

type TemplateContext struct {
	Posts        []Activity
	Post         Activity
	ArchiveItems []ArchiveItem
	ServerRoot   *url.URL // include prefix
	Title        string

	GlobalUpdated string
	*Blogplus
}

func (tc *TemplateContext) ServerRootURL() string {
	return tc.ServerRoot.String()
}

func GetAtomFeed(tc *TemplateContext) (data []byte, err error) {
	feed := AtomFeed{
		Id:         tc.ServerRoot.String(),
		Title:      tc.Blogplus.Title,
		Updated:    tc.GlobalUpdated,
		AuthorName: tc.Blogplus.AuthorName,
		AuthorUri:  tc.Blogplus.AuthorUri,
		Logo:       tc.Blogplus.LogoUrl}
	feed.Link = append(feed.Link, AtomLink{Href: tc.ServerRoot.String()})
	feed.Link = append(feed.Link, AtomLink{Href: tc.ServerRoot.String() + atomFeedPath, Rel: "self"})
	for _, post := range tc.Posts {
		e := AtomEntry{
			Id:    post.Permalink,
			Title: post.Object.Subject,
			Content: AtomContent{
				Content: post.Object.Content,
				Type:    "html"},
			Summary:   post.Object.Content,
			Published: post.Published,
			Updated:   post.Updated}
		feed.Entries = append(feed.Entries, e)
	}
	return xml.Marshal(feed)
}
