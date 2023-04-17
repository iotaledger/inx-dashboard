package dashboard

import (
	"time"

	"go.uber.org/dig"
	"nhooyr.io/websocket"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/web/websockethub"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/inx-dashboard/pkg/dashboard"
)

const (
	broadcastQueueSize            = 20000
	clientSendChannelSize         = 1000
	webSocketWriteTimeout         = time.Duration(5) * time.Second
	maxWebsocketMessageSize int64 = 400 + maxDashboardAuthUsernameSize + 10 // 10 buffer due to variable JWT lengths
)

func init() {
	Component = &app.Component{
		Name:      "Dashboard",
		DepsFunc:  func(cDeps dependencies) { deps = cDeps },
		Params:    params,
		Provide:   provide,
		Configure: configure,
		Run:       run,
	}
}

var (
	Component *app.Component
	deps      dependencies
)

type dependencies struct {
	dig.In
	Dashboard *dashboard.Dashboard
}

func provide(c *dig.Container) error {

	type dashboardDeps struct {
		dig.In
		NodeBridge *nodebridge.NodeBridge
	}

	if err := c.Provide(func(deps dashboardDeps) *dashboard.Dashboard {

		username := ParamsDashboard.Auth.Username
		if len(username) == 0 {
			Component.LogErrorfAndExit("%s cannot be empty", Component.App().Config().GetParameterPath(&(ParamsDashboard.Auth.Username)))
		}
		if len(username) > maxDashboardAuthUsernameSize {
			Component.LogErrorfAndExit("%s has a max length of %d", Component.App().Config().GetParameterPath(&(ParamsDashboard.Auth.Username)), maxDashboardAuthUsernameSize)
		}

		acceptOptions := &websocket.AcceptOptions{
			InsecureSkipVerify: true, // allow any origin for websocket connections
			// Disable compression due to incompatibilities with latest Safari browsers:
			// https://github.com/tilt-dev/tilt/issues/4746
			CompressionMode: websocket.CompressionDisabled,
		}

		hub := websockethub.NewHub(Component.Logger(), acceptOptions, broadcastQueueSize, clientSendChannelSize, maxWebsocketMessageSize)

		Component.LogInfo("Setting up dashboard ...")

		return dashboard.New(
			Component.Logger(),
			Component.Daemon(),
			deps.NodeBridge,
			hub,
			dashboard.WithBindAddress(ParamsDashboard.BindAddress),
			dashboard.WithDeveloperMode(ParamsDashboard.DeveloperMode),
			dashboard.WithDeveloperModeURL(ParamsDashboard.DeveloperModeURL),
			dashboard.WithAuthUsername(ParamsDashboard.Auth.Username),
			dashboard.WithAuthPasswordHash(ParamsDashboard.Auth.PasswordHash),
			dashboard.WithAuthPasswordSalt(ParamsDashboard.Auth.PasswordSalt),
			dashboard.WithAuthSessionTimeout(ParamsDashboard.Auth.SessionTimeout),
			dashboard.WithAuthIdentityFilePath(ParamsDashboard.Auth.IdentityFilePath),
			dashboard.WithAuthIdentityPrivateKey(ParamsDashboard.Auth.IdentityPrivateKey),
			dashboard.WithAuthRateLimitEnabled(ParamsDashboard.Auth.RateLimit.Enabled),
			dashboard.WithAuthRateLimitPeriod(ParamsDashboard.Auth.RateLimit.Period),
			dashboard.WithAuthRateLimitMaxRequests(ParamsDashboard.Auth.RateLimit.MaxRequests),
			dashboard.WithAuthRateLimitMaxBurst(ParamsDashboard.Auth.RateLimit.MaxBurst),
			dashboard.WithWebsocketWriteTimeout(webSocketWriteTimeout),
			dashboard.WithDebugLogRequests(ParamsDashboard.DebugRequestLoggerEnabled),
		)
	}); err != nil {
		return err
	}

	return nil
}

func configure() error {
	deps.Dashboard.Init()

	return nil
}

func run() error {
	deps.Dashboard.Run()

	return nil
}
