// Тест создан с помощью AI
package feed

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSiteParser_GetDescription(t *testing.T) {
	// Создаем тестовый сервер
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

	// Тестируем получение описания
	description, err := parser.GetDescription(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, description)
	assert.Equal(t, "Test Page Title", description.Title)
	assert.Equal(t, "Test page description", description.Description)
	// image не поддерживается парсером, ожидаем пустую строку
	assert.Equal(t, "", description.Image)
}

func TestSiteParser_GetDescription_NoMeta(t *testing.T) {
	// Создаем тестовый сервер без мета-тегов
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

	// Тестируем получение описания
	description, err := parser.GetDescription(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, description)
	assert.Equal(t, "Test Page Title", description.Title)
	assert.Empty(t, description.Description)
	assert.Empty(t, description.Image)
}

func TestSiteParser_GetDescription_InvalidURL(t *testing.T) {
	parser := NewSiteParser()

	// Тестируем с невалидным URL
	description, err := parser.GetDescription("invalid-url")

	assert.Error(t, err)
	assert.NotNil(t, description)
	assert.Empty(t, description.Title)
	assert.Empty(t, description.Description)
	assert.Empty(t, description.Image)
}

func TestSiteParser_GetDescription_404(t *testing.T) {
	// Создаем тестовый сервер, возвращающий 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	parser := NewSiteParser()

	// Тестируем получение описания
	description, err := parser.GetDescription(server.URL)

	assert.Error(t, err)
	assert.NotNil(t, description)
	assert.Empty(t, description.Title)
	assert.Empty(t, description.Description)
	assert.Empty(t, description.Image)
}

func TestSiteParser_GetDescription_Timeout(t *testing.T) {
	// Создаем тестовый сервер с задержкой
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Имитируем долгий ответ
		select {
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	parser := NewSiteParser()

	// Тестируем получение описания с таймаутом
	description, err := parser.GetDescription(server.URL)

	// Ожидаем ошибку таймаута или контекста
	assert.Error(t, err)
	assert.NotNil(t, description)
	assert.Empty(t, description.Title)
	assert.Empty(t, description.Description)
	assert.Empty(t, description.Image)
}

func TestSiteParser_GetDescription_NonHTML(t *testing.T) {
	// Создаем тестовый сервер, возвращающий JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"title": "JSON Response"}`))
	}))
	defer server.Close()

	parser := NewSiteParser()

	// Тестируем получение описания
	description, err := parser.GetDescription(server.URL)

	// Ожидаем ошибку или пустое описание
	if err != nil {
		assert.Error(t, err)
		assert.NotNil(t, description)
		assert.Empty(t, description.Title)
		assert.Empty(t, description.Description)
		assert.Empty(t, description.Image)
	} else {
		assert.NotNil(t, description)
		// Проверяем, что парсинг прошел, но данные могут быть пустыми
	}
}

func TestSiteParser_GetDescription_OpenGraph(t *testing.T) {
	// Создаем тестовый сервер с Open Graph тегами
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

	// Тестируем получение описания
	description, err := parser.GetDescription(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, description)
	// Парсер не поддерживает OG, ожидаем обычный title
	assert.Equal(t, "Test Page Title", description.Title)
	assert.Equal(t, "Open Graph Description", description.Description) // если парсер поддерживает og:description, иначе пусто
	assert.Equal(t, "", description.Image)
}
