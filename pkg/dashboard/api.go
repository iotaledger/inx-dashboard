package dashboard

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"

	"github.com/iotaledger/inx-dashboard/pkg/common"
	"github.com/iotaledger/inx-dashboard/pkg/jwt"
)

const (
	WebsocketCmdRegister   = 0
	WebsocketCmdUnregister = 1
)

func compileRouteAsRegex(route string) *regexp.Regexp {

	r := regexp.QuoteMeta(route)
	r = strings.Replace(r, `\*`, "(.*?)", -1)
	r = r + "$"

	reg, err := regexp.Compile(r)
	if err != nil {
		return nil
	}

	return reg
}

func compileRoutesAsRegexes(routes []string) []*regexp.Regexp {
	var regexes []*regexp.Regexp
	for _, route := range routes {
		reg := compileRouteAsRegex(route)
		if reg == nil {
			panic(fmt.Sprintf("Invalid route in config: %s", route))
		}
		regexes = append(regexes, reg)
	}

	return regexes
}

func (d *Dashboard) devModeReverseProxyMiddleware() echo.MiddlewareFunc {

	apiURL, err := url.Parse(d.developerModeURL)
	if err != nil {
		d.LogFatalfAndExit("wrong devmode url: %s", err)
	}

	return middleware.Proxy(middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: apiURL,
		},
	}))
}

func (d *Dashboard) apiMiddlewares() []echo.MiddlewareFunc {

	// the HTTP REST routes which can be called without authorization.
	// Wildcards using * are allowed
	publicRoutes := []string{
		"/api/routes",
		"/api/core/v2/info",
		"/api/core/v2/blocks*",
		"/api/core/v2/transactions*",
		"/api/core/v2/milestones*",
		"/api/core/v2/outputs*",
		"/api/indexer/v1/*",
	}

	// the HTTP REST routes which need to be called with authorization.
	// Wildcards using * are allowed
	protectedRoutes := []string{
		"/api/core/v2/peers*",
		"/api/*",
	}

	publicRoutesRegEx := compileRoutesAsRegexes(publicRoutes)
	protectedRoutesRegEx := compileRoutesAsRegexes(protectedRoutes)

	matchPublic := func(c echo.Context) bool {
		loweredPath := strings.ToLower(c.Request().URL.EscapedPath())

		for _, reg := range publicRoutesRegEx {
			if reg.MatchString(loweredPath) {
				return true
			}
		}

		return false
	}

	matchProtected := func(c echo.Context) bool {
		loweredPath := strings.ToLower(c.Request().URL.EscapedPath())

		for _, reg := range protectedRoutesRegEx {
			if reg.MatchString(loweredPath) {
				return true
			}
		}

		return false
	}

	// Skip routes explicitly matching the publicRoutes, or not matching the protectedRoutes
	jwtAuthSkipper := func(c echo.Context) bool {
		return matchPublic(c) || !matchProtected(c)
	}

	jwtAllow := func(c echo.Context, subject string, claims *jwt.AuthClaims) bool {
		if d.authUserName == "" {
			return false
		}

		return claims.VerifySubject(d.authUserName)
	}

	return []echo.MiddlewareFunc{
		d.jwtAuth.Middleware(jwtAuthSkipper, jwtAllow),
	}
}

func (d *Dashboard) authRoute(c echo.Context) error {

	type loginRequest struct {
		JWT      string `json:"jwt"`
		User     string `json:"user"`
		Password string `json:"password"`
	}

	request := &loginRequest{}

	if err := c.Bind(request); err != nil {
		return errors.WithMessagef(common.ErrInvalidParameter, "invalid request, error: %s", err)
	}

	if len(request.JWT) > 0 {
		// Verify JWT is still valid
		if !d.jwtAuth.VerifyJWT(request.JWT, func(claims *jwt.AuthClaims) bool {
			return true
		}) {
			return echo.ErrUnauthorized
		}
	} else if !d.basicAuth.VerifyUsernameAndPassword(request.User, request.Password) {
		return echo.ErrUnauthorized
	}

	t, err := d.jwtAuth.IssueJWT()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"jwt": t,
	})
}

func (d *Dashboard) setupRoutes(e *echo.Echo) {

	e.Use(middleware.CSRF())

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusPermanentRedirect, "/dashboard/")
	})
	e.GET("/dashboard", func(c echo.Context) error {
		return c.Redirect(http.StatusPermanentRedirect, "/dashboard/")
	})

	mw := frontendMiddleware()
	if d.developerMode {
		mw = d.devModeReverseProxyMiddleware()
	}
	e.Group("/dashboard/*").Use(mw)

	// Pass all the dashboard request through to the local rest API
	err := d.setupAPIRoutes(e.Group("/dashboard/api", d.apiMiddlewares()...))
	if err != nil {
		d.LogPanicf("failed to setup node routes: %w", err)
	}

	e.GET("/dashboard/ws", d.websocketRoute)

	// Rate-limit the auth endpoint
	rateLimiterConfig := middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(20 / 60.0), // 20 request every 1 minute
				Burst:     30,                    // additional burst of 30 requests
				ExpiresIn: 5 * time.Minute,
			},
		),
	}

	e.POST("/dashboard/auth", d.authRoute, middleware.RateLimiterWithConfig(rateLimiterConfig))
}
