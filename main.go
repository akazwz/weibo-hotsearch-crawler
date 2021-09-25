package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MontFerret/ferret/pkg/compiler"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/drivers/cdp"
	"github.com/MontFerret/ferret/pkg/drivers/http"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

func main() {
	fmt.Println("start to crawl")
	generatePDF(fmt.Sprintf("%s", time.Now().Format("2006-01-02-15-04-05")))
	/*location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatal("时区加载失败")
	}

	c := cron.New(cron.WithLocation(location))
	_, err = c.AddFunc("* * * * * ", func() {
		generatePDF(fmt.Sprintf("%s", time.Now().Format("2006-01-02-15-04-05")))
	})
	if err != nil {
		log.Fatal("定时任务添加失败", err)
	}
	c.Run()
	c.Start()*/
}

func generatePDF(pre string) {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)...)
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// capture pdf
	var outerBefore string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(`https://s.weibo.com/top/summary?cate=realtimehot`),
		chromedp.Sleep(10*time.Second),
		chromedp.OuterHTML("#pl_top_realtimehot", &outerBefore),
	); err != nil {
		log.Fatal(err)
	}
	fmt.Println(outerBefore)
}

func printToPDF(urlstr string, res *[]byte) chromedp.Tasks {

	return chromedp.Tasks{

		chromedp.ActionFunc(func(ctx context.Context) error {

			buf, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
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
