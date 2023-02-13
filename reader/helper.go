package reader

import (
	"fmt"
	"net/http"
)

const testData = `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
 <title>RSS Feed's Title</title>
 <link>http://www.rssfeedslink.com</link>
 <item>
  <title>First Item</title>
  <description>Description Of first Item</description>
  <link>http://www.example.com/blog/post/1</link>
 </item>
 <item>
  <title>Second Item</title>
  <description>Description Of Second Item</description>
  <link>http://www.example.com/blog/post/2</link>
 </item>

</channel>
</rss>`

var (
	port = "8080"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, testData)
}

func buildUrl(path string) string {
	return urlFor("http", port, path)
}

func urlFor(scheme string, serverPort string, path string) string {
	return scheme + "://localhost:" + serverPort + path
}
