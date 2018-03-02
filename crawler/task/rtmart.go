package task

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"honestman/app"
	"honestman/schema"

	"github.com/PuerkitoBio/goquery"
)

var (
	// should hit their db slowly
	// onece 100
	// rturlFormat = "http://www.rt-mart.com.tw/direct/index.php?action=product_search&prod_keyword=&p_data_num=100&page=%d"
	rturlFormat = "http://www.rt-mart.com.tw/direct/index.php?action=product_search&prod_keyword=&p_data_num=%d&page=%d"
	rtNum       = 100
)

// RTmart hold task RT-mart
type RTmart struct {
	Name     string
	Context  *app.Context
	interval int64
}

// NewRTmart new task
func NewRTmart(context *app.Context, interval int64) *RTmart {
	task := new(RTmart)
	task.Name = "RTmart"
	task.Context = context
	task.interval = interval
	return task
}

func (task *RTmart) String() string {
	return fmt.Sprintf("&RTmart")
}

// Run main loop
func (task *RTmart) Run() {

	for {
		// FIXME only one go routine here, more advance version to use worker
		// maybe block from upstream
		begin := time.Now().Truncate(time.Second)
		log.Println(task.Name, begin)
		task.Do()
		log.Println("Fetch all pages", task.Name, time.Now().Sub(begin))
		time.Sleep(time.Duration(task.interval) * time.Second)
	}
}

// Do do the dirty job
func (task *RTmart) Do() {

	var totalPage, total int
	var err error
	var url string
	var doc *goquery.Document

	// first page url
	url = fmt.Sprintf(rturlFormat, rtNum, 1)
	log.Println(url)
	doc, err = goquery.NewDocument(url)
	if err != nil {
		log.Println(err)
		return
	}
	task.process(doc)
	totalStr := doc.Find("span.t02").Text()
	log.Println("Found", totalStr)
	if totalStr != "" {
		total, err = strconv.Atoi(totalStr)
		if err != nil {
			log.Println(err)
			return
		}
		// alwayse plus one page
		totalPage = total/rtNum + 1

		for page := 2; page <= totalPage; page++ {
			// not DDOS the site
			time.Sleep(3 * time.Second)
			url = fmt.Sprintf(rturlFormat, rtNum, page)
			log.Println(url, "of", totalPage)
			doc, err = goquery.NewDocument(url)
			task.process(doc)
			if err != nil {
				log.Println(err)
				continue
			}

		}
	}
}

func (task *RTmart) process(doc *goquery.Document) {
	// <div class="indexProList">
	doc.Find("div.indexProList").Each(func(i int, s *goquery.Selection) {
		var origItem, newItem schema.Item

		if url, ok := s.Find("h5.for_proname > a").Attr("href"); ok {
			newItem.Url = url
		}

		if imgsrc, ok := s.Find("img").Attr("src"); ok {
			newItem.Imgsrc = imgsrc
		}

		newItem.Name = s.Find("h5.for_proname > a").Text()
		price := s.Find("div.for_pricebox > div").Text()

		if newItem.Url != "" {
			// should be an Item
			var err error
			priceStr := price[1:]
			if priceInt, err := strconv.Atoi(priceStr); err == nil {
				newItem.Price = priceInt
			}
			now := time.Now().Truncate(time.Second)
			newItem.Created = now
			newItem.Updated = now
			newItem.Source = task.Name

			// check exist
			err = task.Context.DB.Get(&origItem, "SELECT * from item WHERE url = $1 LIMIT 1", newItem.Url)

			err = task.Context.DB.Get(&origItem, "SELECT * from item WHERE url = $1 LIMIT 1", newItem.Url)

			switch {
			case err == sql.ErrNoRows:
				// no row in that url
				_, err = task.Context.DB.NamedExec(`INSERT INTO item 
				(price, diff, name, url, imgsrc, source, note, created, updated ) VALUES (
				:price,
				:diff,
				:name,
				:url,
				:imgsrc,
				:source,
				:note,
				:created,
				:updated)`, newItem)
				if err != nil {
					log.Println(err)
				}
			case err != nil:
				log.Println(err)
			default:
				// compare existed one
				newItem.Diff = newItem.Price - origItem.Price
				_, err = task.Context.DB.NamedExec(`UPDATE item SET
				price=:price,
				diff=:diff,
				name=:name,
				url=:url,
				imgsrc=:imgsrc,
				source=:source,
				note=:note,
				updated=:updated WHERE url=:url`, newItem)

				if err != nil {
					log.Println(err)
				}

			}
		}
	})
}
