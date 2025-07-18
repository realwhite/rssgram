package feed

import (
	"fmt"
	"math/rand/v2"
	"mime"
	"net/http"
	"strings"
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
	client *http.Client
}

func (p *SiteParser) setUA(req *http.Request) {
	if len(userAgents) > 0 {
		req.Header.Set("User-Agent", userAgents[rand.IntN(len(userAgents))])
	}
}

func (p *SiteParser) GetDescription(url string) (SiteDescription, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return SiteDescription{}, fmt.Errorf("failed to  create request: %w", err)
	}

	p.setUA(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return SiteDescription{}, fmt.Errorf("failed to get content by url %s: %w", url, err)
	} else if resp.StatusCode > http.StatusPermanentRedirect {
		return SiteDescription{}, fmt.Errorf("failed to get content by url %s: status %s", url, resp.Status)
	}

	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	contentTypes := resp.Header["Content-Type"]
	if len(contentTypes) == 0 {
		return SiteDescription{}, fmt.Errorf("Content-Type header is missing")
	}
	mediaType, _, err := mime.ParseMediaType(contentTypes[0])
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

	result := SiteDescription{
		Title:       title,
		Description: description,
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

	if image != "" {
		isValid, _ := p.isImageURLValid(image)

		if isValid {
			result.Image = image
		}

	}

	return result, nil
}

func (p *SiteParser) isImageURLValid(url string) (bool, error) {
	if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "https") {
		return false, nil
	}

	_, mimeType, err := p.getMimeTypeByUrl(url)
	if err != nil {
		return false, fmt.Errorf("failed to get content by url %s: %w", url, err)
	}

	if strings.Contains(mimeType, "image") {
		return true, nil
	}

	return false, nil

}

func (p *SiteParser) makeReq(method, url string) (int, string, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, "", err
	}

	p.setUA(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, "", err
	}

	defer resp.Body.Close()

	return resp.StatusCode, resp.Header.Get("Content-Type"), err
}

func (p *SiteParser) getMimeTypeByUrl(url string) (int, string, error) {
	statusCode, contType, err := p.makeReq(http.MethodHead, url)
	if err != nil {
		return statusCode, "", err
	}

	if statusCode == 405 {
		statusCode, contType, err = p.makeReq(http.MethodGet, url)
		if err != nil {
			return statusCode, contType, err
		}
	}
	return statusCode, contType, err
}

// https://xnacly.me/posts/2024/extract-metadata-from-html/

func NewSiteParser() *SiteParser {
	client := http.Client{
		Timeout: time.Second * 5,
	}
	return &SiteParser{
		client: &client,
	}
}
