package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/ukai/blogplus"
	"net/http"
	"os"
)

var (
	userId string
	key    string
)

func init() {
	flag.StringVar(&userId, "user_id", "", "user id")
	flag.StringVar(&key, "key", "", "api key")
}

func main() {
	flag.Parse()
	fetcher := blogplus.NewFetcher(userId, key)
	fmt.Printf("%#v\n", fetcher)
	b := bufio.NewReader(os.Stdin)
	for {
		line, err := b.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println(line)
		if line == "" {
			fetcher.Fetch(&http.Client{})
		} else {
			fetcher.FetchPost(&http.Client{}, line)
		}
	}
}
