package dashboard

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/inx-app/nodebridge"
	"github.com/iotaledger/inx-dashboard/pkg/dashboard"
)

const (
	broadcastQueueSize            = 20000
	clientSendChannelSize         = 1000
	webSocketWriteTimeout         = time.Duration(3) * time.Second
	maxWebsocketMessageSize int64 = 400 + maxDashboardAuthUsernameSize + 10 // 10 buffer due to variable JWT lengths
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "Dashboard",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Params:    params,
			Provide:   provide,
			Configure: configure,
			Run:       run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies
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
			CoreComponent.LogErrorfAndExit("%s cannot be empty", CoreComponent.App().Config().GetParameterPath(&(ParamsDashboard.Auth.Username)))
		}
		if len(username) > maxDashboardAuthUsernameSize {
			CoreComponent.LogErrorfAndExit("%s has a max length of %d", CoreComponent.App().Config().GetParameterPath(&(ParamsDashboard.Auth.Username)), maxDashboardAuthUsernameSize)
		}

		upgrader := &websocket.Upgrader{
			HandshakeTimeout: webSocketWriteTimeout,
			CheckOrigin:      func(r *http.Request) bool { return true }, // allow any origin for websocket connections
			// Disable compression due to incompatibilities with latest Safari browsers:
			// https://github.com/tilt-dev/tilt/issues/4746
			// https://github.com/gorilla/websocket/issues/731
			EnableCompression: false,
		}

		hub := websockethub.NewHub(CoreComponent.Logger(), upgrader, broadcastQueueSize, clientSendChannelSize, maxWebsocketMessageSize)

		CoreComponent.LogInfo("Setting up dashboard ...")

		return dashboard.New(
			CoreComponent.Logger(),
			CoreComponent.Daemon(),
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
