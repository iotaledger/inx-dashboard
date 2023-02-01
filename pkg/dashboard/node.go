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

func (d *Dashboard) getNodeInfo(ctx context.Context) (*nodeclient.InfoResponse, error) {
	ctxNode, ctxNodecancel := context.WithTimeout(ctx, nodeTimeout)
	defer ctxNodecancel()

	return d.nodeClient.Info(ctxNode)
}

func (d *Dashboard) getNodeInfoExtended(ctx context.Context) (*NodeInfoExtended, error) {
	ctxNode, ctxNodecancel := context.WithTimeout(ctx, nodeTimeout)
	defer ctxNodecancel()

	return d.metricsClient.NodeInfoExtended(ctxNode)
}

func (d *Dashboard) getPeerInfos(ctx context.Context) ([]*nodeclient.PeerResponse, error) {
	ctxNode, ctxNodecancel := context.WithTimeout(ctx, nodeTimeout)
	defer ctxNodecancel()

	return d.nodeClient.Peers(ctxNode)
}

func (d *Dashboard) getSyncStatus() *SyncStatus {
	return &SyncStatus{
		CMI: d.nodeBridge.ConfirmedMilestoneIndex(),
		LMI: d.nodeBridge.LatestMilestoneIndex(),
	}
}

func (d *Dashboard) getGossipMetrics(ctx context.Context) (*GossipMetrics, error) {
	ctxNode, ctxNodecancel := context.WithTimeout(ctx, nodeTimeout)
	defer ctxNodecancel()

	return d.metricsClient.GossipMetrics(ctxNode)
}

func (d *Dashboard) getDatabaseSizeMetric(ctx context.Context) (*DatabaseSizesMetric, error) {
	ctxNode, ctxNodecancel := context.WithTimeout(ctx, nodeTimeout)
	defer ctxNodecancel()

	return d.metricsClient.DatabaseSizes(ctxNode)
}

func (d *Dashboard) getLatestMilestoneIndex() uint32 {
	return d.nodeBridge.LatestMilestoneIndex()
}

func (d *Dashboard) getMilestoneIDHex(ctx context.Context, index uint32) (string, error) {

	ctxNode, ctxNodecancel := context.WithTimeout(ctx, nodeTimeout)
	defer ctxNodecancel()

	milestone, err := d.nodeBridge.Milestone(ctxNode, index)
	if err != nil {
		return "", err
	}

	return milestone.MilestoneID.ToHex(), nil
}
