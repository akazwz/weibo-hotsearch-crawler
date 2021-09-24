package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MontFerret/ferret/pkg/compiler"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/drivers/cdp"
	"github.com/MontFerret/ferret/pkg/drivers/http"
	"log"
)

func main() {
	fmt.Println("start to crawl")
	hotSearches := getHotSearch()

	for _, search := range hotSearches {
		fmt.Println(fmt.Sprintf("%d %s %s %d %s",
			search.Rank,
			search.Content,
			search.Link,
			search.Hot,
			search.Tag,
		))
	}
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

			WAIT_NAVIGATION(doc, 7000)

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

	// create a root context
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
