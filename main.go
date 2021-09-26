package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"strings"
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
		chromedp.Flag("headless", true),
	)...)
	ctx, cancel = chromedp.NewContext(ctx)
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var buf []byte
	var htmlContent string

	if err := chromedp.Run(ctx,
		printToPDF(`https://s.weibo.com/top/summary?cate=realtimehot`, &buf),
		getHtmlContent(`https://s.weibo.com/top/summary?cate=realtimehot`, &htmlContent),
	); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("sample.pdf", buf, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Println(htmlContent)
}

func getHotSearchData(htmlContent string) {
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Println("new dom error")
	}
	dom.Find("").Each(func(i int, selection *goquery.Selection) {

	})

}

func getHtmlContent(url string, html *string) chromedp.Tasks {
	return chromedp.Tasks{
		//chromedp.Emulate(device.IPhone8),
		chromedp.Navigate(url),
		chromedp.WaitReady("div#pl_top_realtimehot", chromedp.ByQuery),
		chromedp.OuterHTML("div#pl_top_realtimehot", html),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return nil
		}),
	}
}

func printToPDF(url string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		//chromedp.Emulate(device.IPhone8),
		chromedp.Navigate(url),
		chromedp.WaitReady("div#pl_top_realtimehot", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPaperWidth(7).WithPaperHeight(21).WithPrintBackground(false).WithPreferCSSPageSize(true).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}
