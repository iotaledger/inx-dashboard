package dashboard

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/timeutil"
	"github.com/iotaledger/inx-app/nodebridge"
	"github.com/iotaledger/inx-dashboard/pkg/daemon"
)

func (d *Dashboard) runNodeInfoFeed() {
	if err := d.daemon.BackgroundWorker("NodeInfo Feed", func(ctx context.Context) {
		ticker := timeutil.NewTicker(func() {
			nodeInfo, err := d.getNodeInfo()
			if err != nil {
				d.LogWarnf("failed to get node info: %s", err)
				return
			}

			publicNodeStatus := getPublicNodeStatusByNodeInfo(nodeInfo, d.nodeBridge.IsNodeAlmostSynced())
			d.hub.BroadcastMsg(&Msg{Type: MsgTypePublicNodeStatus, Data: publicNodeStatus})
			d.hub.BroadcastMsg(&Msg{Type: MsgTypeConfirmedMsMetrics, Data: nodeInfo.Metrics})
		}, 1*time.Second, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}

func (d *Dashboard) runNodeInfoExtendedFeed() {
	if err := d.daemon.BackgroundWorker("NodeInfoExtended Feed", func(ctx context.Context) {
		ticker := timeutil.NewTicker(func() {
			data, err := d.getNodeInfoExtended()
			if err != nil {
				d.LogWarnf("failed to get extended node info: %s", err)
				return
			}
			d.hub.BroadcastMsg(&Msg{Type: MsgTypeNodeInfoExtended, Data: data})
		}, 1*time.Second, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}

func (d *Dashboard) runSyncStatusFeed() {
	onMilestoneIndexChanged := events.NewClosure(func(ms *nodebridge.Milestone) {
		d.hub.BroadcastMsg(&Msg{Type: MsgTypeSyncStatus, Data: d.getSyncStatus()})
	})

	if err := d.daemon.BackgroundWorker("SyncStatus Feed", func(ctx context.Context) {
		d.nodeBridge.Events.LatestMilestoneChanged.Attach(onMilestoneIndexChanged)
		defer d.nodeBridge.Events.LatestMilestoneChanged.Detach(onMilestoneIndexChanged)
		d.nodeBridge.Events.ConfirmedMilestoneChanged.Attach(onMilestoneIndexChanged)
		defer d.nodeBridge.Events.ConfirmedMilestoneChanged.Detach(onMilestoneIndexChanged)
		<-ctx.Done()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}

func (d *Dashboard) runGossipMetricsFeed() {
	if err := d.daemon.BackgroundWorker("GossipMetrics Feed", func(ctx context.Context) {
		ticker := timeutil.NewTicker(func() {
			data, err := d.getGossipMetrics()
			if err != nil {
				d.LogWarnf("failed to get gossip metrics: %s", err)
				return
			}
			d.hub.BroadcastMsg(&Msg{Type: MsgTypeGossipMetrics, Data: data})
		}, 1*time.Second, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}

func (d *Dashboard) runMilestoneLiveFeed() {
	onLatestMilestoneIndexChanged := events.NewClosure(func(ms *nodebridge.Milestone) {
		d.hub.BroadcastMsg(&Msg{
			Type: MsgTypeMilestone,
			Data: &Milestone{
				MilestoneID: ms.MilestoneID.ToHex(),
				Index:       ms.Milestone.Index,
			},
		})
	})

	if err := d.daemon.BackgroundWorker("Milestones Feed", func(ctx context.Context) {
		d.nodeBridge.Events.LatestMilestoneChanged.Attach(onLatestMilestoneIndexChanged)
		defer d.nodeBridge.Events.LatestMilestoneChanged.Detach(onLatestMilestoneIndexChanged)
		<-ctx.Done()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}

func (d *Dashboard) runPeerMetricsFeed() {

	if err := d.daemon.BackgroundWorker("PeerMetrics Feed", func(ctx context.Context) {
		ticker := timeutil.NewTicker(func() {
			data, err := d.getPeerInfos()
			if err != nil {
				return
			}
			d.hub.BroadcastMsg(&Msg{Type: MsgTypePeerMetric, Data: data})
		}, 1*time.Second, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}
