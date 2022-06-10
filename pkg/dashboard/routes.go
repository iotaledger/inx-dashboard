package dashboard

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/iota.go/v3/nodeclient"
)

const (
	APIRoute                 = "/api/v2"
	PluginIndexerRoute       = "/api/plugins/indexer/v1"
	PluginParticipationRoute = "/api/plugins/participation/v1"
	PluginSpammerRoute       = "/api/plugins/spammer/v1"
)

const (
	// ParameterBlockID is used to identify a block by its ID.
	ParameterBlockID = "blockID"

	// ParameterTransactionID is used to identify a transaction by its ID.
	ParameterTransactionID = "transactionID"

	// ParameterOutputID is used to identify an output by its ID.
	ParameterOutputID = "outputID"

	// ParameterMilestoneIndex is used to identify a milestone by index.
	ParameterMilestoneIndex = "milestoneIndex"

	// ParameterMilestoneID is used to identify a milestone by its ID.
	ParameterMilestoneID = "milestoneID"

	// ParameterPeerID is used to identify a peer.
	ParameterPeerID = "peerID"

	// RouteInfo is the route for getting the node info.
	// GET returns the node info.
	RouteInfo = APIRoute + "/info"

	// RouteBlock is the route for getting a block by its blockID.
	// GET returns the block based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteBlock = APIRoute + "/blocks/:" + ParameterBlockID

	// RouteBlockMetadata is the route for getting block metadata by its blockID.
	// GET returns block metadata (including info about "promotion/reattachment needed").
	RouteBlockMetadata = APIRoute + "/blocks/:" + ParameterBlockID + "/metadata"

	// RouteTransactionsIncludedBlock is the route for getting the block that was included in the ledger for a given transaction ID.
	// GET returns the block based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteTransactionsIncludedBlock = APIRoute + "/transactions/:" + ParameterTransactionID + "/included-block"

	// RouteMilestoneByID is the route for getting a milestone by its ID.
	// GET returns the milestone.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteMilestoneByID = APIRoute + "/milestones/:" + ParameterMilestoneID

	// RouteMilestoneByIndex is the route for getting a milestone by its milestoneIndex.
	// GET returns the milestone.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteMilestoneByIndex = APIRoute + "/milestones/by-index/:" + ParameterMilestoneIndex

	// RouteOutput is the route for getting an output by its outputID (transactionHash + outputIndex).
	// GET returns the output based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteOutput = APIRoute + "/outputs/:" + ParameterOutputID

	// RoutePeer is the route for getting peers by their peerID.
	// GET returns the peer
	// DELETE deletes the peer.
	RoutePeer = APIRoute + "/peers/:" + ParameterPeerID

	// RoutePeers is the route for getting all peers of the node.
	// GET returns a list of all peers.
	// POST adds a new peer.
	RoutePeers = APIRoute + "/peers"
)

const (
	// ParameterFoundryID is used to identify a foundry by its ID.
	ParameterFoundryID = "foundryID"

	// ParameterAliasID is used to identify an alias by its ID.
	ParameterAliasID = "aliasID"

	// ParameterNFTID is used to identify a nft by its ID.
	ParameterNFTID = "nftID"

	// RouteOutputsBasic is the route for getting basic outputs filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// Query parameters: "address", "hasStorageReturnCondition", "storageReturnAddress", "hasExpirationCondition",
	//					 "expiresBefore", "expiresAfter", "expiresBeforeMilestone", "expiresAfterMilestone",
	//					 "hasTimelockCondition", "timelockedBefore", "timelockedAfter", "timelockedBeforeMilestone",
	//					 "timelockedAfterMilestone", "sender", "tag", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	RouteOutputsBasic = PluginIndexerRoute + "/outputs/basic"

	// RouteOutputsAliases is the route for getting aliases filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// Query parameters: "stateController", "governor", "issuer", "sender", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	RouteOutputsAliases = PluginIndexerRoute + "/outputs/alias"

	// RouteOutputsAliasByID is the route for getting aliases by their aliasID.
	// GET returns the outputIDs or 404 if no record is found.
	RouteOutputsAliasByID = PluginIndexerRoute + "/outputs/alias/:" + ParameterAliasID

	// RouteOutputsNFTs is the route for getting NFT filtered by the given parameters.
	// Query parameters: "address", "hasStorageReturnCondition", "storageReturnAddress", "hasExpirationCondition",
	//					 "expiresBefore", "expiresAfter", "expiresBeforeMilestone", "expiresAfterMilestone",
	//					 "hasTimelockCondition", "timelockedBefore", "timelockedAfter", "timelockedBeforeMilestone",
	//					 "timelockedAfterMilestone", "issuer", "sender", "tag", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	RouteOutputsNFTs = PluginIndexerRoute + "/outputs/nft"

	// RouteOutputsNFTByID is the route for getting NFT by their nftID.
	// GET returns the outputIDs or 404 if no record is found.
	RouteOutputsNFTByID = PluginIndexerRoute + "/outputs/nft/:" + ParameterNFTID

	// RouteOutputsFoundries is the route for getting foundries filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// Query parameters: "aliasAddress", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	RouteOutputsFoundries = PluginIndexerRoute + "/outputs/foundry"

	// RouteOutputsFoundryByID is the route for getting foundries by their foundryID.
	// GET returns the outputIDs or 404 if no record is found.
	RouteOutputsFoundryByID = PluginIndexerRoute + "/outputs/foundry/:" + ParameterFoundryID
)

const (
	// ParameterParticipationEventID is used to identify an event by its ID.
	ParameterParticipationEventID = "eventID"

	// RouteParticipationEvents is the route to list all events, returning their ID, the event name and status.
	// GET returns a list of all events known to the node. Optional query parameter returns filters events by type (query parameters: "type").
	RouteParticipationEvents = PluginParticipationRoute + "/events"

	// RouteParticipationEvent is the route to access a single participation by its ID.
	// GET gives a quick overview of the participation. This does not include the current standings.
	RouteParticipationEvent = PluginParticipationRoute + "/events/:" + ParameterParticipationEventID

	// RouteParticipationEventStatus is the route to access the status of a single participation by its ID.
	// GET returns the amount of tokens participating and accumulated votes for the ballot if the event contains a ballot. Optional query parameter returns the status for the given milestone index (query parameters: "milestoneIndex").
	RouteParticipationEventStatus = PluginParticipationRoute + "/events/:" + ParameterParticipationEventID + "/status"

	// RouteAdminCreateEvent is the route the node operator can use to add events.
	// POST creates a new event to track
	RouteAdminCreateEvent = PluginParticipationRoute + "/admin/events"

	// RouteAdminDeleteEvent is the route the node operator can use to remove events.
	// DELETE removes a tracked participation.
	RouteAdminDeleteEvent = PluginParticipationRoute + "/admin/events/:" + ParameterParticipationEventID
)

const (
	// RouteSpammerStatus is the route to get the status of the spammer.
	// GET the current status of the spammer.
	RouteSpammerStatus = PluginSpammerRoute + "/status"

	// RouteSpammerStart is the route to start the spammer (with optional changing the settings).
	// POST the settings to change and start the spammer.
	RouteSpammerStart = PluginSpammerRoute + "/start"

	// RouteSpammerStop is the route to stop the spammer.
	// POST to stop the spammer.
	RouteSpammerStop = PluginSpammerRoute + "/stop"
)

func (d *Dashboard) setupNodeRoutes(e *echo.Echo) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	indexerClient, err := d.nodeClient.Indexer(ctx)
	if err != nil {
		if err != nodeclient.ErrIndexerPluginNotAvailable {
			return err
		}
	}

	participationSupported, err := d.nodeClient.NodeSupportsPlugin(ctx, "participation/v1")
	if err != nil {
		return err
	}

	spammerSupported, err := d.nodeClient.NodeSupportsPlugin(ctx, "spammer/v1")
	if err != nil {
		return err
	}

	dashboardRouteGroup := e.Group("dashboard")

	dashboardRouteGroup.GET(RouteInfo, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	dashboardRouteGroup.GET(RouteBlockMetadata, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	dashboardRouteGroup.GET(RouteBlock, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	dashboardRouteGroup.GET(RouteTransactionsIncludedBlock, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	dashboardRouteGroup.GET(RouteMilestoneByID, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	dashboardRouteGroup.GET(RouteMilestoneByIndex, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	dashboardRouteGroup.GET(RouteOutput, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	dashboardRouteGroup.DELETE(RoutePeer, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	dashboardRouteGroup.POST(RoutePeers, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	if indexerClient != nil {

		dashboardRouteGroup.GET(RouteOutputsBasic, func(c echo.Context) error {
			return d.forwardRequest(c)
		})
		dashboardRouteGroup.GET(RouteOutputsAliases, func(c echo.Context) error {
			return d.forwardRequest(c)
		})
		dashboardRouteGroup.GET(RouteOutputsAliasByID, func(c echo.Context) error {
			return d.forwardRequest(c)
		})
		dashboardRouteGroup.GET(RouteOutputsNFTs, func(c echo.Context) error {
			return d.forwardRequest(c)
		})
		dashboardRouteGroup.GET(RouteOutputsNFTByID, func(c echo.Context) error {
			return d.forwardRequest(c)
		})
		dashboardRouteGroup.GET(RouteOutputsFoundries, func(c echo.Context) error {
			return d.forwardRequest(c)
		})
		dashboardRouteGroup.GET(RouteOutputsFoundryByID, func(c echo.Context) error {
			return d.forwardRequest(c)
		})
	}

	if participationSupported {
		dashboardRouteGroup.GET(RouteParticipationEvents, func(c echo.Context) error {
			return d.forwardRequest(c)
		})

		dashboardRouteGroup.GET(RouteParticipationEvent, func(c echo.Context) error {
			return d.forwardRequest(c)
		})

		dashboardRouteGroup.GET(RouteParticipationEventStatus, func(c echo.Context) error {
			return d.forwardRequest(c)
		})

		dashboardRouteGroup.POST(RouteAdminCreateEvent, func(c echo.Context) error {
			return d.forwardRequest(c)
		})

		dashboardRouteGroup.DELETE(RouteAdminDeleteEvent, func(c echo.Context) error {
			return d.forwardRequest(c)
		})
	}

	if spammerSupported {
		dashboardRouteGroup.GET(RouteSpammerStatus, func(c echo.Context) error {
			return d.forwardRequest(c)
		})

		dashboardRouteGroup.POST(RouteSpammerStart, func(c echo.Context) error {
			return d.forwardRequest(c)
		})

		dashboardRouteGroup.POST(RouteSpammerStop, func(c echo.Context) error {
			return d.forwardRequest(c)
		})
	}

	return nil

}

func readAndCloseRequestBody(res *http.Request) ([]byte, error) {
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}
	return resBody, nil
}

func readAndCloseResponseBody(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}
	return resBody, nil
}

func (d *Dashboard) forwardRequest(c echo.Context) error {

	reqBody, err := readAndCloseRequestBody(c.Request())
	if err != nil {
		return err
	}

	// construct request URL
	url := fmt.Sprintf("%s%s", d.nodeClient.BaseURL, strings.Replace(c.Request().URL.Path, "/dashboard", "", 1))

	// construct request
	req, err := http.NewRequestWithContext(c.Request().Context(), c.Request().Method, url, func() io.Reader {
		if reqBody == nil {
			return nil
		}
		return bytes.NewReader(reqBody)
	}())
	if err != nil {
		return fmt.Errorf("unable to build http request: %w", err)
	}

	if c.Request().URL.User != nil {
		// set the userInfo for basic auth
		req.URL.User = c.Request().URL.User
	}

	contentType := c.Request().Header.Get("Content-Type")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// make the request
	res, err := d.nodeClient.HTTPClient().Do(req)
	if err != nil {
		return err
	}

	resBody, err := readAndCloseResponseBody(res)
	if err != nil {
		return err
	}

	return c.JSONBlob(res.StatusCode, resBody)
}
