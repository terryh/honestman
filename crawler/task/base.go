package task

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	// Client with timeout
	Client        = &http.Client{Timeout: time.Duration(time.Second * 15)}
	fakeUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.186 Safari/537.36"
)

// CrawlerTask interface
// refresh in certain interval
type CrawlerTask interface {
	Run() // main loop
	Do()  // doing the dirty job
}

func fetchGet(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fakeUserAgent)
	return Client.Do(req)
}

func fetchGetBytes(url string) (resp []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fakeUserAgent)

	httpresp, err := Client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer httpresp.Body.Close()
	return ioutil.ReadAll(httpresp.Body)
}

func fetchPostBytes(url string, payloadStr string) (resp []byte, err error) {
	payload := []byte(payloadStr)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))

	if err != nil {
		return nil, err
	}

	// FIXME hard code Referer here, just DEMO
	req.Header.Set("User-Agent", fakeUserAgent)
	req.Header.Set("Referer", "https://online.carrefour.com.tw/search?key=+&categoryId=")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	httpresp, err := Client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer httpresp.Body.Close()
	return ioutil.ReadAll(httpresp.Body)
}
