package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/MontFerret/ferret/pkg/compiler"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/drivers/cdp"
	"github.com/MontFerret/ferret/pkg/drivers/http"
	"github.com/akazwz/weibo-hotsearch-crawler/global"
	"github.com/akazwz/weibo-hotsearch-crawler/initialize"
	"github.com/akazwz/weibo-hotsearch-crawler/utils/influx"
	"github.com/robfig/cron/v3"
)

func main() {
	fmt.Println("start to crawl")

	global.VP = initialize.InitViper()
	if global.VP == nil {
		fmt.Println("配置文件初始化失败")
	}

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatal("时区加载失败")
	}

	c := cron.New(cron.WithLocation(location))
	_, err = c.AddFunc("* * * * * ", func() {
		t := time.Now()
		hotSearches := getHotSearch()
		for _, search := range hotSearches {
			tags := map[string]string{}
			fields := map[string]interface{}{}
			tags["rank"] = fmt.Sprintf("%02d", search.Rank)
			fields["content"] = search.Content
			fields["link"] = search.Link
			fields["hot"] = search.Hot
			fields["tag"] = search.Tag

			err = influx.Write("new_hot_search", tags, fields, t)
			if err != nil {
				log.Fatal("influx error:", err)
			}
		}
	})
	if err != nil {
		log.Fatal("定时任务添加失败", err)
	}
	c.Run()
	c.Start()

}

type HotSearch struct {
	Rank    int64  `json:"rank"`
	Content string `json:"content"`
	Link    string `json:"link"`
	Hot     int64  `json:"hot"`
	Tag     string `json:"tag"`
}

func getHotSearch() []*HotSearch {
	query := `
			LET doc = DOCUMENT("https://s.weibo.com/top/summary?cate=realtimehot", {driver: "cdp"})

			WAIT_NAVIGATION(doc, 10000)

			LET data = ELEMENTS(doc, "div#pl_top_realtimehot > table > tbody > tr")
			LET realData = (
				FOR el IN data
					FILTER TO_INT(el[0].innerText) > 0
					LET rank = TO_INT(el[0].innerText)
					LET content = el[1][0].innerText
					LET link = el[1][0].attributes.href
					LET hot = TO_INT(el[1][1].innerText)
					LET tag = el[2].innerText

					RETURN {
						rank: rank, 
						content: content,
						link: link,
						hot: hot,
						tag: tag,
					}
			)
			
			RETURN realData
		`

	comp := compiler.New()

	program, err := comp.Compile(query)

	if err != nil {
		log.Fatal("compile error")
	}

	ctx := context.Background()
	ctx = drivers.WithContext(ctx, cdp.NewDriver())
	ctx = drivers.WithContext(ctx, http.NewDriver(), drivers.AsDefault())

	out, err := program.Run(ctx)

	if err != nil {
		log.Fatal("run error")
	}

	res := make([]*HotSearch, 0)

	err = json.Unmarshal(out, &res)

	if err != nil {
		log.Fatal("json error")
	}

	return res
}
