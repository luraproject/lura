package encoding

import (
	"io"

	"github.com/mmcdole/gofeed"
)

// RSS is the key for the rss encoding
const RSS = "rss"

// NewRSSDecoder returns the RSS decoder
func NewRSSDecoder() Decoder {
	fp := gofeed.NewParser()
	return func(r io.Reader, v *map[string]interface{}) error {
		feed, err := fp.Parse(r)
		if err != nil {
			return err
		}
		*(v) = map[string]interface{}{
			"items":       feed.Items,
			"author":      feed.Author,
			"categories":  feed.Categories,
			"custom":      feed.Custom,
			"copyright":   feed.Copyright,
			"description": feed.Description,
			"type":        feed.FeedType,
			"language":    feed.Language,
			"title":       feed.Title,
			"published":   feed.Published,
			"updated":     feed.Updated,
		}
		if feed.Image != nil {
			(*v)["img_url"] = feed.Image.URL
		}
		return nil
	}
}
