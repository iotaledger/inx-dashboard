package dashboard

import (
	"context"
	"sync"

	"github.com/iancoleman/orderedmap"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/inx-app/nodebridge"
	"github.com/iotaledger/inx-dashboard/pkg/daemon"
	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
)

const (
	VisualizerIDLength = 10
)

func VertexCaller(handler interface{}, params ...interface{}) {
	handler.(func(*VisualizerVertex))(params[0].(*VisualizerVertex))
}

func ConfirmationCaller(handler interface{}, params ...interface{}) {
	handler.(func(milestoneParents []string, excludedIDs []string))(params[0].([]string), params[1].([]string))
}

type Visualizer struct {
	sync.RWMutex

	vertices *orderedmap.OrderedMap
	capacity int
	Events   *VisualizerEvents
}

type VisualizerEvents struct {
	VertexCreated      *events.Event
	VertexSolidUpdated *events.Event
	VertexTipUpdated   *events.Event
	Confirmation       *events.Event
}

func NewVisualizer(capacity int) *Visualizer {
	return &Visualizer{
		vertices: orderedmap.New(),
		capacity: capacity,
		Events: &VisualizerEvents{
			VertexCreated:      events.NewEvent(VertexCaller),
			VertexSolidUpdated: events.NewEvent(VertexCaller),
			VertexTipUpdated:   events.NewEvent(VertexCaller),
			Confirmation:       events.NewEvent(ConfirmationCaller),
		},
	}
}

func (v *Visualizer) removeOldEntries() {
	// remove old entries
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
		vertex = vert.(*VisualizerVertex)
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

		vertex := v.(*VisualizerVertex)
		if vertex.isCreated {
			if !consumer(vertex) {
				break
			}
		}
	}
}

func (d *Dashboard) runVisualizerFeed() {

	onVisualizerVertexCreated := events.NewClosure(func(vertex *VisualizerVertex) {
		if !d.nodeBridge.IsNodeAlmostSynced() {
			return
		}
		d.hub.BroadcastMsg(&Msg{Type: MsgTypeVisualizerVertex, Data: vertex})
	})

	onVisualizerVertexSolidUpdated := events.NewClosure(func(vertex *VisualizerVertex) {
		if !d.nodeBridge.IsNodeAlmostSynced() {
			return
		}

		d.hub.BroadcastMsg(
			&Msg{
				Type: MsgTypeVisualizerSolidInfo,
				Data: &VisualizerMetaInfo{
					ID: vertex.shortID,
				},
			},
		)
	})

	onVisualizerVertexTipUpdated := events.NewClosure(func(vertex *VisualizerVertex) {
		if !d.nodeBridge.IsNodeAlmostSynced() {
			return
		}

		d.hub.BroadcastMsg(
			&Msg{
				Type: MsgTypeVisualizerTipInfo,
				Data: &VisualizerTipInfo{
					ID:    vertex.shortID,
					IsTip: vertex.IsTip,
				},
			},
		)
	})

	onVisualizerConfirmation := events.NewClosure(func(milestoneParents []string, excludedIDs []string) {
		if !d.nodeBridge.IsNodeAlmostSynced() {
			return
		}

		d.hub.BroadcastMsg(
			&Msg{
				Type: MsgTypeVisualizerConfirmedInfo,
				Data: &VisualizerConfirmationInfo{
					IDs:         milestoneParents,
					ExcludedIDs: excludedIDs,
				},
			},
		)
	})

	onBlockSolid := events.NewClosure(func(metadata *inx.BlockMetadata) {
		d.visualizer.SetIsSolid(metadata.BlockId.Unwrap())
	})

	if err := d.daemon.BackgroundWorker("Dashboard[Visualizer]", func(ctx context.Context) {
		ctxWithCancel, cancel := context.WithCancel(ctx)
		defer cancel()

		d.visualizer.Events.VertexCreated.Hook(onVisualizerVertexCreated)
		defer d.visualizer.Events.VertexCreated.Detach(onVisualizerVertexCreated)
		d.visualizer.Events.VertexSolidUpdated.Hook(onVisualizerVertexSolidUpdated)
		defer d.visualizer.Events.VertexSolidUpdated.Detach(onVisualizerVertexSolidUpdated)
		d.visualizer.Events.VertexTipUpdated.Hook(onVisualizerVertexTipUpdated)
		defer d.visualizer.Events.VertexTipUpdated.Detach(onVisualizerVertexTipUpdated)
		d.visualizer.Events.Confirmation.Hook(onVisualizerConfirmation)
		defer d.visualizer.Events.Confirmation.Detach(onVisualizerConfirmation)
		d.tangleListener.Events.BlockSolid.Hook(onBlockSolid)
		defer d.tangleListener.Events.BlockSolid.Detach(onBlockSolid)

		go func() {
			if err := d.nodeBridge.ListenToBlocks(ctxWithCancel, cancel, func(block *iotago.Block) {
				d.visualizer.AddVertex(block)
			}); err != nil {
				d.LogWarnf("Failed to listen to blocks: %v", err)
			}
		}()

		onConfirmedMilestoneChanged := events.NewClosure(func(ms *nodebridge.Milestone) {
			ctx, cancel := context.WithTimeout(ctxWithCancel, nodeTimeout)
			defer cancel()

			conflictingBlocks := iotago.BlockIDs{}

			if err := d.nodeBridge.MilestoneConeMetadata(ctx, cancel, ms.Milestone.Index, func(metadata *inx.BlockMetadata) {
				blockMeta := blockMetadataFromINXBlockMetadata(metadata)

				d.visualizer.SetIsReferenced(blockMeta.BlockId)

				if blockMeta.IsConflicting {
					d.visualizer.SetIsConflicting(blockMeta.BlockId)
					conflictingBlocks = append(conflictingBlocks, blockMeta.BlockId)
				}
			}); err != nil {
				d.LogWarnf("failed to get milestone cone metadata: %v", err)
			}

			d.visualizer.AddConfirmation(ms.Milestone.Parents, conflictingBlocks)
		})
		d.nodeBridge.Events.ConfirmedMilestoneChanged.Hook(onConfirmedMilestoneChanged)
		defer d.nodeBridge.Events.ConfirmedMilestoneChanged.Detach(onConfirmedMilestoneChanged)

		<-ctx.Done()

		d.LogInfo("Stopping Dashboard[Visualizer] ...")
		d.LogInfo("Stopping Dashboard[Visualizer] ... done")
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}
