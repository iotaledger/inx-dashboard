package dashboard

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/websockethub"
	"github.com/iotaledger/inx-dashboard/pkg/dashboard"
	"github.com/iotaledger/inx-dashboard/pkg/nodebridge"
)

const (
	broadcastQueueSize    = 20000
	clientSendChannelSize = 1000
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

const (
	webSocketWriteTimeout         = time.Duration(3) * time.Second
	maxWebsocketMessageSize int64 = 400 + maxDashboardAuthUsernameSize + 10 // 10 buffer due to variable JWT lengths

)

var (
	CoreComponent *app.CoreComponent
	deps          dependencies

	nodeStartAt = time.Now()
)

type dependencies struct {
	dig.In
	Dashboard  *dashboard.Dashboard
	NodeBridge *nodebridge.NodeBridge
}

func provide(c *dig.Container) error {

	type dashboardDeps struct {
		dig.In
		NodeBridge *nodebridge.NodeBridge
	}

	if err := c.Provide(func(deps dashboardDeps) *dashboard.Dashboard {

		username := ParamsDashboard.Auth.Username
		if len(username) == 0 {
			CoreComponent.LogPanicf("%s cannot be empty", CoreComponent.App.Config().GetParameterPath(&(ParamsDashboard.Auth.Username)))
		}
		if len(username) > maxDashboardAuthUsernameSize {
			CoreComponent.LogPanicf("%s has a max length of %d", CoreComponent.App.Config().GetParameterPath(&(ParamsDashboard.Auth.Username)), maxDashboardAuthUsernameSize)
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

		CoreComponent.LogInfo("Setting up dashboard...")
		return dashboard.New(
			CoreComponent.Daemon(),
			ParamsDashboard.BindAddress,
			ParamsDashboard.Auth.Username,
			ParamsDashboard.Auth.PasswordHash,
			ParamsDashboard.Auth.PasswordSalt,
			ParamsDashboard.Auth.SessionTimeout,
			ParamsDashboard.Auth.IdentityFilePath,
			ParamsDashboard.Auth.IdentityPrivateKey,
			hub,
			getIsNodeAlmostSynced,
			getPublicNodeStatus,
			getNodeStatus,
			getPeerInfos,
			getSyncStatus,
			getDatabaseSizeMetric,
			getLatestMilestoneIndex,
			getMilestoneIDHex,
			ParamsDashboard.DevMode,
			CoreComponent.Logger(),
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

func getIsNodeAlmostSynced() bool {
	return true
}

func getPublicNodeStatus() *dashboard.PublicNodeStatus {
	status := &dashboard.PublicNodeStatus{
		SnapshotIndex: 0,
		PruningIndex:  0,
		IsHealthy:     true,
		IsSynced:      true,
	}

	return status
}

func getNodeStatus() *dashboard.NodeStatus {
	status := &dashboard.NodeStatus{
		Version:       "1",
		LatestVersion: "1",
		Uptime:        0,
		NodeAlias:     ParamsDashboard.Alias,
		NodeID:        "NODE ID",
		MemoryUsage:   0,
	}

	return status
}

func getPeerInfos() []*dashboard.PeerInfo {
	return []*dashboard.PeerInfo{}
}

func getSyncStatus() *dashboard.SyncStatus {
	return &dashboard.SyncStatus{
		CMI: 0,
		LMI: 0,
	}
}

func getDatabaseSizeMetric() *dashboard.DatabaseSizeMetric {
	return &dashboard.DatabaseSizeMetric{
		Tangle: 0,
		UTXO:   0,
		Total:  0,
		Time:   time.Now(),
	}
}

func getLatestMilestoneIndex() uint32 {
	return 0
}

func getMilestoneIDHex(index uint32) (string, error) {
	return "lel", nil
}
