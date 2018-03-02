package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"honestman/app"
	"honestman/schema"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	// onece 35 their rule
	carrefourURL         = "https://online.carrefour.com.tw/CarrefourECProduct/GetSearchJson"
	carrefourQueryFormat = "pageIndex=%d&pageSize=%d&OrderById=0"
	cfNum                = 35
	// reRTtotal = regexp.MustCompile(`<span class="t02">([0-9]{2,})</span>件商品`)
	// reConten = regexp.MustCompile(`<span class="t02">([0-9]{2,})</span>件商品`)
)

// CfItem carrefour item
type CfItem struct {
	DisplayId                      int    `json:"DisplayId"`
	ID                             int    `json:"Id"`
	IsWish                         bool   `json:"IsWish"`
	ItemQtyPerPack                 int    `json:"ItemQtyPerPack"`
	ItemQtyPerPackFormat           string `json:"ItemQtyPerPackFormat"`
	Name                           string `json:"Name"`
	PictureUrl                     string `json:"PictureUrl"`
	Price                          string `json:"Price"`
	PromotionProductPicUrl         string `json:"PromotionProductPicUrl"`
	QucikShippingProductListPicUrl string `json:"QucikShippingProductListPicUrl"`
	SeName                         string `json:"SeName"`
	SpecialPrice                   string `json:"SpecialPrice"`
	SpecialStoreProductListPicUrl  string `json:"SpecialStoreProductListPicUrl"`
	Specification                  string `json:"Specification"`
}

// CfJson carrefour json response
type CfJson struct {
	Content struct {
		CategoryId           int      `json:"CategoryId"`
		Count                int      `json:"Count"`
		Key                  string   `json:"Key"`
		OrderById            int      `json:"OrderById"`
		PageSize             int      `json:"PageSize"`
		RewardId             int      `json:"RewardId"`
		StoreActivityBasicId int      `json:"StoreActivityBasicId"`
		SearchCategoryId     string   `json:"searchCategoryId"`
		ProductListModel     []CfItem `json:"ProductListModel"`
	} `json:"content"`
	Success int `json:"success"`
}

// Carrefour hold task RT-mart
type Carrefour struct {
	Name     string
	Context  *app.Context
	interval int64
}

// NewCarrefour new task
func NewCarrefour(context *app.Context, interval int64) *Carrefour {
	task := new(Carrefour)
	task.Name = "Carrefour"
	task.Context = context
	task.interval = interval
	return task
}

func (task *Carrefour) String() string {
	return fmt.Sprintf("&Carrefour")
}

// Run main loop
func (task *Carrefour) Run() {

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
func (task *Carrefour) Do() {

	var totalPage int
	var err error
	var payload string
	var jsonBytes []byte
	var jsresp CfJson

	// first page url
	payload = fmt.Sprintf(carrefourQueryFormat, 1, cfNum)
	log.Println(payload)

	jsonBytes, err = fetchPostBytes(carrefourURL, payload)
	if err != nil {
		log.Println(err)
		return
	}

	err = json.Unmarshal(jsonBytes, &jsresp)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(jsresp.Success, jsresp.Content.Count, len(jsresp.Content.ProductListModel))
	if jsresp.Success == 1 {
		totalPage = jsresp.Content.Count/cfNum + 1
		task.process(jsresp.Content.ProductListModel)
	}

	for page := 2; page <= totalPage; page++ {
		// not DDOS the site
		time.Sleep(3 * time.Second)
		payload = fmt.Sprintf(carrefourQueryFormat, page, cfNum)
		log.Println(payload, "of", totalPage)
		jsonBytes, err = fetchPostBytes(carrefourURL, payload)
		if err != nil {
			log.Println(err)
			continue
		}

		var jsp CfJson
		err = json.Unmarshal(jsonBytes, &jsp)
		if err != nil {
			log.Println(err)
			continue
		}
		if jsresp.Success == 1 {
			task.process(jsp.Content.ProductListModel)
		}
	}
}

func (task *Carrefour) process(items []CfItem) {
	// <div class="indexProList">
	for _, item := range items {
		var origItem, newItem schema.Item
		var err error
		if item.SeName == "" {
			continue
		}

		newItem.Url = fmt.Sprintf("https://online.carrefour.com.tw%s", strings.Split(item.SeName, "?")[0])
		newItem.Imgsrc = item.PictureUrl
		newItem.Name = item.Name
		newItem.Note = item.Specification
		if priceInt, err := strconv.Atoi(item.Price); err == nil {
			newItem.Price = priceInt
		}

		now := time.Now().Truncate(time.Second)
		newItem.Created = now
		newItem.Updated = now
		newItem.Source = task.Name
		// log.Printf("%#v\n", newItem)
		// log.Println(origItem)
		// check exist
		err = task.Context.DB.Get(&origItem, "SELECT * from item WHERE url = $1 LIMIT 1", newItem.Url)

		switch {
		case err == sql.ErrNoRows:
			// no row in that url
			_, err = task.Context.DB.NamedExec(`INSERT INTO item (price, diff, name, url, imgsrc, source, note, created, updated ) VALUES (
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
			// no row in that url
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
}
