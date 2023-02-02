package reader

import (
	"encoding/json"
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
	Title       string    `xml:"title" json:"title"`
	Source      string    `xml:"source" json:"source"`
	SourceURL   string    `xml:"source_url" json:"source_url"`
	Link        string    `xml:"link" json:"link"`
	PublishDate time.Time `xml:"pubdate" json:"publish_date"`
	Description string    `xml:"description" json:"description"`
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

func Parse(urls string) []byte {
	var wg sync.WaitGroup
	var all_result []RssItem
	results := make(map[string][]RssItem)
	c := make(chan []RssItem)
	urlSlice := strings.Split(urls, ",")
	for _, url := range urlSlice {
		wg.Add(1)
		url := url
		go func() {
			result := parseXml(url)
			c <- result
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()
	for result := range c {
		all_result = append(all_result, result...)
	}
	results["items"] = all_result
	jsonStr, _ := json.Marshal(results["items"])

	return jsonStr
}

func parseXml(url string) []RssItem {
	xmlBytes, err := getXML(url)
	if err != nil {
		fmt.Printf("Error during getXml: %s", err)
	}
	rssFeed := rss{}
	err = xml.Unmarshal(xmlBytes, &rssFeed)
	if err != nil {
		fmt.Printf("Error during Unmarshal: %s", err)
	}

	return rssFeed.Channel.Items
}

func getXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}
