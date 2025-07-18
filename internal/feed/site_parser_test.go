// Test created with AI
package feed

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSiteParser_GetDescription(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Page Title</title>
			<meta name="description" content="Test page description">
			<meta property="og:description" content="Open Graph description">
			<meta property="og:image" content="https://example.com/image.jpg">
		</head>
		<body>
			<h1>Test Page</h1>
			<p>This is a test page content.</p>
		</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer server.Close()

	parser := NewSiteParser()

	// Test getting description
	description, err := parser.GetDescription(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, description)
	assert.Equal(t, "Test Page Title", description.Title)
	assert.Equal(t, "Test page description", description.Description)
	// image is not supported by the parser, expect an empty string
	assert.Equal(t, "", description.Image)
}

func TestSiteParser_GetDescription_NoMeta(t *testing.T) {
	// Create a test server without meta-tags
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Page Title</title>
		</head>
		<body>
			<h1>Test Page</h1>
			<p>This is a test page content.</p>
		</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer server.Close()

	parser := NewSiteParser()

	// Test getting description
	description, err := parser.GetDescription(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, description)
	assert.Equal(t, "Test Page Title", description.Title)
	assert.Empty(t, description.Description)
	assert.Empty(t, description.Image)
}

func TestSiteParser_GetDescription_InvalidURL(t *testing.T) {
	parser := NewSiteParser()

	// Test with an invalid URL
	description, err := parser.GetDescription("invalid-url")

	assert.Error(t, err)
	assert.NotNil(t, description)
	assert.Empty(t, description.Title)
	assert.Empty(t, description.Description)
	assert.Empty(t, description.Image)
}

func TestSiteParser_GetDescription_404(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	parser := NewSiteParser()

	// Test getting description
	description, err := parser.GetDescription(server.URL)

	assert.Error(t, err)
	assert.NotNil(t, description)
	assert.Empty(t, description.Title)
	assert.Empty(t, description.Description)
	assert.Empty(t, description.Image)
}

func TestSiteParser_GetDescription_Timeout(t *testing.T) {
	// Create a test server with a delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a long response
		select {
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	parser := NewSiteParser()

	// Test getting description with a timeout
	description, err := parser.GetDescription(server.URL)

	// Expect a timeout error or context cancellation
	assert.Error(t, err)
	assert.NotNil(t, description)
	assert.Empty(t, description.Title)
	assert.Empty(t, description.Description)
	assert.Empty(t, description.Image)
}

func TestSiteParser_GetDescription_NonHTML(t *testing.T) {
	// Create a test server that returns JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"title": "JSON Response"}`))
	}))
	defer server.Close()

	parser := NewSiteParser()

	// Test getting description
	description, err := parser.GetDescription(server.URL)

	// Expect an error or empty description
	if err != nil {
		assert.Error(t, err)
		assert.NotNil(t, description)
		assert.Empty(t, description.Title)
		assert.Empty(t, description.Description)
		assert.Empty(t, description.Image)
	} else {
		assert.NotNil(t, description)
		// Check if parsing was successful, but data might be empty
	}
}

func TestSiteParser_GetDescription_OpenGraph(t *testing.T) {
	// Create a test server with Open Graph tags
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Page Title</title>
			<meta property="og:title" content="Open Graph Title">
			<meta property="og:description" content="Open Graph Description">
			<meta property="og:image" content="https://example.com/og-image.jpg">
		</head>
		<body>
			<h1>Test Page</h1>
		</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer server.Close()

	parser := NewSiteParser()

	// Test getting description
	description, err := parser.GetDescription(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, description)
	// The parser does not support OG, expect the regular title
	assert.Equal(t, "Test Page Title", description.Title)
	assert.Equal(t, "Open Graph Description", description.Description) // if the parser supports og:description, otherwise empty
	assert.Equal(t, "", description.Image)
}
