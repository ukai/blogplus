package blogplus

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sort"
)

const (
	createTable = `create table blogplus (id text not null primary key, published text, datespec text, post blob);
create index published_idx on blogplus (published desc);
create index datespec_idx on blogplus (datespec desc);
`
)

func InitDB(driver, datasource string) (*sql.DB, error) {
	log.Println("Initialize db:", driver, ":", datasource)
	os.Remove(datasource)
	db, err := sql.Open(driver, datasource)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(createTable)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

type DBStorage struct {
	db     *sql.DB
	filter func(Activity) bool
}

func NewDBStorage(driver, datasource string) (*DBStorage, error) {
	db, err := sql.Open(driver, datasource)
	if err != nil {
		return nil, err
	}

	return &DBStorage{db: db}, nil
}

func (s *DBStorage) SetFilter(filter func(Activity) bool) {
	s.filter = filter
}

func (s *DBStorage) StorePosts(req *http.Request, posts []Activity) {
	stmt, err := s.db.Prepare(`insert or replace into blogplus(id, published, datespec, post) values(?, ?, ?, ?)`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	for _, post := range posts {
		if s.filter != nil && !s.filter(post) {
			continue
		}
		datespec := GetDatespec(post.Published)
		data, err := EncodeActivity(post)
		if err != nil {
			log.Println("encode error:", err)
			continue
		}
		_, err = stmt.Exec(post.Id, post.Published, datespec, data)
		if err != nil {
			log.Println("StorePosts:", err)
		}
	}
}

func scanPost(rows *sql.Rows) (post Activity, err error) {
	var data []byte
	err = rows.Scan(&data)
	if err != nil {
		return post, err
	}
	return DecodeActivity(data)
}

func (s *DBStorage) GetLatestPosts(req *http.Request) []Activity {
	rows, err := s.db.Query(`select post from blogplus order by published desc limit 10`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var posts []Activity
	for rows.Next() {
		post, err := scanPost(rows)
		if err == nil {
			posts = append(posts, post)
		}
	}
	return posts
}

func (s *DBStorage) GetPost(req *http.Request, activityId string) (Activity, bool) {
	stmt, err := s.db.Prepare(`select post from blogplus where id = ?`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	var post Activity
	rows, err := stmt.Query(activityId)
	if err != nil {
		return post, false
	}
	defer rows.Close()
	if !rows.Next() {
		return post, false
	}
	post, err = scanPost(rows)
	if err != nil {
		log.Println("get post error:", err)
	}
	return post, err == nil
}

func (s *DBStorage) GetDates(req *http.Request) []ArchiveItem {
	rows, err := s.db.Query(`select datespec, count(*) from blogplus group by datespec`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var archiveItems ArchiveItemList
	for rows.Next() {
		var ai ArchiveItem
		err = rows.Scan(&ai.Datespec, &ai.Count)
		if err != nil {
			continue
		}
		archiveItems = append(archiveItems, ai)
	}
	sort.Sort(archiveItems)
	return archiveItems
}

func (s *DBStorage) GetArchivedPosts(req *http.Request, datespec string) []Activity {
	stmt, err := s.db.Prepare(`select post from blogplus where datespec = ? order by published desc`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(datespec)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var posts []Activity
	for rows.Next() {
		post, err := scanPost(rows)
		if err == nil {
			posts = append(posts, post)
		}
	}
	return posts
}
