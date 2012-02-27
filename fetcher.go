package blogplus

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	BaseURL = "https://www.googleapis.com/plus/v1/"
)

type Fetcher struct {
	userId string
	key    string

	mu        sync.Mutex
	fetchETag string
}

func NewFetcher(userId, key string) *Fetcher {
	fetcher := &Fetcher{userId: userId, key: key}
	return fetcher
}

func (fetcher *Fetcher) GetActivities(client *http.Client, pageToken string) (*ActivityFeed, error) {
	url := BaseURL + "people/" + fetcher.userId + "/activities/public?num=100&key=" + fetcher.key
	if pageToken != "" {
		url += "&pageToken=" + pageToken
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if pageToken != "" && fetcher.fetchETag != "" {
		fetcher.mu.Lock()
		req.Header.Add("If-None-Match", fetcher.fetchETag)
		fetcher.mu.Unlock()
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if pageToken != "" {
		fetcher.mu.Lock()
		fetcher.fetchETag = resp.Header.Get("ETag")
		fetcher.mu.Unlock()
	}
	defer resp.Body.Close()
	var data ActivityFeed
	err = json.NewDecoder(resp.Body).Decode(&data)
	return &data, err
}

func (fetcher *Fetcher) getSinglePost(client *http.Client, activityId string) (post Activity, err error) {
	url := BaseURL + fmt.Sprintf("activities/%s?num=100&key=%s", activityId, fetcher.key)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return post, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return post, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&post)
	return post, err
}

func (fetcher *Fetcher) Fetch(client *http.Client) []Activity {
	activityFeed, err := fetcher.GetActivities(client, "")
	if err != nil {
		log.Println("getActivities:", err)
	}
	return activityFeed.Items
}

func (fetcher *Fetcher) FetchPost(client *http.Client, activityId string) []Activity {
	activity, err := fetcher.getSinglePost(client, activityId)
	if err != nil {
		log.Println("getSinglePost:", err)
		return nil
	}
	return []Activity{activity}
}
