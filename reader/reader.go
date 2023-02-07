package reader

import (
	"encoding/xml"
	"fmt"
	"github.com/araddon/dateparse"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var source, sourceUrl string

type rss struct {
	Channel channel `xml:"channel"`
}
type channel struct {
	Title TitleFeed `xml:"title"`
	Link  LinkFeed  `xml:"link"`
	Items []RssItem `xml:"item"`
}

type TitleFeed string
type LinkFeed string

type RssItem struct {
	Title       string    `xml:"title"`
	Source      string    `xml:"source"`
	SourceURL   string    `xml:"source_url"`
	Link        string    `xml:"link"`
	PublishDate time.Time `xml:"pubdate"`
	Description string    `xml:"description"`
}

func (e *TitleFeed) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	var t xml.Token
	for t, err = d.Token(); err == nil; t, err = d.Token() {
		switch tt := t.(type) {
		case xml.CharData:
			source = string(tt)
		}
	}
	return nil
}

func (e *LinkFeed) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	var t xml.Token
	for t, err = d.Token(); err == nil; t, err = d.Token() {
		switch tt := t.(type) {
		case xml.CharData:
			sourceUrl = string(tt)
		}
	}
	return nil
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
	e.Source = source
	e.SourceURL = sourceUrl
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
	xmlBytes, err := getXML(url)
	if err != nil {
		return nil, fmt.Errorf("error during getXml: %s", err)
	}
	rssFeed := rss{}
	err = xml.Unmarshal(xmlBytes, &rssFeed)
	if err != nil {
		return nil, fmt.Errorf("error during Unmarshal: %s", err)
	}

	return rssFeed.Channel.Items, nil
}

func getXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}
