package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	start := time.Now()
	// 国内使用镜像站
	res, err := http.Get("https://hub.fastgit.org/akazwz")
	if err != nil {
		log.Println("get github error")
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("close error")
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Println("status code error")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println("new doc error")
	}

	contributions := doc.Find("#js-pjax-container > div.container-xl.px-3.px-md-4.px-lg-5 > div > div.flex-shrink-0.col-12.col-md-9.mb-4.mb-md-0 > div:nth-child(2) > div > div.mt-4.position-relative > div > div.col-12.col-lg-10 > div.js-yearly-contributions > div:nth-child(1)")
	contributeCountText := contributions.Find("h2").Text()
	trimCount := strings.TrimSpace(contributeCountText)
	countArr := strings.Split(trimCount, " ")
	countSum := countArr[0]
	fmt.Println(strings.TrimSpace(countSum))
	dataBoxDiv := contributions.Find("div")
	dataDiv := dataBoxDiv.Find("div")
	dataFrom := dataDiv.AttrOr("data-from", "data from")
	dataTo := dataDiv.AttrOr("data-to", "data to")
	dataSvg := dataDiv.Find("svg > g")

	dataArr := make([][]string, 0)
	dataArr = append(dataArr, []string{"count", "date", "level"})
	dataSvg.Find("g").Each(func(index int, selection *goquery.Selection) {
		selection.Find("rect").Each(func(i int, rect *goquery.Selection) {
			count := rect.AttrOr("data-count", "0")
			date := rect.AttrOr("data-date", "0")
			level := rect.AttrOr("data-level", "0")
			arr := []string{count, date, level}
			dataArr = append(dataArr, arr)
		})
	})
	fmt.Println(dataArr)
	fmt.Println(dataFrom)
	fmt.Println(dataTo)
	end := time.Now()
	fmt.Println(end.Sub(start))
}
