package dashboard

import (
	"context"
	"encoding/json"
	"time"

	"github.com/iotaledger/hive.go/timeutil"
	"github.com/iotaledger/inx-dashboard/pkg/daemon"
)

func (s *DatabaseSizeMetric) MarshalJSON() ([]byte, error) {
	size := struct {
		Tangle int64 `json:"tangle"`
		UTXO   int64 `json:"utxo"`
		Total  int64 `json:"total"`
		Time   int64 `json:"ts"`
	}{
		Tangle: s.Tangle,
		UTXO:   s.UTXO,
		Total:  s.Total,
		Time:   s.Time.Unix(),
	}

	return json.Marshal(size)
}

func (d *Dashboard) currentDatabaseSize() *DatabaseSizeMetric {
	/*
		tangleDbSize, err := deps.TangleDatabase.Size()
		if err != nil {
			d.LogWarnf("error in tangle database size calculation: %s", err)
			return nil
		}

		utxoDbSize, err := deps.UTXODatabase.Size()
		if err != nil {
			d.LogWarnf("error in utxo database size calculation: %s", err)
			return nil
		}

		newValue := &DatabaseSizeMetric{
			Tangle: tangleDbSize,
			UTXO:   utxoDbSize,
			Total:  tangleDbSize + utxoDbSize,
			Time:   time.Now(),
		}
	*/

	newMetric := d.getDatabaseSizeMetric()

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
			d.hub.BroadcastMsg(&Msg{Type: MsgTypeDatabaseSizeMetric, Data: []*DatabaseSizeMetric{dbSizeMetric}})
		}, 1*time.Minute, ctx)
		ticker.WaitForGracefulShutdown()
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}
