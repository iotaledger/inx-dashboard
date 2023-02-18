package dashboard

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"nhooyr.io/websocket"

	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/syncutils"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/inx-dashboard/pkg/jwt"
)

var (
	timeoutNodeInfos = 2 * time.Second
)

type WebSocketMsgType byte

const (
	// MsgTypeSyncStatus is the type of the SyncStatus message.
	MsgTypeSyncStatus WebSocketMsgType = iota
	// MsgTypePublicNodeStatus is the type of the PublicNodeStatus message.
	MsgTypePublicNodeStatus
	// MsgTypeNodeInfoExtended is the type of the NodeInfoExtended message.
	MsgTypeNodeInfoExtended
	// MsgTypeGossipMetrics is the type of the GossipMetrics message.
	MsgTypeGossipMetrics
	// MsgTypeMilestone is the type of the Milestone message.
	MsgTypeMilestone
	// MsgTypePeerMetric is the type of the PeerMetric message.
	MsgTypePeerMetric
	// MsgTypeConfirmedMsMetrics is the type of the ConfirmedMsMetrics message.
	MsgTypeConfirmedMsMetrics
	// MsgTypeVisualizerVertex is the type of the Vertex message for the visualizer.
	MsgTypeVisualizerVertex
	// MsgTypeVisualizerSolidInfo is the type of the SolidInfo message for the visualizer.
	MsgTypeVisualizerSolidInfo
	// MsgTypeVisualizerConfirmedInfo is the type of the ConfirmedInfo message for the visualizer.
	MsgTypeVisualizerConfirmedInfo
	// MsgTypeVisualizerMilestoneInfo is the type of the MilestoneInfo message for the visualizer.
	MsgTypeVisualizerMilestoneInfo
	// MsgTypeVisualizerTipInfo is the type of the TipInfo message for the visualizer.
	MsgTypeVisualizerTipInfo
	// MsgTypeDatabaseSizeMetric is the type of the database Size message for the metrics.
	MsgTypeDatabaseSizeMetric
)

func (d *Dashboard) websocketRoute(ctx echo.Context) error {
	defer func() {
		if r := recover(); r != nil {
			d.LogErrorf("recovered from panic within WS handle func: %s", r)
		}
	}()

	publicTopics := []WebSocketMsgType{
		MsgTypeSyncStatus,
		MsgTypePublicNodeStatus,
		MsgTypeGossipMetrics,
		MsgTypeMilestone,
		MsgTypeConfirmedMsMetrics,
		MsgTypeVisualizerVertex,
		MsgTypeVisualizerSolidInfo,
		MsgTypeVisualizerConfirmedInfo,
		MsgTypeVisualizerMilestoneInfo,
		MsgTypeVisualizerTipInfo,
	}

	isProtectedTopic := func(topic WebSocketMsgType) bool {
		for _, publicTopic := range publicTopics {
			if topic == publicTopic {
				return false
			}
		}

		return true
	}

	// this function sends the initial values for some topics
	sendInitValue := func(client *websockethub.Client, initValuesSent map[WebSocketMsgType]struct{}, topic WebSocketMsgType) {
		// always send the initial values for the Vertex topic, ignore others that were already sent
		if _, sent := initValuesSent[topic]; sent && (topic != MsgTypeVisualizerVertex) {
			return
		}
		initValuesSent[topic] = struct{}{}

		ctxNodeInfos, ctxNodeInfosCancel := context.WithTimeout(client.Context(), timeoutNodeInfos)
		defer ctxNodeInfosCancel()

		ctxMsg, ctxMsgCancel := context.WithTimeout(client.Context(), d.websocketWriteTimeout)
		defer ctxMsgCancel()

		//nolint:exhaustive // false positive
		switch topic {

		case MsgTypeSyncStatus:
			_ = client.Send(ctxMsg, &Msg{Type: MsgTypeSyncStatus, Data: d.getSyncStatus()})

		case MsgTypePublicNodeStatus:
			nodeInfo, err := d.getNodeInfo(ctxNodeInfos)
			if err != nil {
				d.LogWarnf("failed to get node info: %s", err)

				return
			}

			data := getPublicNodeStatusByNodeInfo(nodeInfo, d.nodeBridge.IsNodeAlmostSynced())
			_ = d.hub.BroadcastMsg(ctxMsg, &Msg{Type: MsgTypePublicNodeStatus, Data: data})

		case MsgTypeNodeInfoExtended:
			data, err := d.getNodeInfoExtended(ctxNodeInfos)
			if err != nil {
				d.LogWarnf("failed to get extended node info: %s", err)

				return
			}
			_ = client.Send(ctxMsg, &Msg{Type: MsgTypeNodeInfoExtended, Data: data})

		case MsgTypeGossipMetrics:
			data, err := d.getGossipMetrics(ctxNodeInfos)
			if err != nil {
				d.LogWarnf("failed to get gossip metrics: %s", err)

				return
			}
			_ = client.Send(ctxMsg, &Msg{Type: MsgTypeGossipMetrics, Data: data})

		case MsgTypeMilestone:
			start := d.getLatestMilestoneIndex()
			for msIndex := start - 10; msIndex <= start; msIndex++ {
				if milestoneIDHex, err := d.getMilestoneIDHex(ctxNodeInfos, msIndex); err == nil {
					_ = client.Send(ctxMsg, &Msg{Type: MsgTypeMilestone, Data: &Milestone{MilestoneID: milestoneIDHex, Index: msIndex}})
				} else {
					d.LogWarnf("failed to get milestone %d: %s", msIndex, err)

					return
				}
			}

		case MsgTypePeerMetric:
			data, err := d.getPeerInfos(ctxNodeInfos)
			if err != nil {
				d.LogWarnf("failed to get peer infos: %s", err)

				return
			}
			_ = client.Send(ctxMsg, &Msg{Type: MsgTypePeerMetric, Data: data})

		case MsgTypeConfirmedMsMetrics:
			data, err := d.getNodeInfo(ctxNodeInfos)
			if err != nil {
				d.LogWarnf("failed to get node info: %s", err)

				return
			}
			_ = client.Send(ctxMsg, &Msg{Type: MsgTypeConfirmedMsMetrics, Data: data.Metrics})

		case MsgTypeVisualizerVertex:
			d.visualizer.ForEachCreated(func(vertex *VisualizerVertex) bool {
				// don't drop the messages to fill the visualizer without missing any vertex
				_ = client.Send(ctxMsg, &Msg{Type: MsgTypeVisualizerVertex, Data: vertex}, true)

				return true
			}, VisualizerInitValuesCount)

		case MsgTypeDatabaseSizeMetric:
			_ = client.Send(ctxMsg, &Msg{Type: MsgTypeDatabaseSizeMetric, Data: d.cachedDatabaseSizeMetrics})
		}
	}

	topicsLock := syncutils.RWMutex{}
	registeredTopics := make(map[WebSocketMsgType]struct{})
	initValuesSent := make(map[WebSocketMsgType]struct{})

	d.hub.Events().ClientConnected.Attach(event.NewClosure(func(event *websockethub.ClientConnectionEvent) {
		d.LogDebugf("WebSocket client (ID: %d) connection established", event.ID)
		d.subscriptionManager.Connect(event.ID)
	}))

	d.hub.Events().ClientDisconnected.Attach(event.NewClosure(func(event *websockethub.ClientConnectionEvent) {
		d.LogDebugf("WebSocket client (ID: %d) connection closed", event.ID)
		d.subscriptionManager.Disconnect(event.ID)
	}))

	return d.hub.ServeWebsocket(ctx.Response(), ctx.Request(),
		// onCreate gets called when the client is created
		func(client *websockethub.Client) {
			client.FilterCallback = func(_ *websockethub.Client, data interface{}) bool {
				msg, ok := data.(*Msg)
				if !ok {
					return false
				}

				topicsLock.RLock()
				_, registered := registeredTopics[msg.Type]
				topicsLock.RUnlock()

				return registered
			}
			client.ReceiveChan = make(chan *websockethub.WebsocketMsg, 100)

			go func() {
				for {
					// we need to nest the client.ReceiveChan into the default case because
					// the select cases are executed in random order if multiple
					// conditions are true at the time of entry in the select case.
					select {
					case <-client.ExitSignal:
						// client was disconnected
						return
					default:
						select {
						case <-client.ExitSignal:
							// client was disconnected
							return
						case msg, ok := <-client.ReceiveChan:
							if !ok {
								// client was disconnected
								return
							}

							if msg.MsgType == websocket.MessageBinary {
								if len(msg.Data) < 2 {
									continue
								}

								cmd := msg.Data[0]
								topic := WebSocketMsgType(msg.Data[1])

								if cmd == WebsocketCmdRegister {

									if isProtectedTopic(topic) {
										// Check for the presence of a JWT and verify it
										if len(msg.Data) < 3 {
											// Dot not allow unsecure subscriptions to protected topics
											continue
										}
										token := string(msg.Data[2:])
										if !d.jwtAuth.VerifyJWT(token, func(claims *jwt.AuthClaims) bool {
											return true
										}) {
											// Dot not allow unsecure subscriptions to protected topics
											continue
										}
									}

									// register topic fo this client
									d.subscriptionManager.Subscribe(client.ID(), topic)

									topicsLock.Lock()
									registeredTopics[topic] = struct{}{}
									topicsLock.Unlock()

									sendInitValue(client, initValuesSent, topic)

								} else if cmd == WebsocketCmdUnregister {

									// unregister topic fo this client
									d.subscriptionManager.Unsubscribe(client.ID(), topic)

									topicsLock.Lock()
									delete(registeredTopics, topic)
									topicsLock.Unlock()
								}
							}
						}
					}
				}
			}()
		},

		// onConnect gets called when the client was registered
		nil,

		// onDisconnect gets called when the client was disconnected
		nil,
	)
}
