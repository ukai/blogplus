package main

import (
	"github.com/ukai/blogplus"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	FetcherCount = 1000
)

type Controller struct {
	mu          sync.Mutex
	fetchCount  int
	fetchCounts map[string]int

	event   chan *string
	done    chan bool
	timeout time.Duration
}

func NewController(timeout time.Duration) *Controller {
	fs := &Controller{
		fetchCounts: make(map[string]int),
		event:       make(chan *string),
		done:        make(chan bool),
		timeout:     timeout}
	return fs
}

func fetchAllPosts(fetcher *blogplus.Fetcher, storage blogplus.Storage) {
	client := &http.Client{}
	var req *http.Request
	latest_ids := make(map[string]bool)
	for _, post := range storage.GetLatestPosts(req) {
		latest_ids[post.Id] = true
	}
	log.Println("fetch the latest...")
	activityFeed, err := fetcher.GetActivities(client, "")
	if err != nil {
		log.Println("fetcher error:", err)
		return
	}
	var allItems []blogplus.Activity
Loop:
	for len(activityFeed.Items) > 0 {
		for _, post := range activityFeed.Items {
			if latest_ids[post.Id] {
				break Loop
			}
		}
		allItems = append(allItems, activityFeed.Items...)
		if activityFeed.NextPageToken == "" {
			break Loop
		}
		activityFeed, err = fetcher.GetActivities(client, activityFeed.NextPageToken)
		if err != nil {
			log.Println("fetcher error:", err)
			break Loop
		}
	}
	storage.StorePosts(req, allItems)
	log.Println("fetch the latest done")
}

func (c *Controller) Run(fetcher *blogplus.Fetcher, storage blogplus.Storage) {
	fetchAllPosts(fetcher, storage)
	for {
		var activityId *string
		select {
		case activityId = <-c.event:
			if activityId == nil {
				c.done <- true
				return
			}
		case <-time.After(c.timeout):
		}
		var posts []blogplus.Activity
		if activityId != nil && *activityId != "" {
			posts = fetcher.FetchPost(&http.Client{}, *activityId)
		} else {
			posts = fetcher.Fetch(&http.Client{})
		}
		var req *http.Request
		storage.StorePosts(req, posts)
	}
}

func (c *Controller) fetch() {
	c.event <- new(string)
}

func (c *Controller) fetchPost(activityId string) {
	c.event <- &activityId
}

func (c *Controller) ForceFetch(req *http.Request) {
	c.fetch()
}

func (c *Controller) MaybeFetch(req *http.Request) {
	doFetch := false
	c.mu.Lock()
	c.fetchCount += 1
	if c.fetchCount > FetcherCount {
		c.fetchCount = 0
		doFetch = true
	}
	c.mu.Unlock()
	if doFetch {
		c.fetch()
	}
}

func (c *Controller) MaybeFetchPost(req *http.Request, activityId string) {
	doFetch := false
	c.mu.Lock()
	if _, ok := c.fetchCounts[activityId]; !ok {
		doFetch = true
		c.fetchCounts[activityId] = 0
	}
	c.fetchCounts[activityId] += 1
	if c.fetchCounts[activityId] > FetcherCount {
		c.fetchCounts[activityId] = 0
		doFetch = true
	}
	c.mu.Unlock()
	if doFetch {
		c.fetchPost(activityId)
	}
}

func (c *Controller) Finish(req *http.Request) {
	c.event <- nil
	<-c.done
}
