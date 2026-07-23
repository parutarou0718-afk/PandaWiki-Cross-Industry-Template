package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

// FeedItem represents a single item in any feed format
// FeedItem 表示任意Feed格式中的单个条目
// 字段说明：
// Title: 条目标题
// Link: 条目链接（URL）
// Description: 条目描述内容
// Published: 发布时间（字符串格式，具体格式由Feed源决定）
type FeedItem struct {
	Title       string // 条目标题
	Link        string // 条目链接URL
	Description string // 条目描述内容
	Published   string // 发布时间（字符串格式）
}

// Feed represents a generic feed structure
type Feed struct {
	Title       string
	Description string
	Link        string
	Items       []FeedItem
}

// cleanXMLContent removes illegal XML characters from the content
func cleanXMLContent(content string) string {
	return strings.Map(func(r rune) rune {
		// Check if the character is a valid XML character
		// XML 1.0 spec: https://www.w3.org/TR/xml/#charsets
		if r == 0x9 || r == 0xA || r == 0xD || (r >= 0x20 && r <= 0xD7FF) || (r >= 0xE000 && r <= 0xFFFD) || (r >= 0x10000 && r <= 0x10FFFF) {
			return r
		}
		return -1 // Remove invalid characters
	}, content)
}

// ParseFeed 解析指定URL的Feed内容，返回通用Feed结构
// 参数：
// url: 要解析的Feed内容URL
// 返回值：
// *Feed: 解析后的通用Feed结构（包含标题、描述、链接和条目列表）
// error: 解析过程中出现的错误（网络错误、格式不支持等）
func ParseFeed(url string) (*Feed, error) {
	// Get feed content
	content, err := HTTPGet(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed content: %v", err)
	}

	// Decode content
	decoded := DecodeBytes(content)
	// Clean illegal XML characters
	cleaned := cleanXMLContent(decoded)
	decodedBytes := []byte(cleaned)

	// Try to detect feed format and parse accordingly
	if strings.Contains(cleaned, "<rss") {
		return parseRSS(decodedBytes)
	} else if strings.Contains(cleaned, "<feed") {
		return parseAtom(decodedBytes)
	} else if strings.Contains(cleaned, "\"version\":") {
		return parseJSONFeed(decodedBytes)
	}

	return nil, fmt.Errorf("unsupported feed format")
}

// parseRSS 解析RSS格式（如RSS 2.0）的内容
// 参数：content - RSS格式的字节内容
// 返回值：解析后的通用Feed结构或错误
// 注意：处理链接时按以下优先级获取：link标签的href属性 > link标签文本值 > Atom扩展链接 > Guid（永久链接）
func parseRSS(content []byte) (*Feed, error) {
	type RSSFeed struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Link        string `xml:"link"`
			AtomLink    struct {
				Href string `xml:"href,attr"`
			} `xml:"http://www.w3.org/2005/Atom link"`
			Items []struct {
				Title string `xml:"title"`
				Links []struct {
					Href  string `xml:"href,attr"`
					Value string `xml:",chardata"`
				} `xml:"link"`
				Description string `xml:"description"`
				PubDate     string `xml:"pubDate"`
				Guid        struct {
					IsPermaLink string `xml:"isPermaLink,attr"`
					Value       string `xml:",chardata"`
				} `xml:"guid"`
				AtomLink struct {
					Href string `xml:"href,attr"`
				} `xml:"http://www.w3.org/2005/Atom link"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	var rssFeed RSSFeed
	if err := xml.Unmarshal(content, &rssFeed); err != nil {
		return nil, fmt.Errorf("failed to parse RSS: %v", err)
	}

	feed := &Feed{
		Title:       rssFeed.Channel.Title,
		Description: rssFeed.Channel.Description,
		Link:        rssFeed.Channel.Link,
		Items:       make([]FeedItem, 0),
	}

	for _, item := range rssFeed.Channel.Items {
		feedItem := FeedItem{
			Title:       item.Title,
			Description: item.Description,
			Published:   item.PubDate,
		}

		// Try to get link from various sources in order of preference
		if len(item.Links) > 0 {
			// Try href attribute first, then value
			if item.Links[0].Href != "" {
				feedItem.Link = item.Links[0].Href
			} else if item.Links[0].Value != "" {
				feedItem.Link = item.Links[0].Value
			}
		} else if item.AtomLink.Href != "" {
			feedItem.Link = item.AtomLink.Href
		} else if item.Guid.Value != "" && (item.Guid.IsPermaLink == "" || item.Guid.IsPermaLink == "true") {
			feedItem.Link = item.Guid.Value
		}

		feed.Items = append(feed.Items, feedItem)
	}

	return feed, nil
}

// parseAtom 解析Atom 1.0格式的内容
// 参数：content - Atom格式的字节内容
// 返回值：解析后的通用Feed结构或错误
// 注意：Feed链接取第一个link元素的href属性（建议优先使用rel="alternate"的链接）
func parseAtom(content []byte) (*Feed, error) {
	type AtomFeed struct {
		XMLName  xml.Name `xml:"feed"`
		Title    string   `xml:"title"`
		Subtitle string   `xml:"subtitle"`
		Link     []struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Entries []struct {
			Title string `xml:"title"`
			Link  []struct {
				Href string `xml:"href,attr"`
			} `xml:"link"`
			Summary string `xml:"summary"`
			Updated string `xml:"updated"`
		} `xml:"entry"`
	}

	var atomFeed AtomFeed
	if err := xml.Unmarshal(content, &atomFeed); err != nil {
		return nil, fmt.Errorf("failed to parse Atom: %v", err)
	}

	feed := &Feed{
		Title:       atomFeed.Title,
		Description: atomFeed.Subtitle,
		Items:       make([]FeedItem, 0),
	}

	if len(atomFeed.Link) > 0 {
		feed.Link = atomFeed.Link[0].Href
	}

	for _, entry := range atomFeed.Entries {
		item := FeedItem{
			Title:       entry.Title,
			Description: entry.Summary,
			Published:   entry.Updated,
		}
		if len(entry.Link) > 0 {
			item.Link = entry.Link[0].Href
		}
		feed.Items = append(feed.Items, item)
	}

	return feed, nil
}

// parseJSONFeed 解析JSON Feed格式（如1.1版本）的内容
// 参数：content - JSON Feed格式的字节内容
// 返回值：解析后的通用Feed结构或错误
// 字段映射：home_page_url -> Feed.Link; date_published -> FeedItem.Published
func parseJSONFeed(content []byte) (*Feed, error) {
	type JSONFeed struct {
		Version     string `json:"version"`
		Title       string `json:"title"`
		Description string `json:"description"`
		HomePageURL string `json:"home_page_url"`
		Items       []struct {
			Title         string `json:"title"`
			URL           string `json:"url"`
			ContentText   string `json:"content_text"`
			DatePublished string `json:"date_published"`
		} `json:"items"`
	}

	var jsonFeed JSONFeed
	if err := json.Unmarshal(content, &jsonFeed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON Feed: %v", err)
	}

	feed := &Feed{
		Title:       jsonFeed.Title,
		Description: jsonFeed.Description,
		Link:        jsonFeed.HomePageURL,
		Items:       make([]FeedItem, 0),
	}

	for _, item := range jsonFeed.Items {
		feed.Items = append(feed.Items, FeedItem{
			Title:       item.Title,
			Link:        item.URL,
			Description: item.ContentText,
			Published:   item.DatePublished,
		})
	}

	return feed, nil
}
