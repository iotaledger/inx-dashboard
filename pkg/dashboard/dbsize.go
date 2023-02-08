package dashboard

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/core/timeutil"
	"github.com/iotaledger/inx-dashboard/pkg/daemon"
)

func (d *Dashboard) currentDatabaseSize(ctx context.Context) *DatabaseSizesMetric {
	newMetric, err := d.getDatabaseSizeMetric(ctx)
	if err != nil {
		d.LogWarnf("error in database size calculation: %s", err)

		return nil
	}

	d.cachedDatabaseSizeMetrics = append(d.cachedDatabaseSizeMetrics, newMetric)
	if len(d.cachedDatabaseSizeMetrics) > 600 {
		d.cachedDatabaseSizeMetrics = d.cachedDatabaseSizeMetrics[len(d.cachedDatabaseSizeMetrics)-600:]
	}

	return newMetric
}

func (d *Dashboard) runDatabaseSizeCollector() {
	if err := d.daemon.BackgroundWorker("Dashboard[DBSize]", func(ctx context.Context) {
		// Gather first metric so we have a starting point
		d.currentDatabaseSize(ctx)

		ticker := timeutil.NewTicker(func() {
			dbSizeMetric := d.currentDatabaseSize(ctx)
			if dbSizeMetric == nil {
				return
			}

			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg, &Msg{Type: MsgTypeDatabaseSizeMetric, Data: []*DatabaseSizesMetric{dbSizeMetric}})
		}, 1*time.Minute, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}
