package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/akazwz/weibo-hotsearch-crawler/global"
	"github.com/akazwz/weibo-hotsearch-crawler/initialize"
	"github.com/akazwz/weibo-hotsearch-crawler/utils/influx"
	"github.com/akazwz/weibo-hotsearch-crawler/utils/notify"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"log"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("start to crawl")

	global.VP = initialize.InitViper()
	if global.VP == nil {
		fmt.Println("配置文件初始化失败")
	}

	notify.SendVerifyMessage("15153953308", "333333")

	/*//generatePDF(fmt.Sprintf("%s", time.Now().Format("2006-01-02-15-04-05")))
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatal("时区加载失败")
	}

	// 开启定时任务
	c := cron.New(cron.WithLocation(location))
	_, err = c.AddFunc("* * * * * ", func() {
		crawlHotSearch(time.Now())
	})

	if err != nil {
		log.Fatal("定时任务添加失败", err)
	}
	c.Run()
	c.Start()*/
}

func crawlHotSearch(t time.Time) {
	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://127.0.0.1:9222/")
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	//var buf []byte
	var htmlContent string

	if err := chromedp.Run(ctx,
		//printToPDF(`https://s.weibo.com/top/summary?cate=realtimehot`, &buf),
		getHtmlContent(`https://s.weibo.com/top/summary?cate=realtimehot`, &htmlContent),
	); err != nil {
		log.Println(err)
	}
	/*if err := ioutil.WriteFile("sample.pdf", buf, 0644); err != nil {
		log.Fatal(err)
	}*/
	getHotSearchData(htmlContent, t)
}

func getHotSearchData(htmlContent string, t time.Time) {
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

		tags := map[string]string{}
		fields := map[string]interface{}{}
		tags["rank"] = rank
		fields["content"] = content
		fields["link"] = link
		fields["hot"] = hot
		fields["tag"] = tag
		fields["icon"] = icon

		err = influx.Write("new-hot", tags, fields, t)
		if err != nil {
			log.Println("influx error:", err)
		}
	})
}

func getHtmlContent(url string, html *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		// 等待热搜内容加载完毕
		chromedp.WaitVisible("#pl_top_realtimehot", chromedp.ByQuery),
		// 获取热搜数据html
		chromedp.OuterHTML("#pl_top_realtimehot", html, chromedp.ByQuery),
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
