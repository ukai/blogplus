package main

import (
	"flag"
	"fmt"
	"github.com/ukai/blogplus"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

var (
	userId     string
	key        string
	addr       string
	timeout    time.Duration
	driver     string
	datasource string
	initDb     bool

	title      string
	authorName string
	authorUri  string
	scheme     string
	host       string

	staticDir    string
	templateDir  string
	dumpTemplate bool
)

func init() {
	flag.StringVar(&userId, "user_id", "", "user id")
	flag.StringVar(&key, "key", "", "api key")
	flag.StringVar(&addr, "addr", ":80", "listen address")
	flag.DurationVar(&timeout, "timeout", 1*time.Hour, "timeout")
	flag.StringVar(&driver, "driver", "sqlite3", "database driver")
	flag.StringVar(&datasource, "datasource", "blogplus.db", "datasource")
	flag.BoolVar(&initDb, "init_db", false, "initialize db")
	flag.StringVar(&title, "title", "blogplus test", "title")
	flag.StringVar(&authorName, "author_name", "test user", "author's name")
	flag.StringVar(&authorUri, "author_uri", "http://example.com", "author's uri")
	flag.StringVar(&scheme, "scheme", "http", "url scheme")
	flag.StringVar(&host, "host", "", "url host")
	flag.StringVar(&staticDir, "static_dir", "", "static_dir")
	flag.StringVar(&templateDir, "template_dir", "", "template_dir")
	flag.BoolVar(&dumpTemplate, "dump_template", false, "dump template")
}

func main() {
	flag.Parse()
	c := NewController(timeout)
	var s blogplus.Storage
	var err error
	if driver == "memory" {
		s = blogplus.NewMemStorage()
	} else {
		if initDb {
			blogplus.InitDB(driver, datasource)
		}
		s, err = blogplus.NewDBStorage(driver, datasource)
		if err != nil {
			panic(err)
		}
	}
	s.SetFilter(blogplus.IsMeaningfulPost)
	fetcher := blogplus.NewFetcher(userId, key)

	b := blogplus.NewBlogplus(s, c)
	b.Title = title
	b.AuthorName = authorName
	b.AuthorUri = authorUri
	b.Scheme = scheme
	if host == "" {
		b.Host = "localhost" + addr
	} else {
		b.Host = host
	}
	b.SetStaticDir(staticDir)
	if dumpTemplate {
		if templateDir == "" {
			panic("need template_dir")
		}
		fmt.Printf("Extracting template in %s...", templateDir)
		b.ExtractTemplates(templateDir)
		fmt.Println("done")
		return
	}
	if templateDir != "" {
		b.LoadTemplates(templateDir)
	}

	go c.Run(fetcher, s)
	http.Handle("/", b)
	log.Println("start serving ", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
