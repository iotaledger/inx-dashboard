package dashboard

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/basicauth"
	hivedaemon "github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/websockethub"
	"github.com/iotaledger/inx-app/httpserver"
	"github.com/iotaledger/inx-app/nodebridge"
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
	daemon             hivedaemon.Daemon
	bindAddress        string
	authUserName       string
	authPasswordHash   string
	authPasswordSalt   string
	authSessionTimeout time.Duration
	identityFilePath   string
	identityPrivateKey string
	developerMode      bool
	developerModeURL   string
	nodeBridge         *nodebridge.NodeBridge
	tangleListener     *nodebridge.TangleListener
	hub                *websockethub.Hub
	debugLogRequests   bool

	basicAuth     *basicauth.BasicAuth
	jwtAuth       *jwt.JWTAuth
	nodeClient    *nodeclient.Client
	metricsClient *MetricsClient

	visualizer *Visualizer

	cachedDatabaseSizeMetrics []*DatabaseSizesMetric
}

func New(
	log *logger.Logger,
	daemon hivedaemon.Daemon,
	bindAddress string,
	authUserName string,
	authPasswordHash string,
	authPasswordSalt string,
	authSessionTimeout time.Duration,
	identityFilePath string,
	identityPrivateKey string,
	developerMode bool,
	developerModeURL string,
	nodeBridge *nodebridge.NodeBridge,
	hub *websockethub.Hub,
	debugLogRequests bool) *Dashboard {

	return &Dashboard{
		WrappedLogger:      logger.NewWrappedLogger(log),
		daemon:             daemon,
		bindAddress:        bindAddress,
		authUserName:       authUserName,
		authPasswordHash:   authPasswordHash,
		authPasswordSalt:   authPasswordSalt,
		authSessionTimeout: authSessionTimeout,
		identityFilePath:   identityFilePath,
		identityPrivateKey: identityPrivateKey,
		developerMode:      developerMode,
		developerModeURL:   developerModeURL,
		nodeBridge:         nodeBridge,
		hub:                hub,
		debugLogRequests:   debugLogRequests,
		visualizer:         NewVisualizer(VisualizerCapacity),
	}
}

func (d *Dashboard) Init() {

	basicAuth, err := basicauth.NewBasicAuth(
		d.authUserName,
		d.authPasswordHash,
		d.authPasswordSalt)
	if err != nil {
		d.LogPanicf("basic auth initialization failed: %w", err)
	}
	d.basicAuth = basicAuth

	// make sure nobody copies around the identity file since it contains the private key of the JWT auth
	d.LogInfof(`WARNING: never share your "%s" file as it contains your JWT private key!`, d.identityFilePath)

	// load up the previously generated identity or create a new one
	privKey, newlyCreated, err := jwt.LoadOrCreateIdentityPrivateKey(d.identityFilePath, d.identityPrivateKey)
	if err != nil {
		d.LogPanic(err)
	}

	if newlyCreated {
		d.LogInfof(`stored new private key for identity under "%s"`, d.identityFilePath)
	} else {
		d.LogInfof(`loaded existing private key for identity from "%s"`, d.identityFilePath)
	}

	pubKey := privKey.Public().(ed25519.PublicKey)
	hashedPubKey := blake2b.Sum256(pubKey[:])
	identity := hex.EncodeToString(hashedPubKey[:])

	jwtAuth, err := jwt.NewJWTAuth(
		d.authUserName,
		d.authSessionTimeout,
		identity,
		privKey,
	)
	if err != nil {
		d.LogPanicf("JWT auth initialization failed: %w", err)
	}
	d.jwtAuth = jwtAuth

	d.nodeClient = d.nodeBridge.INXNodeClient()
	d.tangleListener = nodebridge.NewTangleListener(d.nodeBridge)
	d.metricsClient = NewMetricsClient(d.nodeClient)
}

func (d *Dashboard) Run() {
	e := httpserver.NewEcho(d.Logger(), nil, d.debugLogRequests)
	d.setupRoutes(e)

	go func() {
		d.LogInfof("You can now access the dashboard using: http://%s", d.bindAddress)

		if err := e.Start(d.bindAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
			d.LogWarnf("Stopped dashboard server due to an error (%s)", err)
		}
	}()

	if err := d.daemon.BackgroundWorker("Dashboard[WSSend]", func(ctx context.Context) {
		go d.hub.Run(ctx)
		<-ctx.Done()
		d.LogInfo("Stopping Dashboard[WSSend] ...")
		d.LogInfo("Stopping Dashboard[WSSend] ... done")
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
