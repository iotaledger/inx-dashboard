package dashboard

// Msg represents a websocket message.
type Msg struct {
	Type byte        `json:"type"`
	Data interface{} `json:"data"`
}

// PublicNodeStatus represents the public node status.
type PublicNodeStatus struct {
	PruningIndex uint32 `json:"pruningIndex"`
	IsHealthy    bool   `json:"isHealthy"`
	IsSynced     bool   `json:"isSynced"`
}

// NodeInfoExtended represents extended information about the node.
type NodeInfoExtended struct {
	Version       string `json:"version"`
	LatestVersion string `json:"latestVersion"`
	Uptime        int64  `json:"uptime"`
	NodeID        string `json:"nodeId"`
	NodeAlias     string `json:"nodeAlias"`
	MemoryUsage   int64  `json:"memUsage"`
}

// SyncStatus represents the node sync status.
type SyncStatus struct {
	CMI uint32 `json:"cmi"`
	LMI uint32 `json:"lmi"`
}

// GossipMetrics represents a websocket message.
type GossipMetrics struct {
	Incoming uint32 `json:"incoming"`
	New      uint32 `json:"new"`
	Outgoing uint32 `json:"outgoing"`
}

// Milestone represents a milestone for the livefeed.
type Milestone struct {
	MilestoneID string `json:"milestoneId"`
	Index       uint32 `json:"index"`
}

// DatabaseSizesMetric represents database size metrics.
type DatabaseSizesMetric struct {
	Tangle int64 `json:"tangle"`
	UTXO   int64 `json:"utxo"`
	Total  int64 `json:"total"`
	Time   int64 `json:"ts"`
}

// VisualizerVertex defines a vertex in a DAG.
type VisualizerVertex struct {
	ID                   string   `json:"id"`
	Parents              []string `json:"parents"`
	IsSolid              bool     `json:"isSolid"`
	IsReferenced         bool     `json:"isReferenced"`
	IsConflicting        bool     `json:"isConflicting"`
	IsTransaction        bool     `json:"isTransaction"`
	IsMilestone          bool     `json:"isMilestone"`
	IsTip                bool     `json:"isTip"`
	shortID              string
	isCreated            bool
	isReferencedByOthers bool
}

// VisualizerMetaInfo signals that metadata of a given block changed.
type VisualizerMetaInfo struct {
	ID string `json:"id"`
}

// tipinfo holds information about whether a given block is a tip or not.
type VisualizerTipInfo struct {
	ID    string `json:"id"`
	IsTip bool   `json:"isTip"`
}

// VisualizerConfirmationInfo signals confirmation of a milestone block with a list of exluded blocks in the past cone.
type VisualizerConfirmationInfo struct {
	IDs         []string `json:"ids"`
	ExcludedIDs []string `json:"excludedIds"`
}
