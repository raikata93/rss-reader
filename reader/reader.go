package reader

import (
	"encoding/xml"
	"fmt"
	"github.com/araddon/dateparse"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rss struct {
	Channel channel `xml:"channel"`
}
type channel struct {
	Title string    `xml:"title"`
	Link  string    `xml:"link"`
	Items []RssItem `xml:"item"`
}

type RssItem struct {
	Title       string    `xml:"title"`
	Source      string    `xml:"source"`
	SourceURL   string    `xml:"source_url"`
	Link        string    `xml:"link"`
	PublishDate time.Time `xml:"pubdate"`
	Description string    `xml:"description"`
}

func (e *RssItem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	tokName := ""
	var t xml.Token
	for t, err = d.Token(); err == nil; t, err = d.Token() {
		switch tt := t.(type) {
		case xml.StartElement:
			tokName = tt.Name.Local
		case xml.CharData:
			switch tokName {
			case "title":
				e.Title = string(tt)
			case "link":
				e.Link = string(tt)
			case "pubDate":
				e.PublishDate, _ = dateparse.ParseAny(string(tt))
			case "description":
				e.Description = string(tt)
			}
		case xml.EndElement:
			tokName = ""
		}
	}
	return nil
}

func Parse(urls string) ([]RssItem, error) {
	var (
		wg        sync.WaitGroup
		allResult []RssItem
	)
	c := make(chan []RssItem)
	urlSlice := strings.Split(urls, ",")
	wg.Add(len(urlSlice))
	for _, url := range urlSlice {
		url := url
		go func() {
			defer wg.Done()
			result, err := parseXml(url)
			if err != nil {
				return
			}
			c <- result
		}()
	}
	go func() {
		defer close(c)
		wg.Wait()
	}()

	for result := range c {
		allResult = append(allResult, result...)
	}

	return allResult, nil
}

func parseXml(url string) ([]RssItem, error) {
	resp, err := getXML(url)
	defer resp.Body.Close()
	rss := rss{}
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(&rss)
	if err != nil {
		return nil, fmt.Errorf("Error Decode: %v\n", err)
	}

	var data []RssItem
	for _, item := range rss.Channel.Items {
		item.Source = rss.Channel.Title
		item.SourceURL = rss.Channel.Link
		data = append(data, item)
	}
	return data, nil
}

func getXML(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return &http.Response{}, fmt.Errorf("GET error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &http.Response{}, fmt.Errorf("status error: %v", resp.StatusCode)
	}
	return resp, nil
}
