package reader

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
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
	Title       string    `xml:"title" json:"title"`
	Source      string    `xml:"source" json:"source"`
	SourceURL   string    `xml:"source_url" json:"source_url"`
	Link        string    `xml:"link" json:"link"`
	PublishDate time.Time `xml:"pubdate" json:"publish_date"`
	Description string    `xml:"description" json:"description"`
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
