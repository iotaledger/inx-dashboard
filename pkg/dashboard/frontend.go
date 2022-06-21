package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

//go:embed frontend
var frontendFiles embed.FS

func distFileSystem() http.FileSystem {
	f, err := fs.Sub(frontendFiles, "dist")
	if err != nil {
		panic(err)
	}
	return http.FS(f)
}

func calculateMimeType(e echo.Context) string {
	url := e.Request().URL.String()

	switch {
	case strings.HasSuffix(url, ".html"):
		return echo.MIMETextHTMLCharsetUTF8
	case strings.HasSuffix(url, ".css"):
		return "text/css"
	case strings.HasSuffix(url, ".js"):
		return echo.MIMEApplicationJavaScript
	case strings.HasSuffix(url, ".json"):
		return echo.MIMEApplicationJSONCharsetUTF8
	case strings.HasSuffix(url, ".png"):
		return "image/png"
	case strings.HasSuffix(url, ".svg"):
		return "image/svg+xml"
	default:
		return echo.MIMETextHTMLCharsetUTF8
	}
}

func frontendMiddleware() echo.MiddlewareFunc {
	fs := distFileSystem()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			contentType := calculateMimeType(c)

			path := strings.TrimPrefix(c.Request().URL.Path, "/")
			if len(path) == 0 {
				path = "index.html"
				contentType = echo.MIMETextHTMLCharsetUTF8
			}

			staticBlob, err := fs.Open(path)
			if err != nil {
				// If the asset cannot be found, fall back to the index.html for routing
				path = "index.html"
				contentType = echo.MIMETextHTMLCharsetUTF8
				staticBlob, err = fs.Open(path)
				if err != nil {
					return next(c)
				}
			}
			return c.Stream(http.StatusOK, contentType, staticBlob)
		}
	}
}
