package dashboard

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"

	dashboard "github.com/iotaledger/hornet-dashboard"
	"github.com/iotaledger/inx-dashboard/pkg/common"
	"github.com/iotaledger/inx-dashboard/pkg/jwt"
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

/*
func (d *Dashboard) apiMiddleware() echo.MiddlewareFunc {

	publicRoutesRegEx := compileRoutesAsRegexes(ParamsRestAPI.PublicRoutes)
	protectedRoutesRegEx := compileRoutesAsRegexes(ParamsRestAPI.ProtectedRoutes)

	matchPublic := func(c echo.Context) bool {
		loweredPath := strings.ToLower(c.Path())

		for _, reg := range publicRoutesRegEx {
			if reg.MatchString(loweredPath) {
				return true
			}
		}
		return false
	}

	matchExposed := func(c echo.Context) bool {
		loweredPath := strings.ToLower(c.Path())

		for _, reg := range append(publicRoutesRegEx, protectedRoutesRegEx...) {
			if reg.MatchString(loweredPath) {
				return true
			}
		}
		return false
	}

	jwtAllow := func(c echo.Context, subject string, claims *jwt.AuthClaims) bool {
		if d.authUserName == "" {
			return false
		}
		return claims.VerifySubject(d.authUserName) && d.dashboardAllowedAPIRoute(c)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {

		// Skip routes matching the publicRoutes
		publicSkipper := func(c echo.Context) bool {
			return matchPublic(c)
		}

		jwtMiddlewareHandler := d.jwtAuth.Middleware(publicSkipper, jwtAllow)(next)

		return func(c echo.Context) error {

			// Check if the route should be exposed (public or protected) or is required by the dashboard
			if matchExposed(c) || d.dashboardAllowedAPIRoute(c) {
				// Apply JWT middleware
				return jwtMiddlewareHandler(c)
			}

			return echo.ErrForbidden
		}
	}
}
*/
var dashboardAllowedRoutes = map[string][]string{
	http.MethodGet: {
		"/api/v2/info",
		"/api/v2/blocks",
		"/api/v2/transactions",
		"/api/v2/milestones",
		"/api/v2/outputs",
		"/api/v2/peers",
		"/api/plugins/indexer/v1",
		"/api/plugins/spammer/v1",
		"/api/plugins/participation/v1/events",
	},
	http.MethodPost: {
		"/api/v2/peers",
		"/api/plugins/spammer/v1",
		"/api/plugins/participation/v1/admin/events",
	},
	http.MethodDelete: {
		"/api/v2/peers",
		"/api/plugins/participation/v1/admin/events",
	},
}

func (d *Dashboard) checkAllowedAPIRoute(context echo.Context, allowedRoutes map[string][]string) bool {

	// Check for which route we will allow to access the API
	routesForMethod, exists := allowedRoutes[context.Request().Method]
	if !exists {
		return false
	}

	path := context.Request().URL.EscapedPath()
	for _, prefix := range routesForMethod {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func (d *Dashboard) dashboardAllowedAPIRoute(context echo.Context) bool {
	return d.checkAllowedAPIRoute(context, dashboardAllowedRoutes)
}

const (
	WebsocketCmdRegister   = 0
	WebsocketCmdUnregister = 1
)

func (d *Dashboard) devModeReverseProxyMiddleware() echo.MiddlewareFunc {

	apiURL, err := url.Parse("http://127.0.0.1:9090")
	if err != nil {
		d.LogFatalf("wrong devmode url: %s", err)
	}

	return middleware.Proxy(middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: apiURL,
		},
	}))
}

func (d *Dashboard) apiMiddlewares() []echo.MiddlewareFunc {

	apiBindAddr := "localhost:14265"
	_, apiBindPort, err := net.SplitHostPort(apiBindAddr)
	if err != nil {
		d.LogFatalf("wrong REST API bind address: %s", err)
	}

	apiURL, err := url.Parse(fmt.Sprintf("http://localhost:%s", apiBindPort))
	if err != nil {
		d.LogFatalf("wrong dashboard API url: %s", err)
	}

	proxySkipper := func(context echo.Context) bool {
		// Only proxy allowed routes, skip all others
		return !d.dashboardAllowedAPIRoute(context)
	}

	balancer := middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: apiURL,
		},
	})

	proxyConfig := middleware.ProxyConfig{
		Skipper:  proxySkipper,
		Balancer: balancer,
	}

	// the HTTP REST routes which can be called without authorization.
	// Wildcards using * are allowed
	publicRoutes := []string{
		"/api/plugins/indexer/v1/*",
	}

	// the HTTP REST routes which need to be called with authorization even if the API is not protected.
	// Wildcards using * are allowed
	protectedRoutes := []string{
		"/api/v2/peers*",
		"/api/plugins/*",
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

	// Skip routes explicitely matching the publicRoutes, or not matching the protectedRoutes
	jwtAuthSkipper := func(c echo.Context) bool {
		return matchPublic(c) || !matchProtected(c)
	}

	jwtAllow := func(_ echo.Context, subject string, claims *jwt.AuthClaims) bool {
		return claims.VerifySubject(subject)
	}

	return []echo.MiddlewareFunc{
		d.jwtAuth.Middleware(jwtAuthSkipper, jwtAllow),
		middleware.ProxyWithConfig(proxyConfig),
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

	mw := dashboard.FrontendMiddleware()
	if d.developerMode {
		mw = d.devModeReverseProxyMiddleware()
	}
	e.Group("/*").Use(mw)

	// Pass all the dashboard request through to the local rest API
	e.Group("/api", d.apiMiddlewares()...)

	e.GET("/dashboard/ws", d.websocketRoute)

	// Rate-limit the auth endpoint
	rateLimiterConfig := middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(5 / 60.0), // 5 request every 1 minute
				Burst:     10,                   // additional burst of 10 requests
				ExpiresIn: 5 * time.Minute,
			},
		),
	}

	e.POST("/dashboard/auth", d.authRoute, middleware.RateLimiterWithConfig(rateLimiterConfig))
}
