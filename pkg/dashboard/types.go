package dashboard

import (
	"time"
)

// BPSMetrics represents a websocket message.
type BPSMetrics struct {
	Incoming uint32 `json:"incoming"`
	New      uint32 `json:"new"`
	Outgoing uint32 `json:"outgoing"`
}

// ConfirmationInfo signals confirmation of a milestone block with a list of exluded blocks in the past cone.
type ConfirmationInfo struct {
	IDs         []string `json:"ids"`
	ExcludedIDs []string `json:"excludedIds"`
}

// ConfirmedMilestoneMetric holds information about a confirmed milestone.
type ConfirmedMilestoneMetric struct {
	MilestoneIndex            uint32  `json:"milestoneIndex"`
	BlocksPerSecond           float64 `json:"blocksPerSecond"`
	ReferencedBlocksPerSecond float64 `json:"referencedBlocksPerSecond"`
	ReferencedRate            float64 `json:"referencedRate"`
	TimeSinceLastMilestone    float64 `json:"timeSinceLastMilestone"`
}

// DatabaseSizeMetric represents database size metrics.
type DatabaseSizeMetric struct {
	Tangle int64
	UTXO   int64
	Total  int64
	Time   time.Time
}

// MetaInfo signals that metadata of a given block changed.
type MetaInfo struct {
	ID string `json:"id"`
}

// LivefeedMilestone represents a milestone for the livefeed.
type LivefeedMilestone struct {
	MilestoneID string `json:"milestoneId"`
	Index       uint32 `json:"index"`
}

// NodeStatus represents the node status.
type NodeStatus struct {
	Version       string `json:"version"`
	LatestVersion string `json:"latestVersion"`
	Uptime        int64  `json:"uptime"`
	NodeID        string `json:"nodeId"`
	NodeAlias     string `json:"nodeAlias"`
	MemoryUsage   int64  `json:"memUsage"`
}

// PublicNodeStatus represents the public node status.
type PublicNodeStatus struct {
	SnapshotIndex uint32 `json:"snapshotIndex"`
	PruningIndex  uint32 `json:"pruningIndex"`
	IsHealthy     bool   `json:"isHealthy"`
	IsSynced      bool   `json:"isSynced"`
}

// SyncStatus represents the node sync status.
type SyncStatus struct {
	CMI uint32 `json:"cmi"`
	LMI uint32 `json:"lmi"`
}

// tipinfo holds information about whether a given block is a tip or not.
type TipInfo struct {
	ID    string `json:"id"`
	IsTip bool   `json:"isTip"`
}

// Vertex defines a vertex in a DAG.
type Vertex struct {
	ID           string   `json:"id"`
	Parents      []string `json:"parents"`
	IsSolid      bool     `json:"isSolid"`
	IsReferenced bool     `json:"isReferenced"`
	IsMilestone  bool     `json:"isMilestone"`
	IsTip        bool     `json:"isTip"`
}

// Msg represents a websocket message.
type Msg struct {
	Type byte        `json:"type"`
	Data interface{} `json:"data"`
}

// Heartbeat contains information about a nodes current solid and pruned milestone index
// and its connected and synced neighbors count.
type Heartbeat struct {
	SolidMilestoneIndex  uint32 `json:"solidMilestoneIndex"`
	PrunedMilestoneIndex uint32 `json:"prunedMilestoneIndex"`
	LatestMilestoneIndex uint32 `json:"latestMilestoneIndex"`
	ConnectedNeighbors   int    `json:"connectedNeighbors"`
	SyncedNeighbors      int    `json:"syncedNeighbors"`
}

// PeerGossipMetrics represents a snapshot of the gossip protocol metrics.
type PeerGossipMetrics struct {
	NewBlocks                 uint32 `json:"newBlocks"`
	KnownBlocks               uint32 `json:"knownBlocks"`
	ReceivedBlocks            uint32 `json:"receivedBlocks"`
	ReceivedBlockRequests     uint32 `json:"receivedBlockRequests"`
	ReceivedMilestoneRequests uint32 `json:"receivedMilestoneRequests"`
	ReceivedHeartbeats        uint32 `json:"receivedHeartbeats"`
	SentBlocks                uint32 `json:"sentBlocks"`
	SentBlockRequests         uint32 `json:"sentBlockRequests"`
	SentMilestoneRequests     uint32 `json:"sentMilestoneRequests"`
	SentHeartbeats            uint32 `json:"sentHeartbeats"`
	DroppedPackets            uint32 `json:"droppedPackets"`
}

// Info represents information about an ongoing gossip protocol.
type GossipInfo struct {
	Heartbeat *Heartbeat        `json:"heartbeat"`
	Metrics   PeerGossipMetrics `json:"metrics"`
}

// PeerInfo defines the response of a GET peer REST API call.
type PeerInfo struct {
	// The libp2p identifier of the peer.
	ID string `json:"id"`
	// The libp2p multi addresses of the peer.
	MultiAddresses []string `json:"multiAddresses"`
	// The alias of the peer.
	Alias *string `json:"alias,omitempty"`
	// The relation (static, autopeered) of the peer.
	Relation string `json:"relation"`
	// Whether the peer is connected.
	Connected bool `json:"connected"`
	// The gossip protocol information of the peer.
	Gossip *GossipInfo `json:"gossip,omitempty"`
}
