package dashboard

import (
	"context"
	"fmt"
	"sync"

	"github.com/iancoleman/orderedmap"
	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/inx-dashboard/pkg/daemon"
	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
)

const (
	VisualizerIDLength = 10
)

type Visualizer struct {
	sync.RWMutex

	// the logger used to log events.
	*logger.WrappedLogger

	nodeBridge *nodebridge.NodeBridge

	vertices *orderedmap.OrderedMap
	capacity int
	running  *atomic.Bool
	active   *atomic.Bool
	//nolint:containedctx // false positive
	ctx                     context.Context
	ctxCancelListenToBlocks context.CancelFunc
	wgListenToBlocks        sync.WaitGroup

	Events *VisualizerEvents
}

type VisualizerEvents struct {
	VertexCreated      *event.Event1[*VisualizerVertex]
	VertexSolidUpdated *event.Event1[*VisualizerVertex]
	VertexTipUpdated   *event.Event1[*VisualizerVertex]
	// params: milestoneParents []string, excludedIDs []string
	Confirmation *event.Event2[[]string, []string]
}

func NewVisualizer(log *logger.Logger, nodeBridge *nodebridge.NodeBridge, capacity int) *Visualizer {
	return &Visualizer{
		WrappedLogger: logger.NewWrappedLogger(log),
		nodeBridge:    nodeBridge,
		vertices:      orderedmap.New(),
		capacity:      capacity,
		running:       atomic.NewBool(false),
		active:        atomic.NewBool(false),
		Events: &VisualizerEvents{
			VertexCreated:      event.New1[*VisualizerVertex](),
			VertexSolidUpdated: event.New1[*VisualizerVertex](),
			VertexTipUpdated:   event.New1[*VisualizerVertex](),
			Confirmation:       event.New2[[]string, []string](),
		},
	}
}

func (v *Visualizer) Run(ctx context.Context) {
	v.Lock()
	defer v.Unlock()

	if v.running.Swap(true) {
		// already running
		return
	}

	v.ctx = ctx
}

func (v *Visualizer) UpdateState(active bool) {
	if !v.running.Load() {
		// do not update the state until the visualizer is running
		return
	}

	v.Lock()
	defer v.Unlock()

	if oldActive := v.active.Swap(active); oldActive == active {
		// state didn't change
		return
	}

	// state changed => subscribe or unsubscribe from INX streams
	if active {
		v.startListenToBlocks()

		return
	}

	// visualizer was deactivated
	// stop listening to blocks
	v.stopListenToBlocks()
}

func (v *Visualizer) stopListenToBlocks() {
	// stop listening to blocks
	if v.ctxCancelListenToBlocks != nil {
		v.ctxCancelListenToBlocks()

		// wait until ListenToBlocks stopped
		v.wgListenToBlocks.Wait()
	}

	// clear the visualizer
	v.clear()
}

func (v *Visualizer) startListenToBlocks() {
	v.stopListenToBlocks()

	// visualizer was activated
	// => start listening to blocks
	v.wgListenToBlocks = sync.WaitGroup{}
	v.wgListenToBlocks.Add(1)

	ctx, cancel := context.WithCancel(v.ctx)
	v.ctxCancelListenToBlocks = cancel

	go func() {
		if err := v.nodeBridge.ListenToBlocks(ctx, cancel, func(block *iotago.Block) {
			v.AddVertex(block)
		}); err != nil {
			v.LogWarnf("Failed to listen to blocks: %v", err)
		}
		v.wgListenToBlocks.Done()
	}()
}

func (v *Visualizer) clear() {
	v.vertices = orderedmap.New()
}

func (v *Visualizer) removeOldEntries() {
	// remove old entries
	//nolint:ifshort // false positive
	keys := v.vertices.Keys()
	if len(keys) >= v.capacity {
		v.vertices.Delete(keys[0])
	}
}

func newVertex(blockID iotago.BlockID) *VisualizerVertex {
	return &VisualizerVertex{
		ID:      blockID.ToHex(),
		shortID: blockID.ToHex()[:VisualizerIDLength],
	}
}

func (v *Visualizer) getEntry(blockID iotago.BlockID) (*VisualizerVertex, bool) {
	id := string(blockID[:])

	var vertex *VisualizerVertex
	vert, exists := v.vertices.Get(id)
	if exists {
		var ok bool
		vertex, ok = vert.(*VisualizerVertex)
		if !ok {
			panic(fmt.Sprintf("expected *VisualizerVertex, got %T", vert))
		}
	} else {
		vertex = newVertex(blockID)
		v.vertices.Set(id, vertex)
		v.removeOldEntries()
	}

	return vertex, exists
}

func (v *Visualizer) AddVertex(block *iotago.Block) {
	v.Lock()
	defer v.Unlock()

	blockID := block.MustID()

	parentsHex := make([]string, len(block.Parents))
	for i, parent := range block.Parents {
		v.setIsReferencedByOthers(parent)
		parentsHex[i] = parent.ToHex()[:VisualizerIDLength]
	}

	vertex, _ := v.getEntry(blockID)
	vertex.Parents = parentsHex
	vertex.isCreated = true
	vertex.IsTip = !vertex.isReferencedByOthers
	vertex.IsTransaction = block.Payload != nil && block.Payload.PayloadType() == iotago.PayloadTransaction
	vertex.IsMilestone = block.Payload != nil && block.Payload.PayloadType() == iotago.PayloadMilestone

	// always trigger the created event, even if the vertex existed before, it was not send yet
	v.Events.VertexCreated.Trigger(vertex)
}

func (v *Visualizer) SetIsSolid(blockID iotago.BlockID) {
	v.Lock()
	defer v.Unlock()

	vertex, exists := v.getEntry(blockID)
	vertex.IsSolid = true
	if exists {
		// trigger the solid event only if the vertex was already created
		v.Events.VertexSolidUpdated.Trigger(vertex)
	}
}

func (v *Visualizer) SetIsReferenced(blockID iotago.BlockID) {
	v.Lock()
	defer v.Unlock()

	vertex, _ := v.getEntry(blockID)
	vertex.IsReferenced = true
}

func (v *Visualizer) SetIsConflicting(blockID iotago.BlockID) {
	v.Lock()
	defer v.Unlock()

	vertex, _ := v.getEntry(blockID)
	vertex.IsConflicting = true
}

func (v *Visualizer) setIsReferencedByOthers(blockID iotago.BlockID) {
	vertex, exists := v.getEntry(blockID)
	vertex.isReferencedByOthers = true

	//nolint:ifshort // false positive
	isTip := vertex.IsTip
	vertex.IsTip = false

	if exists && isTip {
		// trigger the tip event only if the vertex was already created and was a tip
		v.Events.VertexTipUpdated.Trigger(vertex)
	}
}

func (v *Visualizer) AddConfirmation(parents iotago.BlockIDs, conflictingBlocks iotago.BlockIDs) {
	v.Lock()
	defer v.Unlock()

	parentsHex := make([]string, len(parents))
	for i, parent := range parents {
		parentsHex[i] = parent.ToHex()[:VisualizerIDLength]
	}

	conflictingBlocksHex := make([]string, len(conflictingBlocks))
	for i, block := range conflictingBlocks {
		conflictingBlocksHex[i] = block.ToHex()[:VisualizerIDLength]
	}

	v.Events.Confirmation.Trigger(parentsHex, conflictingBlocksHex)
}

func (v *Visualizer) ForEachCreated(consumer func(vertex *VisualizerVertex) bool, elementsCount ...int) {
	v.RLock()
	defer v.RUnlock()

	keys := v.vertices.Keys()

	if len(elementsCount) > 0 && len(keys) > elementsCount[0] {
		keys = keys[len(keys)-elementsCount[0]:]
	}

	for _, key := range keys {
		v, exists := v.vertices.Get(key)
		if !exists {
			continue
		}

		vertex, ok := v.(*VisualizerVertex)
		if !ok {
			panic(fmt.Sprintf("expected *VisualizerVertex, got %T", v))
		}

		if vertex.isCreated {
			if !consumer(vertex) {
				break
			}
		}
	}
}

func (v *Visualizer) ApplyConfirmedMilestoneChanged(ms *nodebridge.Milestone) {

	if !v.active.Load() {
		// visualizer is not active
		return
	}

	ctx, cancel := context.WithTimeout(v.ctx, nodeTimeout)
	defer cancel()

	conflictingBlocks := iotago.BlockIDs{}

	if err := v.nodeBridge.MilestoneConeMetadata(ctx, cancel, ms.Milestone.Index, func(metadata *inx.BlockMetadata) {
		blockMeta := blockMetadataFromINXBlockMetadata(metadata)

		v.SetIsReferenced(blockMeta.BlockID)

		if blockMeta.IsConflicting {
			v.SetIsConflicting(blockMeta.BlockID)
			conflictingBlocks = append(conflictingBlocks, blockMeta.BlockID)
		}
	}); err != nil {
		v.LogWarnf("failed to get milestone cone metadata: %v", err)
	}

	v.AddConfirmation(ms.Milestone.Parents, conflictingBlocks)
}

func (d *Dashboard) runVisualizerFeed() {

	if err := d.daemon.BackgroundWorker("Dashboard[Visualizer]", func(ctx context.Context) {

		onVisualizerVertexCreated := func(vertex *VisualizerVertex) {
			if !d.nodeBridge.IsNodeAlmostSynced() {
				return
			}

			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg,
				&Msg{
					Type: MsgTypeVisualizerVertex,
					Data: vertex,
				},
			)
		}

		onVisualizerVertexSolidUpdated := func(vertex *VisualizerVertex) {
			if !d.nodeBridge.IsNodeAlmostSynced() {
				return
			}

			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg,
				&Msg{
					Type: MsgTypeVisualizerSolidInfo,
					Data: &VisualizerMetaInfo{
						ID: vertex.shortID,
					},
				},
			)
		}

		onVisualizerVertexTipUpdated := func(vertex *VisualizerVertex) {
			if !d.nodeBridge.IsNodeAlmostSynced() {
				return
			}

			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg,
				&Msg{
					Type: MsgTypeVisualizerTipInfo,
					Data: &VisualizerTipInfo{
						ID:    vertex.shortID,
						IsTip: vertex.IsTip,
					},
				},
			)
		}

		onVisualizerConfirmation := func(milestoneParents []string, excludedIDs []string) {
			if !d.nodeBridge.IsNodeAlmostSynced() {
				return
			}

			ctxMsg, ctxMsgCancel := context.WithTimeout(ctx, d.websocketWriteTimeout)
			defer ctxMsgCancel()

			_ = d.hub.BroadcastMsg(ctxMsg,
				&Msg{
					Type: MsgTypeVisualizerConfirmedInfo,
					Data: &VisualizerConfirmationInfo{
						IDs:         milestoneParents,
						ExcludedIDs: excludedIDs,
					},
				},
			)
		}

		onBlockSolid := func(metadata *inx.BlockMetadata) {
			d.visualizer.SetIsSolid(metadata.BlockId.Unwrap())
		}

		onConfirmedMilestoneChanged := d.visualizer.ApplyConfirmedMilestoneChanged

		// register events
		unhook := lo.Batch(
			d.visualizer.Events.VertexCreated.Hook(onVisualizerVertexCreated).Unhook,
			d.visualizer.Events.VertexSolidUpdated.Hook(onVisualizerVertexSolidUpdated).Unhook,
			d.visualizer.Events.VertexTipUpdated.Hook(onVisualizerVertexTipUpdated).Unhook,
			d.visualizer.Events.Confirmation.Hook(onVisualizerConfirmation).Unhook,
			d.tangleListener.Events.BlockSolid.Hook(onBlockSolid).Unhook,
			d.nodeBridge.Events.ConfirmedMilestoneChanged.Hook(onConfirmedMilestoneChanged).Unhook,
		)

		d.visualizer.Run(ctx)

		<-ctx.Done()
		unhook()

		d.LogInfo("Stopping Dashboard[Visualizer] ...")
		d.LogInfo("Stopping Dashboard[Visualizer] ... done")
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}
