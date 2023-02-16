package dashboard

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/core/basicauth"
	hivedaemon "github.com/iotaledger/hive.go/core/daemon"
	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/generics/options"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/subscriptionmanager"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/inx-app/pkg/httpserver"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/inx-dashboard/pkg/daemon"
	"github.com/iotaledger/inx-dashboard/pkg/jwt"
	"github.com/iotaledger/iota.go/v3/nodeclient"
)

const (
	VisualizerInitValuesCount = 3000
	VisualizerCapacity        = 3000
)

type Dashboard struct {
	// the logger used to log events.
	*logger.WrappedLogger

	// used to access the global daemon.
	daemon     hivedaemon.Daemon
	nodeBridge *nodebridge.NodeBridge
	hub        *websockethub.Hub

	bindAddress              string
	developerMode            bool
	developerModeURL         string
	authUsername             string
	authPasswordHash         string
	authPasswordSalt         string
	authSessionTimeout       time.Duration
	authIdentityFilePath     string
	authIdentityPrivateKey   string
	authRateLimitEnabled     bool
	authRateLimitPeriod      time.Duration
	authRateLimitMaxRequests int
	authRateLimitMaxBurst    int
	websocketWriteTimeout    time.Duration
	debugLogRequests         bool

	basicAuth      *basicauth.BasicAuth
	jwtAuth        *jwt.Auth
	nodeClient     *nodeclient.Client
	tangleListener *nodebridge.TangleListener
	metricsClient  *MetricsClient

	visualizer          *Visualizer
	subscriptionManager *subscriptionmanager.SubscriptionManager[websockethub.ClientID, WebSocketMsgType]

	cachedDatabaseSizeMetrics []*DatabaseSizesMetric
}

func WithBindAddress(bindAddress string) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.bindAddress = bindAddress
	}
}

func WithDeveloperMode(developerMode bool) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.developerMode = developerMode
	}
}

func WithDeveloperModeURL(developerModeURL string) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.developerModeURL = developerModeURL
	}
}

func WithAuthUsername(authUsername string) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.authUsername = authUsername
	}
}

func WithAuthPasswordHash(authPasswordHash string) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.authPasswordHash = authPasswordHash
	}
}

func WithAuthPasswordSalt(authPasswordSalt string) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.authPasswordSalt = authPasswordSalt
	}
}

func WithAuthSessionTimeout(authSessionTimeout time.Duration) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.authSessionTimeout = authSessionTimeout
	}
}

func WithAuthIdentityFilePath(authIdentityFilePath string) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.authIdentityFilePath = authIdentityFilePath
	}
}

func WithAuthIdentityPrivateKey(authIdentityPrivateKey string) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.authIdentityPrivateKey = authIdentityPrivateKey
	}
}

func WithAuthRateLimitEnabled(authRateLimitEnabled bool) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.authRateLimitEnabled = authRateLimitEnabled
	}
}

func WithAuthRateLimitPeriod(authRateLimitPeriod time.Duration) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.authRateLimitPeriod = authRateLimitPeriod
	}
}

func WithAuthRateLimitMaxRequests(authRateLimitMaxRequests int) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.authRateLimitMaxRequests = authRateLimitMaxRequests
	}
}

func WithAuthRateLimitMaxBurst(authRateLimitMaxBurst int) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.authRateLimitMaxBurst = authRateLimitMaxBurst
	}
}

func WithWebsocketWriteTimeout(writeTimeout time.Duration) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.websocketWriteTimeout = writeTimeout
	}
}

func WithDebugLogRequests(debugLogRequests bool) options.Option[Dashboard] {
	return func(d *Dashboard) {
		d.debugLogRequests = debugLogRequests
	}
}

func New(
	log *logger.Logger,
	daemon hivedaemon.Daemon,
	nodeBridge *nodebridge.NodeBridge,
	hub *websockethub.Hub,
	opts ...options.Option[Dashboard]) *Dashboard {

	d := options.Apply(&Dashboard{
		WrappedLogger: logger.NewWrappedLogger(log),
		daemon:        daemon,
		nodeBridge:    nodeBridge,
		hub:           hub,

		bindAddress:              "localhost:8081",
		developerMode:            false,
		developerModeURL:         "http://127.0.0.1:9090",
		authUsername:             "admin",
		authPasswordHash:         "0000000000000000000000000000000000000000000000000000000000000000",
		authPasswordSalt:         "0000000000000000000000000000000000000000000000000000000000000000",
		authSessionTimeout:       72 * time.Hour,
		authIdentityFilePath:     "identity.key",
		authIdentityPrivateKey:   "",
		authRateLimitEnabled:     true,
		authRateLimitPeriod:      1 * time.Minute,
		authRateLimitMaxRequests: 20,
		authRateLimitMaxBurst:    30,
		websocketWriteTimeout:    5 * time.Second,
		debugLogRequests:         false,

		visualizer:          NewVisualizer(log, nodeBridge, VisualizerCapacity),
		subscriptionManager: subscriptionmanager.New[websockethub.ClientID, WebSocketMsgType](),
	}, opts)

	// events
	d.subscriptionManager.Events().TopicAdded.Hook(event.NewClosure(func(event *subscriptionmanager.TopicEvent[WebSocketMsgType]) {
		d.checkVisualizerSubscriptions()
	}))
	d.subscriptionManager.Events().TopicRemoved.Hook(event.NewClosure(func(event *subscriptionmanager.TopicEvent[WebSocketMsgType]) {
		d.checkVisualizerSubscriptions()
	}))

	return d
}

func (d *Dashboard) checkVisualizerSubscriptions() {

	active := false
	for _, topic := range []WebSocketMsgType{
		MsgTypeVisualizerVertex,
		MsgTypeVisualizerSolidInfo,
		MsgTypeVisualizerConfirmedInfo,
		MsgTypeVisualizerMilestoneInfo,
		MsgTypeVisualizerTipInfo} {
		if d.subscriptionManager.TopicHasSubscribers(topic) {
			active = true

			break
		}
	}

	d.visualizer.UpdateState(active)
}

func (d *Dashboard) Init() {

	basicAuth, err := basicauth.NewBasicAuth(
		d.authUsername,
		d.authPasswordHash,
		d.authPasswordSalt)
	if err != nil {
		d.LogErrorfAndExit("basic auth initialization failed: %w", err)
	}
	d.basicAuth = basicAuth

	// make sure nobody copies around the identity file since it contains the private key of the JWT auth
	d.LogInfof(`WARNING: never share your "%s" file as it contains your JWT private key!`, d.authIdentityFilePath)

	// load up the previously generated identity or create a new one
	privKey, newlyCreated, err := jwt.LoadOrCreateIdentityPrivateKey(d.authIdentityFilePath, d.authIdentityPrivateKey)
	if err != nil {
		d.LogErrorAndExit(err)
	}

	if newlyCreated {
		d.LogInfof(`stored new private key for identity under "%s"`, d.authIdentityFilePath)
	} else {
		d.LogInfof(`loaded existing private key for identity from "%s"`, d.authIdentityFilePath)
	}

	pubKey, ok := privKey.Public().(ed25519.PublicKey)
	if !ok {
		panic(fmt.Sprintf("expected ed25519.PublicKey, got %T", privKey.Public()))
	}

	hashedPubKey := blake2b.Sum256(pubKey[:])
	identity := hex.EncodeToString(hashedPubKey[:])

	jwtAuth, err := jwt.NewAuth(
		d.authUsername,
		d.authSessionTimeout,
		identity,
		privKey,
	)
	if err != nil {
		d.LogErrorfAndExit("JWT auth initialization failed: %w", err)
	}
	d.jwtAuth = jwtAuth

	d.nodeClient = d.nodeBridge.INXNodeClient()
	d.tangleListener = nodebridge.NewTangleListener(d.nodeBridge)
	d.metricsClient = NewMetricsClient(d.nodeClient)
}

func (d *Dashboard) Run() {
	e := httpserver.NewEcho(d.Logger(), nil, d.debugLogRequests)
	d.setupRoutes(e)

	if err := d.daemon.BackgroundWorker("Dashboard", func(ctx context.Context) {
		d.LogInfo("Starting Dashboard server ...")

		e.Server.BaseContext = func(l net.Listener) context.Context {
			// set BaseContext to be the same as the worker,
			// so that requests being processed don't hang the shutdown procedure
			return ctx
		}

		go func() {
			d.LogInfof("You can now access the dashboard using: http://%s", d.bindAddress)
			if err := e.Start(d.bindAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
				d.LogErrorfAndExit("Stopped dashboard server due to an error (%s)", err)
			}
		}()

		d.LogInfo("Starting Dashboard server ... done")
		<-ctx.Done()
		d.LogInfo("Stopping Dashboard server ...")

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCtxCancel()

		//nolint:contextcheck // false positive
		if err := e.Shutdown(shutdownCtx); err != nil {
			d.LogWarn(err)
		}

		d.LogInfo("Stopping Dashboard server... done")
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}

	if err := d.daemon.BackgroundWorker("WebsocketHub", func(ctx context.Context) {
		go d.hub.Run(ctx)
		<-ctx.Done()
		d.LogInfo("Stopping WebsocketHub ...")
		d.LogInfo("Stopping WebsocketHub ... done")
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}

	if err := d.daemon.BackgroundWorker("TangleListener", func(ctx context.Context) {
		d.tangleListener.Run(ctx)
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}

	d.runNodeInfoFeed()
	d.runNodeInfoExtendedFeed()
	d.runGossipMetricsFeed()
	d.runSyncStatusFeed()
	d.runPeerMetricsFeed()
	d.runMilestoneLiveFeed()
	d.runVisualizerFeed()
	d.runDatabaseSizeCollector()
}
