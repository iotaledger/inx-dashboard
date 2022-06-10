package dashboard

import (
	"context"

	"github.com/iotaledger/inx-dashboard/pkg/daemon"
)

func (d *Dashboard) runMilestoneLiveFeed() {
	/*
		onLatestMilestoneIndexChanged := events.NewClosure(func(msIndex uint32) {
			if milestoneIDHex, err := d.getMilestoneIDHex(msIndex); err == nil {
				d.hub.BroadcastMsg(&Msg{Type: MsgTypeMilestone, Data: &LivefeedMilestone{MilestoneID: milestoneIDHex, Index: msIndex}})
			}
		})
	*/

	if err := d.daemon.BackgroundWorker("Dashboard[TxUpdater]", func(ctx context.Context) {
		/*
			deps.Tangle.Events.LatestMilestoneIndexChanged.Attach(onLatestMilestoneIndexChanged)
			defer deps.Tangle.Events.LatestMilestoneIndexChanged.Detach(onLatestMilestoneIndexChanged)
		*/
		<-ctx.Done()
		d.LogInfo("Stopping Dashboard[TxUpdater] ...")
		d.LogInfo("Stopping Dashboard[TxUpdater] ... done")
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}
