package encoding

import (
	"bytes"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestRSSEncoder(t *testing.T) {
	subject := bytes.NewBuffer([]byte(`<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">

<channel>
  <title>W3Schools Home Page</title>
  <link>https://www.w3schools.com</link>
  <description>Free web building tutorials</description>
  <image>
    <url>https://www.w3schools.com/pretty_image.png</url>
    <description>Free web building tutorials</description>
  </image>
  <language>en_us</language>
  <item>
    <title>RSS Tutorial</title>
    <link>https://www.w3schools.com/xml/xml_rss.asp</link>
    <description>New RSS tutorial on W3Schools</description>
  </item>
  <item>
    <title>XML Tutorial</title>
    <link>https://www.w3schools.com/xml</link>
    <description>New XML tutorial on W3Schools</description>
  </item>
</channel>

</rss>`))

	decoder := NewRSSDecoder()

	var result map[string]interface{}

	if err := decoder(subject, &result); err != nil {
		t.Error(err)
	}

	if result["type"] != "rss" {
		t.Error("Error unexpected result type:", result["type"])
	}

	if result["description"] != "Free web building tutorials" {
		t.Error("Error unexpected description:", result["description"])
	}

	if result["language"] != "en_us" {
		t.Error("Error unexpected language:", result["language"])
	}

	if result["img_url"] != "https://www.w3schools.com/pretty_image.png" {
		t.Error("Error unexpected image url:", result["img_url"])
	}

	if len(result["items"].([]*gofeed.Item)) != 2 {
		t.Error("Error unexpected number of result items", result["items"])
	}

}

func TestRSSEncoder_ko(t *testing.T) {
	decoder := NewRSSDecoder()

	var result map[string]interface{}

	if err := decoder(bytes.NewBuffer([]byte(``)), &result); err == nil {
		t.Error("The decoder didn't return an error")
	}

}
