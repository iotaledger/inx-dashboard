package dashboard

import (
	"context"

	"github.com/iotaledger/inx-dashboard/pkg/daemon"
)

const (
	VisualizerIDLength = 7
)

func (d *Dashboard) runVisualizerFeed() {
	/*
		onReceivedNewBlock := events.NewClosure(func(cachedBlock *storage.CachedBlock, _ uint32, _ uint32) {
			cachedBlock.ConsumeBlockAndMetadata(func(block *storage.Block, metadata *storage.BlockMetadata) { // block -1
				if !deps.SyncManager.IsNodeAlmostSynced() {
					return
				}

				parentsHex := make([]string, len(block.Parents()))
				for i, parent := range block.Parents() {
					parentsHex[i] = parent.ToHex()[:VisualizerIDLength]
				}
				d.hub.BroadcastMsg(
					&Msg{
						Type: MsgTypeVertex,
						Data: &Vertex{
							ID:           block.BlockID().ToHex(),
							Parents:      parentsHex,
							IsSolid:      metadata.IsSolid(),
							IsReferenced: metadata.IsReferenced(),
							IsMilestone:  false,
							IsTip:        false,
						},
					},
				)
			})
		})

		onBlockSolid := events.NewClosure(func(cachedBlockMeta *storage.CachedMetadata) {
			cachedBlockMeta.ConsumeMetadata(func(metadata *storage.BlockMetadata) { // meta -1

				if !deps.SyncManager.IsNodeAlmostSynced() {
					return
				}
				d.hub.BroadcastMsg(
					&Msg{
						Type: MsgTypeSolidInfo,
						Data: &MetaInfo{
							ID: metadata.BlockID().ToHex()[:VisualizerIDLength],
						},
					},
				)
			})
		})
	*/
	/*
		onReceivedNewMilestoneBlock := events.NewClosure(func(blockID iotago.BlockID) {
			if !deps.SyncManager.IsNodeAlmostSynced() {
				return
			}

			d.hub.BroadcastMsg(
				&Msg{
					Type: MsgTypeMilestoneInfo,
					Data: &MetaInfo{
						ID: blockID.ToHex()[:VisualizerIDLength],
					},
				},
			)
		})
	*/
	/*
		onMilestoneConfirmed := events.NewClosure(func(confirmation *whiteflag.Confirmation) {
			if !deps.SyncManager.IsNodeAlmostSynced() {
				return
			}

			milestoneParents := make([]string, len(confirmation.MilestoneParents))
			for i, parent := range confirmation.MilestoneParents {
				milestoneParents[i] = parent.ToHex()[:VisualizerIDLength]
			}

			excludedIDs := make([]string, len(confirmation.Mutations.BlocksExcludedWithConflictingTransactions))
			for i, blockID := range confirmation.Mutations.BlocksExcludedWithConflictingTransactions {
				excludedIDs[i] = blockID.BlockID.ToHex()[:VisualizerIDLength]
			}

			d.hub.BroadcastMsg(
				&Msg{
					Type: MsgTypeConfirmedInfo,
					Data: &ConfirmationInfo{
						IDs:         milestoneParents,
						ExcludedIDs: excludedIDs,
					},
				},
			)
		})
	*/
	/*
		// TODO: replace this with code that checks if a block is referenced or not
		onTipAdded := events.NewClosure(func(tip *tipselect.Tip) {
			if !deps.SyncManager.IsNodeAlmostSynced() {
				return
			}

			d.hub.BroadcastMsg(
				&Msg{
					Type: MsgTypeTipInfo,
					Data: &TipInfo{
						ID:    tip.BlockID.ToHex()[:VisualizerIDLength],
						IsTip: true,
					},
				},
			)
		})

		onTipRemoved := events.NewClosure(func(tip *tipselect.Tip) {
			if !deps.SyncManager.IsNodeAlmostSynced() {
				return
			}

			d.hub.BroadcastMsg(
				&Msg{
					Type: MsgTypeTipInfo,
					Data: &TipInfo{
						ID:    tip.BlockID.ToHex()[:VisualizerIDLength],
						IsTip: false,
					},
				},
			)
		})
	*/

	if err := d.daemon.BackgroundWorker("Dashboard[Visualizer]", func(ctx context.Context) {
		/*
			deps.Tangle.Events.ReceivedNewBlock.Attach(onReceivedNewBlock)
			defer deps.Tangle.Events.ReceivedNewBlock.Detach(onReceivedNewBlock)
			deps.Tangle.Events.BlockSolid.Attach(onBlockSolid)
			defer deps.Tangle.Events.BlockSolid.Detach(onBlockSolid)
			deps.Tangle.Events.ReceivedNewMilestoneBlock.Attach(onReceivedNewMilestoneBlock)
			defer deps.Tangle.Events.ReceivedNewMilestoneBlock.Detach(onReceivedNewMilestoneBlock)
			deps.Tangle.Events.MilestoneConfirmed.Attach(onMilestoneConfirmed)
			defer deps.Tangle.Events.MilestoneConfirmed.Detach(onMilestoneConfirmed)

			if deps.TipSelector != nil {
				deps.TipSelector.Events.TipAdded.Attach(onTipAdded)
				defer deps.TipSelector.Events.TipAdded.Detach(onTipAdded)
				deps.TipSelector.Events.TipRemoved.Attach(onTipRemoved)
				defer deps.TipSelector.Events.TipRemoved.Detach(onTipRemoved)
			}
		*/

		<-ctx.Done()

		d.LogInfo("Stopping Dashboard[Visualizer] ...")
		d.LogInfo("Stopping Dashboard[Visualizer] ... done")
	}, daemon.PriorityStopDashboard); err != nil {
		d.LogPanicf("failed to start worker: %s", err)
	}
}
