package nodebridge

import (
	"context"
	"sync"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	inx "github.com/iotaledger/inx/go"
)

type NodeBridge struct {
	// the logger used to log events.
	*logger.WrappedLogger

	client         inx.INXClient
	NodeConfig     *inx.NodeConfiguration
	tangleListener *TangleListener
	Events         *Events

	isSyncedMutex      sync.RWMutex
	latestMilestone    *inx.Milestone
	confirmedMilestone *inx.Milestone
	pruningIndex       uint32

	tipsMetricMutex  sync.RWMutex
	nonLazyPoolSize  uint32
	semiLazyPoolSize uint32
}

type Events struct {
	BlockSolid                *events.Event
	ConfirmedMilestoneChanged *events.Event
}

func INXBlockMetadataCaller(handler interface{}, params ...interface{}) {
	handler.(func(metadata *inx.BlockMetadata))(params[0].(*inx.BlockMetadata))
}

func INXMilestoneCaller(handler interface{}, params ...interface{}) {
	handler.(func(metadata *inx.Milestone))(params[0].(*inx.Milestone))
}

func NewNodeBridge(ctx context.Context, client inx.INXClient, log *logger.Logger) (*NodeBridge, error) {
	log.Info("Connecting to node and reading protocol parameters...")

	retryBackoff := func(_ uint) time.Duration {
		return 1 * time.Second
	}

	nodeConfig, err := client.ReadNodeConfiguration(ctx, &inx.NoParams{}, grpc_retry.WithMax(5), grpc_retry.WithBackoff(retryBackoff))
	if err != nil {
		return nil, err
	}

	nodeStatus, err := client.ReadNodeStatus(ctx, &inx.NoParams{})
	if err != nil {
		return nil, err
	}

	return &NodeBridge{
		WrappedLogger:  logger.NewWrappedLogger(log),
		client:         client,
		NodeConfig:     nodeConfig,
		tangleListener: newTangleListener(),
		Events: &Events{
			BlockSolid:                events.NewEvent(INXBlockMetadataCaller),
			ConfirmedMilestoneChanged: events.NewEvent(INXMilestoneCaller),
		},
		latestMilestone:    nodeStatus.GetLatestMilestone(),
		confirmedMilestone: nodeStatus.GetConfirmedMilestone(),
		pruningIndex:       nodeStatus.GetTanglePruningIndex(),
	}, nil
}

func (n *NodeBridge) Run(ctx context.Context) {
	<-ctx.Done()
}

func (n *NodeBridge) ReadNodeStatus(ctx context.Context) (*inx.NodeStatus, error) {

	nodeStatus, err := n.client.ReadNodeStatus(ctx, &inx.NoParams{})
	if err != nil {
		return nil, err
	}

	return nodeStatus, nil
}
