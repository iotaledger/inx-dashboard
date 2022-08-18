package dashboard

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/core/timeutil"
	"github.com/iotaledger/inx-dashboard/pkg/daemon"
)

func (d *Dashboard) currentDatabaseSize() *DatabaseSizesMetric {
	newMetric, err := d.getDatabaseSizeMetric()
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

	// Gather first metric so we have a starting point
	d.currentDatabaseSize()

	if err := d.daemon.BackgroundWorker("Dashboard[DBSize]", func(ctx context.Context) {
		ticker := timeutil.NewTicker(func() {
			dbSizeMetric := d.currentDatabaseSize()
			if dbSizeMetric == nil {
				return
			}

			d.hub.BroadcastMsg(&Msg{Type: MsgTypeDatabaseSizeMetric, Data: []*DatabaseSizesMetric{dbSizeMetric}})
		}, 1*time.Minute, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}
