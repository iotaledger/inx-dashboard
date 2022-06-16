package dashboard

import (
	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
)

type BlockMetadata struct {
	BlockId        iotago.BlockID
	Parents        iotago.BlockIDs
	IsSolid        bool
	IsReferenced   bool
	IsConflicting  bool
	ShouldPromote  bool
	ShouldReattach bool
}

func blockMetadataFromINXBlockMetadata(metadata *inx.BlockMetadata) *BlockMetadata {
	return &BlockMetadata{
		BlockId:        metadata.UnwrapBlockID(),
		Parents:        metadata.UnwrapParents(),
		IsSolid:        metadata.GetSolid(),
		IsReferenced:   metadata.GetReferencedByMilestoneIndex() != 0,
		IsConflicting:  metadata.GetConflictReason() != 0,
		ShouldPromote:  metadata.GetShouldPromote(),
		ShouldReattach: metadata.GetShouldReattach(),
	}
}
