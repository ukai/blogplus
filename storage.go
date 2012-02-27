package blogplus

import (
	"container/heap"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type Storage interface {
	SetFilter(func(Activity) bool)
	StorePosts(req *http.Request, posts []Activity)
	GetLatestPosts(req *http.Request) []Activity
	GetPost(req *http.Request, activityId string) (Activity, bool)
	GetDates(req *http.Request) []ArchiveItem
	GetArchivedPosts(req *http.Request, datespec string) []Activity
}

type ArchiveItem struct {
	Datespec string
	Count    int
}

type ArchiveItemList []ArchiveItem

func (a ArchiveItemList) Len() int { return len(a) }
func (a ArchiveItemList) Less(i, j int) bool {
	return a[i].Datespec > a[j].Datespec
}
func (a ArchiveItemList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type latestActivityList struct {
	a []Activity
}

func (l *latestActivityList) Len() int { return len(l.a) }
func (l *latestActivityList) Less(i, j int) bool {
	return l.a[i].Published > l.a[j].Published
}
func (l *latestActivityList) Swap(i, j int) { l.a[i], l.a[j] = l.a[j], l.a[i] }
func (l *latestActivityList) Push(x interface{}) {
	l.a = append(l.a, x.(Activity))
}
func (l *latestActivityList) Pop() interface{} {
	r := l.a[len(l.a)-1]
	l.a = l.a[:len(l.a)-1]
	return r
}

type MemStorage struct {
	m      map[string]Activity   // activityid -> post
	a      map[string][]Activity // datespec -> list of post
	h      latestActivityList
	filter func(Activity) bool
	mu     sync.Mutex
}

func NewMemStorage() *MemStorage {
	s := &MemStorage{
		m: make(map[string]Activity),
		a: make(map[string][]Activity)}
	heap.Init(&s.h)
	return s
}

func (s *MemStorage) SetFilter(filter func(Activity) bool) {
	s.filter = filter
}

func GetDatespec(published string) string {
	s := strings.Split(published, "-")
	if len(s) >= 2 {
		return fmt.Sprintf("%s-%s", s[0], s[1])
	}
	return ""
}

func (s *MemStorage) StorePosts(req *http.Request, posts []Activity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, post := range posts {
		if s.filter != nil && !s.filter(post) {
			continue
		}
		log.Printf("store: %s\n", post.Id)
		s.m[post.Id] = post
		updated := false
		datespec := GetDatespec(post.Published)
		for i, p := range s.a[datespec] {
			if p.Id == post.Id {
				s.a[datespec][i] = post
				updated = true
				break
			}
		}
		if !updated {
			s.a[datespec] = append(s.a[datespec], post)
		}
		heap.Push(&s.h, post)
		if s.h.Len() >= 10 {
			_ = heap.Remove(&s.h, s.h.Len()-1)
		}
	}
}

func (s *MemStorage) GetLatestPosts(req *http.Request) []Activity {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.h.a
}

func (s *MemStorage) GetPost(req *http.Request, activityId string) (Activity, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.m[activityId]
	return a, ok
}

func (s *MemStorage) GetDates(req *http.Request) []ArchiveItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	var a ArchiveItemList
	for datespec, l := range s.a {
		a = append(a, ArchiveItem{Datespec: datespec, Count: len(l)})
	}
	sort.Sort(a)
	return a
}

func (s *MemStorage) GetArchivedPosts(req *http.Request, datespec string) []Activity {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l, found := s.a[datespec]; found {
		return l
	}
	return nil
}
