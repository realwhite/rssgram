package internal

import (
	"fmt"
	"math/rand/v2"
	"mime"
	"net/http"
	"time"

	"golang.org/x/net/html"

	"github.com/go-shiori/dom"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
}

type SiteDescription struct {
	Title       string
	Description string
	Image       string
}

type SiteParser struct {
}

func (p *SiteParser) GetDescription(url string) (SiteDescription, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return SiteDescription{}, fmt.Errorf("failed to  create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgents[rand.IntN(len(userAgents))])

	resp, err := client.Do(req)
	if err != nil {
		return SiteDescription{}, fmt.Errorf("failed to get content by url %s: %w", url, err)
	} else if resp.StatusCode > http.StatusPermanentRedirect {
		return SiteDescription{}, fmt.Errorf("failed to get content by url %s: status %s", url, resp.Status)
	}

	defer resp.Body.Close()

	mediaType, _, err := mime.ParseMediaType(resp.Header["Content-Type"][0])
	if err != nil {
		return SiteDescription{}, fmt.Errorf("failed to parse media type: %w", err)
	}

	if mediaType != "text/html" {
		return SiteDescription{}, nil
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return SiteDescription{}, fmt.Errorf("failed to parse html: %w", err)
	}

	var title, description, image string

	// get site title
	titleItem := dom.QuerySelector(doc, "title")
	if titleItem != nil {
		title = dom.InnerText(titleItem)
	}

	// get site description
	regularDescriptionItem := dom.QuerySelector(doc, "meta[name=description]")
	if regularDescriptionItem != nil {
		description = dom.GetAttribute(regularDescriptionItem, "content")
	}

	if description == "" {
		ogDescriptionItem := dom.QuerySelector(doc, `meta[property="og:description"]`)
		if ogDescriptionItem != nil {
			description = dom.GetAttribute(ogDescriptionItem, "content")
		}
	}

	// get image
	regularImageItem := dom.QuerySelector(doc, "meta[name=image]")
	if regularImageItem != nil {
		image = dom.GetAttribute(regularImageItem, "content")
	}

	if image == "" {
		ogImageItem := dom.QuerySelector(doc, `meta[property="og:image"]`)
		if ogImageItem != nil {
			image = dom.GetAttribute(ogImageItem, "content")
		}
	}

	return SiteDescription{
		title,
		description,
		image,
	}, nil
}

// https://xnacly.me/posts/2024/extract-metadata-from-html/

func NewSiteParser() *SiteParser {
	return &SiteParser{}
}
