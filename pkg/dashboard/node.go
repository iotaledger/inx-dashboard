package dashboard

import (
	"context"
	"time"

	"github.com/iotaledger/iota.go/v3/nodeclient"
)

const (
	nodeTimeout = 5 * time.Second
)

func getPublicNodeStatusByNodeInfo(nodeInfo *nodeclient.InfoResponse, isAlmostSynced bool) *PublicNodeStatus {
	return &PublicNodeStatus{
		PruningIndex: nodeInfo.Status.PruningIndex,
		IsHealthy:    nodeInfo.Status.IsHealthy,
		IsSynced:     isAlmostSynced,
	}
}

func (d *Dashboard) getNodeInfo() (*nodeclient.InfoResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), nodeTimeout)
	defer cancel()

	return d.nodeClient.Info(ctx)
}

func (d *Dashboard) getNodeInfoExtended() (*NodeInfoExtended, error) {
	ctx, cancel := context.WithTimeout(context.Background(), nodeTimeout)
	defer cancel()

	return d.metricsClient.NodeInfoExtended(ctx)
}

func (d *Dashboard) getPeerInfos() ([]*nodeclient.PeerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), nodeTimeout)
	defer cancel()

	return d.nodeClient.Peers(ctx)
}

func (d *Dashboard) getSyncStatus() *SyncStatus {
	return &SyncStatus{
		CMI: d.nodeBridge.ConfirmedMilestoneIndex(),
		LMI: d.nodeBridge.LatestMilestoneIndex(),
	}
}

func (d *Dashboard) getGossipMetrics() (*GossipMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), nodeTimeout)
	defer cancel()

	return d.metricsClient.GossipMetrics(ctx)
}

func (d *Dashboard) getDatabaseSizeMetric() (*DatabaseSizesMetric, error) {
	ctx, cancel := context.WithTimeout(context.Background(), nodeTimeout)
	defer cancel()

	return d.metricsClient.DatabaseSizes(ctx)
}

func (d *Dashboard) getLatestMilestoneIndex() uint32 {
	return d.nodeBridge.LatestMilestoneIndex()
}

func (d *Dashboard) getMilestoneIDHex(index uint32) (string, error) {
	milestone, err := d.nodeBridge.Milestone(index)
	if err != nil {
		return "", err
	}

	return milestone.MilestoneID.ToHex(), nil
}
