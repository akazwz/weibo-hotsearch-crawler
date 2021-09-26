package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"log"
	"strconv"
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

	//var buf []byte
	var htmlContent string

	if err := chromedp.Run(ctx,
		//printToPDF(`https://s.weibo.com/top/summary?cate=realtimehot`, &buf),
		getHtmlContent(`https://s.weibo.com/top/summary?cate=realtimehot`, &htmlContent),
	); err != nil {
		log.Fatal(err)
	}
	/*if err := ioutil.WriteFile("sample.pdf", buf, 0644); err != nil {
		log.Fatal(err)
	}*/
	getHotSearchData(htmlContent)
}

func getHotSearchData(htmlContent string) {
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Println("new dom error")
	}
	dom.Find("#pl_top_realtimehot > table > tbody > tr").Each(func(i int, selection *goquery.Selection) {
		rank := selection.Find("td").Eq(0).Text()
		rankInt, err := strconv.Atoi(rank)
		// 非真正热搜内容直接返回
		if err != nil {
			return
		}
		// 热搜排名
		rank = fmt.Sprintf("%02d", rankInt)
		// 热搜内容
		content := selection.Find("td").Eq(1).Find("a").Text()
		// 热搜链接
		link := selection.Find("td").Eq(1).Find("a").AttrOr("href", "/weibo?="+content)
		// 热搜的排名和 tag分类
		hotAndTag := selection.Find("td").Eq(1).Find("span").Text()
		// 热搜的iconText, 比如 新 爆 等
		icon := selection.Find("td").Eq(2).Text()

		trimSpaceHotAndTag := strings.TrimSpace(hotAndTag)
		hotAndTagArr := strings.Split(trimSpaceHotAndTag, " ")
		// 热搜 热度
		hot := trimSpaceHotAndTag
		// 热搜 tag 比如 综艺 电影等
		tag := ""
		// 如果有 tag 的情况
		if len(hotAndTagArr) > 1 {
			hot = strings.TrimSpace(hotAndTagArr[1])
			tag = strings.TrimSpace(hotAndTagArr[0])
		}
		fmt.Println("rank:" + rank)
		fmt.Println("content:" + content)
		fmt.Println("link:" + link)
		fmt.Println("hot:" + hot)
		fmt.Println("tag：" + tag)
		fmt.Println("icon：" + icon)
		fmt.Println("------------------------------")
	})
}

func getHtmlContent(url string, html *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		// 等待热搜内容加载完毕
		chromedp.WaitReady("#pl_top_realtimehot", chromedp.ByQuery),
		// 获取热搜数据html
		chromedp.OuterHTML("#pl_top_realtimehot", html),
	}
}

func printToPDF(url string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		//chromedp.Emulate(device.IPhone8),
		chromedp.Navigate(url),
		chromedp.WaitReady("#pl_top_realtimehot", chromedp.ByQuery),
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
