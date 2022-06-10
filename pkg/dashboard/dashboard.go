package dashboard

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/basicauth"
	hivedaemon "github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/websockethub"
	"github.com/iotaledger/inx-dashboard/pkg/daemon"
	"github.com/iotaledger/inx-dashboard/pkg/jwt"
)

type (
	getIsNodeAlmostSyncedFunc   func() bool
	getPublicNodeStatusFunc     func() *PublicNodeStatus
	getNodeStatusFunc           func() *NodeStatus
	getPeerInfosFunc            func() []*PeerInfo
	getSyncStatusFunc           func() *SyncStatus
	getDatabaseSizeMetricFunc   func() *DatabaseSizeMetric
	getLatestMilestoneIndexFunc func() uint32
	getMilestoneIDHexFunc       func(index uint32) (string, error)
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
	hub                *websockethub.Hub
	basicAuth          *basicauth.BasicAuth
	jwtAuth            *jwt.JWTAuth

	getIsNodeAlmostSynced   getIsNodeAlmostSyncedFunc
	getPublicNodeStatus     getPublicNodeStatusFunc
	getNodeStatus           getNodeStatusFunc
	getPeerInfos            getPeerInfosFunc
	getSyncStatus           getSyncStatusFunc
	getDatabaseSizeMetric   getDatabaseSizeMetricFunc
	getLatestMilestoneIndex getLatestMilestoneIndexFunc
	getMilestoneIDHex       getMilestoneIDHexFunc
	developerMode           bool

	cachedDatabaseSizeMetrics []*DatabaseSizeMetric
	cachedMilestoneMetrics    []*ConfirmedMilestoneMetric
}

func New(
	daemon hivedaemon.Daemon,
	bindAddress string,
	authUserName string,
	authPasswordHash string,
	authPasswordSalt string,
	authSessionTimeout time.Duration,
	identityFilePath string,
	identityPrivateKey string,
	hub *websockethub.Hub,
	getIsNodeAlmostSynced getIsNodeAlmostSyncedFunc,
	getPublicNodeStatus getPublicNodeStatusFunc,
	getNodeStatus getNodeStatusFunc,
	getPeerInfos getPeerInfosFunc,
	getSyncStatus getSyncStatusFunc,
	getDatabaseSizeMetric getDatabaseSizeMetricFunc,
	getLatestMilestoneIndex getLatestMilestoneIndexFunc,
	getMilestoneIDHex getMilestoneIDHexFunc,
	developerMode bool,
	log *logger.Logger) *Dashboard {

	return &Dashboard{
		WrappedLogger:           logger.NewWrappedLogger(log),
		bindAddress:             bindAddress,
		authUserName:            authUserName,
		authPasswordHash:        authPasswordHash,
		authPasswordSalt:        authPasswordSalt,
		authSessionTimeout:      authSessionTimeout,
		identityFilePath:        identityFilePath,
		identityPrivateKey:      identityPrivateKey,
		hub:                     hub,
		getIsNodeAlmostSynced:   getIsNodeAlmostSynced,
		getPublicNodeStatus:     getPublicNodeStatus,
		getNodeStatus:           getNodeStatus,
		getPeerInfos:            getPeerInfos,
		getSyncStatus:           getSyncStatus,
		getDatabaseSizeMetric:   getDatabaseSizeMetric,
		getLatestMilestoneIndex: getLatestMilestoneIndex,
		getMilestoneIDHex:       getMilestoneIDHex,
		developerMode:           developerMode,
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

}

func newEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	return e
}

func (d *Dashboard) Run() {
	e := newEcho()
	d.setupRoutes(e)

	go func() {
		d.LogInfof("You can now access the dashboard using: http://%s", d.bindAddress)

		if err := e.Start(d.bindAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
			d.LogWarnf("Stopped dashboard server due to an error (%s)", err)
		}
	}()
}

func (d *Dashboard) run() {
	/*
		onBPSMetricsUpdated := events.NewClosure(func(bpsMetrics *BPSMetrics) {
			d.hub.BroadcastMsg(&Msg{Type: MsgTypeBPSMetric, Data: bpsMetrics})
			d.hub.BroadcastMsg(&Msg{Type: MsgTypePublicNodeStatus, Data: d.getPublicNodeStatus()})
			d.hub.BroadcastMsg(&Msg{Type: MsgTypeNodeStatus, Data: d.getNodeStatus()})
			d.hub.BroadcastMsg(&Msg{Type: MsgTypePeerMetric, Data: d.getPeerInfos()})
		})

		onConfirmedMilestoneIndexChanged := events.NewClosure(func(_ uint32) {
			d.hub.BroadcastMsg(&Msg{Type: MsgTypeSyncStatus, Data: d.getSyncStatus()})
		})

		onLatestMilestoneIndexChanged := events.NewClosure(func(_ uint32) {
			d.hub.BroadcastMsg(&Msg{Type: MsgTypeSyncStatus, Data: d.getSyncStatus()})
		})

		onNewConfirmedMilestoneMetric := events.NewClosure(func(metric *ConfirmedMilestoneMetric) {
			d.cachedMilestoneMetrics = append(d.cachedMilestoneMetrics, metric)
			if len(d.cachedMilestoneMetrics) > 20 {
				d.cachedMilestoneMetrics = d.cachedMilestoneMetrics[len(d.cachedMilestoneMetrics)-20:]
			}
			d.hub.BroadcastMsg(&Msg{Type: MsgTypeConfirmedMsMetrics, Data: []*ConfirmedMilestoneMetric{metric}})
		})
	*/
	if err := d.daemon.BackgroundWorker("Dashboard[WSSend]", func(ctx context.Context) {
		go d.hub.Run(ctx)
		/*
			deps.Tangle.Events.BPSMetricsUpdated.Attach(onBPSMetricsUpdated)
			deps.Tangle.Events.ConfirmedMilestoneIndexChanged.Attach(onConfirmedMilestoneIndexChanged)
			deps.Tangle.Events.LatestMilestoneIndexChanged.Attach(onLatestMilestoneIndexChanged)
			deps.Tangle.Events.NewConfirmedMilestoneMetric.Attach(onNewConfirmedMilestoneMetric)
		*/
		<-ctx.Done()
		/*
			d.LogInfo("Stopping Dashboard[WSSend] ...")
			deps.Tangle.Events.BPSMetricsUpdated.Detach(onBPSMetricsUpdated)
			deps.Tangle.Events.ConfirmedMilestoneIndexChanged.Detach(onConfirmedMilestoneIndexChanged)
			deps.Tangle.Events.LatestMilestoneIndexChanged.Detach(onLatestMilestoneIndexChanged)
			deps.Tangle.Events.NewConfirmedMilestoneMetric.Detach(onNewConfirmedMilestoneMetric)
		*/
		d.LogInfo("Stopping Dashboard[WSSend] ... done")
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}

	// run the milestone live feed
	d.runMilestoneLiveFeed()
	// run the visualizer feed
	d.runVisualizerFeed()
	// run the database size collector
	d.runDatabaseSizeCollector()
}
