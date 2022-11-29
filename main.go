package main

import (
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strconv"
)

type HatenaFeed struct {
    HatenaBookmarks []struct {
        Title string `xml:"title"`
        Link  string `xml:"link"`
        Desc  string `xml:"description"`
        Date  string `xml:"date"`
        Count int    `xml:"bookmarkcount"`
    } `xml:"item"`
}

type RSS2 struct {
    XMLName     xml.Name `xml:"rss"`
    Version     string   `xml:"version,attr"`
    Title       string   `xml:"channel>title"`
    Link        string   `xml:"channel>link"`
    Description string   `xml:"channel>description"`
    ItemList    []Item   `xml:"channel>item"`
}

type Item struct {
    Title string `xml:"title"`
    Link  string `xml:"link"`
    Desc  string `xml:"description"`
    Date  string `xml:"pubDate"`
}

func main() {
    port := os.Getenv("PORT")

    if port == "" {
        log.Fatal("$PORT must be set")
    }

    http.HandleFunc("/", handler)
    http.ListenAndServe(":"+port, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
    rss := getRSS("http://b.hatena.ne.jp/hotentry/all.rss")
    feed := HatenaFeed{}
    err := xml.Unmarshal([]byte(rss), &feed)
    if err != nil {
        fmt.Printf("error: %v", err)
        return
    }

    // URLパスを取得して数値に変換
    thresholdCount, _ := strconv.Atoi(r.URL.Path[1:])

    // 取得した閾値以上のはてブ数のスライスを作る
    items := make([]Item, 0)
    for _, item := range feed.HatenaBookmarks {
        if item.Count > thresholdCount {
            items = append(items, Item{item.Title, item.Link, item.Desc, item.Date})
        }
    }

    // RSSの内容を設定 / 取得した記事追加
    newFeed := RSS2{
        Version:     "2.0",
        Title:       "はてブフィルター",
        Link:        "https://hatebufilter.an.r.appspot.com/",
        Description: "はてブのブックマーク数でフィルタリングできるRSSフィード",
    }
    newFeed.ItemList = make([]Item, len(items))
    for i, item := range items {
        newFeed.ItemList[i].Title = item.Title
        newFeed.ItemList[i].Link = item.Link
        newFeed.ItemList[i].Desc = item.Desc
        newFeed.ItemList[i].Date = item.Date
    }

    // XMLに変換
    result, err := xml.MarshalIndent(newFeed, "  ", "    ")
    if err != nil {
        fmt.Printf("error: %v\n", err)
        return
    }

    // Webページに出力
    fmt.Fprint(w, "<?xml version='1.0' encoding='UTF-8'?>")
    fmt.Fprint(w, string(result))
}

func getRSS(url string) string {
    resp, err := http.Get("http://b.hatena.ne.jp/hotentry/all.rss")
    if err != nil {
        // エラーハンドリングを書く
    }
    defer resp.Body.Close()

    // _を使うことでエラーを無視できる
    body, _ := ioutil.ReadAll(resp.Body)

    return string(body)
}