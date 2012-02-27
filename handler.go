package blogplus

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	text_template "text/template"
)

var (
	activityIdRe = regexp.MustCompile("[a-zA-Z0-9]+")
	dateSpecRe   = regexp.MustCompile("\\d+-\\d+")
)

const (
	mainPath       = "/"
	postPath       = "/post/"
	archivePath    = "/archive/"
	atomFeedPath   = "/feed"
	forceFetchPath = "/forcefetch"
	archivesJsPath = "/js/archives.js"
)

type Blogplus struct {
	storage    Storage
	c          Controller
	Title      string
	AuthorName string
	AuthorUri  string
	LogoUrl    string

	Scheme string
	Host   string
	Prefix string

	staticDir string
	fs        http.Handler
}

type Controller interface {
	ForceFetch(req *http.Request)
	MaybeFetch(req *http.Request)
	MaybeFetchPost(req *http.Request, activityId string)
}

func NewBlogplus(s Storage, c Controller) *Blogplus {
	return &Blogplus{storage: s, c: c, Prefix: ""}
}

func (b *Blogplus) SetStaticDir(dir string) {
	if dir != "" {
		b.staticDir = "/" + dir
		b.fs = http.FileServer(http.Dir("."))
		headerTempl = ""
		if f, err := os.Open(filepath.Join(dir, "favicon.png")); err == nil {
			headerTempl += fmt.Sprintf(`<link rel="shortcut icon" href="{{.Blogplus.Prefix}}/%s/favicon.png"/>`, dir)
			f.Close()
		} else {
			log.Println("missing favicon.png")
		}
		if f, err := os.Open(filepath.Join(dir, "style.css")); err != nil {
			err = ioutil.WriteFile(filepath.Join(dir, "style.css"),
				[]byte(styleCss), os.FileMode(0644))
			if err != nil {
				log.Println("create style.css error:", err)
			}
		} else {
			f.Close()
		}
		headerTempl += fmt.Sprintf(`<link rel="stylesheet" href="{{.Blogplus.Prefix}}/%s/style.css"/>`, dir)
		_, err := HeaderTempl.Parse(headerTempl)
		if err != nil {
			panic(err)
		}
	}
}

func (b *Blogplus) ExtractTemplates(dir string) {
	_ = os.MkdirAll(dir, os.FileMode(0755))

	createTempl(dir, "base.tmpl", baseTempl)
	createTempl(dir, "header.tmpl", headerTempl)
	createTempl(dir, "entry.tmpl", entryTempl)
	createTempl(dir, "sidebar.tmpl", sidebarTempl)
	createTempl(dir, "archive.tmpl", archiveTempl)
	createTempl(dir, "atom_feed.tmpl", atomFeedTempl)
	createTempl(dir, "archives.js.tmpl", archivesJsTempl)
	createTempl(dir, "image_attachment.tmpl", imageAttachmentTempl)
	createTempl(dir, "text_attachment.tmpl", textAttachmentTempl)
}

func (b *Blogplus) LoadTemplates(dir string) {
	BaseTempl = template.New("base")
	loadTempl(dir, "base.tmpl", BaseTempl)
	HeaderTempl = BaseTempl.New("header")
	loadTempl(dir, "header.tmpl", HeaderTempl)
	EntryTempl = BaseTempl.New("entry")
	loadTempl(dir, "entry.tmpl", EntryTempl)
	SidebarTempl = BaseTempl.New("sidebar")
	loadTempl(dir, "sidebar.tmpl", SidebarTempl)
	ArchiveTempl = SidebarTempl.New("archive")
	loadTempl(dir, "archive.tmpl", ArchiveTempl)
	AtomFeedTempl = template.New("atom_feed")
	loadTempl(dir, "atom_feed.tmpl", AtomFeedTempl)
	ArchivesJsTempl = text_template.New("archives.js")
	loadTextTempl(dir, "archives.js.tmpl", ArchivesJsTempl)
	ImageAttachmentTempl = template.New("image_attachment")
	loadTempl(dir, "image_attachment.tmpl", ImageAttachmentTempl)
	TextAttachmentTempl = template.New("text_attachment")
	loadTempl(dir, "text_attachment.tmpl", TextAttachmentTempl)
}

func (b *Blogplus) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case b.Prefix + mainPath:
		b.ServeMain(w, req)
	case b.Prefix + atomFeedPath:
		b.ServeFeed(w, req)
	case b.Prefix + forceFetchPath:
		b.ServeForceFetch(w, req)
	case b.Prefix + archivesJsPath:
		b.ServeArchivesJs(w, req)
	default:
		if strings.HasPrefix(req.URL.Path, b.Prefix+postPath) {
			b.ServePost(w, req)
		} else if strings.HasPrefix(req.URL.Path, b.Prefix+archivePath) {
			b.ServeArchive(w, req)
		} else if strings.HasPrefix(req.URL.Path, b.Prefix+b.staticDir) {
			b.fs.ServeHTTP(w, req)
		} else {
			http.NotFound(w, req)
		}
	}
}

func getServerRoot(b *Blogplus, req *http.Request) *url.URL {
	scheme := b.Scheme
	if scheme == "" {
		scheme = req.URL.Scheme
	}
	host := b.Host
	if host == "" {
		host = req.URL.Host
	}
	return &url.URL{Scheme: scheme, Host: host, Path: b.Prefix}
}

func getPostUrl(b *Blogplus, req *http.Request) *url.URL {
	scheme := b.Scheme
	if scheme == "" {
		scheme = req.URL.Scheme
	}
	host := b.Host
	if host == "" {
		host = req.URL.Host
	}
	return &url.URL{Scheme: scheme, Host: host, Path: b.Prefix + postPath}
}

func (b *Blogplus) ServeMain(w http.ResponseWriter, req *http.Request) {
	var posts []Activity
	postUrl := getPostUrl(b, req)
	for _, post := range b.storage.GetLatestPosts(req) {
		processPost(&post, postUrl)
		posts = append(posts, post)
	}
	b.c.MaybeFetch(req)
	err := BaseTempl.Execute(w,
		&TemplateContext{
			Posts: posts, ArchiveItems: b.storage.GetDates(req),
			ServerRoot: getServerRoot(b, req),
			Title:      b.Title,
			Blogplus:   b})
	if err != nil {
		log.Println("template error:", err)
	}
}

func (b *Blogplus) ServePost(w http.ResponseWriter, req *http.Request) {
	activityId := path.Base(req.URL.Path)
	if !activityIdRe.MatchString(activityId) {
		log.Println("unexpected activityId:", activityId)
		http.NotFound(w, req)
		return
	}
	postUrl := getPostUrl(b, req)
	post, found := b.storage.GetPost(req, activityId)
	if !found {
		http.NotFound(w, req)
		return
	}
	processPost(&post, postUrl)
	b.c.MaybeFetchPost(req, activityId)
	err := BaseTempl.Execute(w,
		&TemplateContext{
			Post: post, ArchiveItems: b.storage.GetDates(req),
			ServerRoot: getServerRoot(b, req),
			Title:      b.Title + " " + post.Title,
			Blogplus:   b})
	if err != nil {
		log.Println("template error:", err)
	}
}

func (b *Blogplus) ServeArchive(w http.ResponseWriter, req *http.Request) {
	datespec := path.Base(req.URL.Path)
	if !dateSpecRe.MatchString(datespec) {
		log.Println("unexpected datespec:", datespec)
		http.NotFound(w, req)
		return
	}
	postUrl := getPostUrl(b, req)
	var posts []Activity
	for _, post := range b.storage.GetArchivedPosts(req, datespec) {
		processPost(&post, postUrl)
		posts = append(posts, post)
	}
	if len(posts) == 0 {
		log.Println("no posts")
		http.NotFound(w, req)
		return
	}
	err := BaseTempl.Execute(w,
		&TemplateContext{
			Posts: posts, ArchiveItems: b.storage.GetDates(req),
			ServerRoot: getServerRoot(b, req),
			Title:      b.Title,
			Blogplus:   b})
	if err != nil {
		log.Println("template error:", err)
	}
}

func (b *Blogplus) ServeFeed(w http.ResponseWriter, req *http.Request) {
	postUrl := getPostUrl(b, req)
	globalUpdated := ""
	var posts []Activity
	for _, post := range b.storage.GetLatestPosts(req) {
		processPost(&post, postUrl)
		posts = append(posts, post)
		if globalUpdated < post.Updated {
			globalUpdated = post.Updated
		}
	}
	// TODO(ukai): weekly-2012-02-22: template will mangle <? to &lt;?...
	_, err := io.WriteString(w, `<?xml version="1.0" encoding="utf-8"?>`)
	if err != nil {
		log.Println("servefeed:", err)
	}
	err = AtomFeedTempl.Execute(w,
		&TemplateContext{
			Posts:         posts,
			ServerRoot:    getServerRoot(b, req),
			Title:         b.Title,
			GlobalUpdated: globalUpdated,
			Blogplus:      b})
	if err != nil {
		log.Println("template error:", err)
	}
}

func (b *Blogplus) ServeForceFetch(w http.ResponseWriter, req *http.Request) {
	b.c.ForceFetch(req)
	http.Redirect(w, req, b.Prefix+mainPath, http.StatusFound)
}

func (b *Blogplus) ServeArchivesJs(w http.ResponseWriter, req *http.Request) {
	err := ArchivesJsTempl.Execute(w,
		&TemplateContext{
			ServerRoot: getServerRoot(b, req),
			Blogplus:   b})
	if err != nil {
		log.Println("template error:", err)
	}
}
