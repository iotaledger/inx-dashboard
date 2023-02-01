package dashboard

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/timeutil"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/inx-dashboard/pkg/daemon"
)

func (d *Dashboard) runNodeInfoFeed() {
	if err := d.daemon.BackgroundWorker("NodeInfo Feed", func(ctx context.Context) {
		ticker := timeutil.NewTicker(func() {
			// skip if no client is connected
			if d.hub.Clients() == 0 {
				return
			}

			nodeInfo, err := d.getNodeInfo(ctx)
			if err != nil {
				d.LogWarnf("failed to get node info: %s", err)

				return
			}

			publicNodeStatus := getPublicNodeStatusByNodeInfo(nodeInfo, d.nodeBridge.IsNodeAlmostSynced())

			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg, &Msg{Type: MsgTypePublicNodeStatus, Data: publicNodeStatus})
			_ = d.hub.BroadcastMsg(ctxMsg, &Msg{Type: MsgTypeConfirmedMsMetrics, Data: nodeInfo.Metrics})
		}, 1*time.Second, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}

func (d *Dashboard) runNodeInfoExtendedFeed() {
	if err := d.daemon.BackgroundWorker("NodeInfoExtended Feed", func(ctx context.Context) {
		ticker := timeutil.NewTicker(func() {
			// skip if no client is connected
			if d.hub.Clients() == 0 {
				return
			}

			data, err := d.getNodeInfoExtended(ctx)
			if err != nil {
				d.LogWarnf("failed to get extended node info: %s", err)

				return
			}

			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg, &Msg{Type: MsgTypeNodeInfoExtended, Data: data})
		}, 1*time.Second, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}

func (d *Dashboard) runSyncStatusFeed() {
	if err := d.daemon.BackgroundWorker("SyncStatus Feed", func(ctx context.Context) {
		onMilestoneIndexChanged := events.NewClosure(func(ms *nodebridge.Milestone) {
			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg, &Msg{Type: MsgTypeSyncStatus, Data: d.getSyncStatus()})
		})

		d.nodeBridge.Events.LatestMilestoneChanged.Hook(onMilestoneIndexChanged)
		defer d.nodeBridge.Events.LatestMilestoneChanged.Detach(onMilestoneIndexChanged)
		d.nodeBridge.Events.ConfirmedMilestoneChanged.Hook(onMilestoneIndexChanged)
		defer d.nodeBridge.Events.ConfirmedMilestoneChanged.Detach(onMilestoneIndexChanged)
		<-ctx.Done()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}

func (d *Dashboard) runGossipMetricsFeed() {
	if err := d.daemon.BackgroundWorker("GossipMetrics Feed", func(ctx context.Context) {
		ticker := timeutil.NewTicker(func() {
			// skip if no client is connected
			if d.hub.Clients() == 0 {
				return
			}

			data, err := d.getGossipMetrics(ctx)
			if err != nil {
				d.LogWarnf("failed to get gossip metrics: %s", err)

				return
			}

			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg, &Msg{Type: MsgTypeGossipMetrics, Data: data})
		}, 1*time.Second, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}

func (d *Dashboard) runMilestoneLiveFeed() {

	if err := d.daemon.BackgroundWorker("Milestones Feed", func(ctx context.Context) {
		onLatestMilestoneIndexChanged := events.NewClosure(func(ms *nodebridge.Milestone) {
			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg,
				&Msg{
					Type: MsgTypeMilestone,
					Data: &Milestone{
						MilestoneID: ms.MilestoneID.ToHex(),
						Index:       ms.Milestone.Index,
					},
				})
		})

		d.nodeBridge.Events.LatestMilestoneChanged.Hook(onLatestMilestoneIndexChanged)
		defer d.nodeBridge.Events.LatestMilestoneChanged.Detach(onLatestMilestoneIndexChanged)
		<-ctx.Done()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}

func (d *Dashboard) runPeerMetricsFeed() {

	if err := d.daemon.BackgroundWorker("PeerMetrics Feed", func(ctx context.Context) {
		ticker := timeutil.NewTicker(func() {
			// skip if no client is connected
			if d.hub.Clients() == 0 {
				return
			}

			data, err := d.getPeerInfos(ctx)
			if err != nil {
				return
			}

			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg, &Msg{Type: MsgTypePeerMetric, Data: data})
		}, 1*time.Second, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}
