package dashboard

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	FeatureCoreAPI          = "core/v2"
	FeatureDashboardMetrics = "dashboard-metrics/v1"
	FeatureIndexer          = "indexer/v1"
	FeatureParticipation    = "participation/v1"
	FeatureSpammer          = "spammer/v1"

	BasePath              = "/api"
	CoreAPIRoute          = BasePath + "/" + FeatureCoreAPI
	DashboardMetricsRoute = BasePath + "/" + FeatureDashboardMetrics
	IndexerRoute          = BasePath + "/" + FeatureIndexer
	ParticipationRoute    = BasePath + "/" + FeatureParticipation
	SpammerRoute          = BasePath + "/" + FeatureSpammer
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

	// RouteRoutes is the route for getting the routes the node exposes.
	// GET returns the routes.
	RouteRoutes = BasePath + "/routes"

	// RouteCoreInfo is the route for getting the node info.
	// GET returns the node info.
	RouteCoreInfo = CoreAPIRoute + "/info"

	// RouteCoreBlock is the route for getting a block by its blockID.
	// GET returns the block based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteCoreBlock = CoreAPIRoute + "/blocks/:" + ParameterBlockID

	// RouteCoreBlockMetadata is the route for getting block metadata by its blockID.
	// GET returns block metadata (including info about "promotion/reattachment needed").
	RouteCoreBlockMetadata = CoreAPIRoute + "/blocks/:" + ParameterBlockID + "/metadata"

	// RouteCoreTransactionsIncludedBlock is the route for getting the block that was included in the ledger for a given transaction ID.
	// GET returns the block based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteCoreTransactionsIncludedBlock = CoreAPIRoute + "/transactions/:" + ParameterTransactionID + "/included-block"

	// RouteCoreMilestoneByID is the route for getting a milestone by its ID.
	// GET returns the milestone.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteCoreMilestoneByID = CoreAPIRoute + "/milestones/:" + ParameterMilestoneID

	// RouteCoreMilestoneByIndex is the route for getting a milestone by its milestoneIndex.
	// GET returns the milestone.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteCoreMilestoneByIndex = CoreAPIRoute + "/milestones/by-index/:" + ParameterMilestoneIndex

	// RouteCoreOutput is the route for getting an output by its outputID (transactionHash + outputIndex).
	// GET returns the output based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteCoreOutput = CoreAPIRoute + "/outputs/:" + ParameterOutputID

	// RouteCorePeer is the route for getting peers by their peerID.
	// GET returns the peer
	// DELETE deletes the peer.
	RouteCorePeer = CoreAPIRoute + "/peers/:" + ParameterPeerID

	// RouteCorePeers is the route for getting all peers of the node.
	// GET returns a list of all peers.
	// POST adds a new peer.
	RouteCorePeers = CoreAPIRoute + "/peers"
)

const (
	// ParameterFoundryID is used to identify a foundry by its ID.
	ParameterFoundryID = "foundryID"

	// ParameterAliasID is used to identify an alias by its ID.
	ParameterAliasID = "aliasID"

	// ParameterNFTID is used to identify a nft by its ID.
	ParameterNFTID = "nftID"

	// RouteIndexerOutputsBasic is the route for getting basic outputs filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// Query parameters: "address", "hasStorageReturnCondition", "storageReturnAddress", "hasExpirationCondition",
	//					 "expiresBefore", "expiresAfter", "expiresBeforeMilestone", "expiresAfterMilestone",
	//					 "hasTimelockCondition", "timelockedBefore", "timelockedAfter", "timelockedBeforeMilestone",
	//					 "timelockedAfterMilestone", "sender", "tag", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	RouteIndexerOutputsBasic = IndexerRoute + "/outputs/basic"

	// RouteIndexerOutputsAliases is the route for getting aliases filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// Query parameters: "stateController", "governor", "issuer", "sender", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	RouteIndexerOutputsAliases = IndexerRoute + "/outputs/alias"

	// RouteIndexerOutputsAliasByID is the route for getting aliases by their aliasID.
	// GET returns the outputIDs or 404 if no record is found.
	RouteIndexerOutputsAliasByID = IndexerRoute + "/outputs/alias/:" + ParameterAliasID

	// RouteIndexerOutputsNFTs is the route for getting NFT filtered by the given parameters.
	// Query parameters: "address", "hasStorageReturnCondition", "storageReturnAddress", "hasExpirationCondition",
	//					 "expiresBefore", "expiresAfter", "expiresBeforeMilestone", "expiresAfterMilestone",
	//					 "hasTimelockCondition", "timelockedBefore", "timelockedAfter", "timelockedBeforeMilestone",
	//					 "timelockedAfterMilestone", "issuer", "sender", "tag", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	RouteIndexerOutputsNFTs = IndexerRoute + "/outputs/nft"

	// RouteIndexerOutputsNFTByID is the route for getting NFT by their nftID.
	// GET returns the outputIDs or 404 if no record is found.
	RouteIndexerOutputsNFTByID = IndexerRoute + "/outputs/nft/:" + ParameterNFTID

	// RouteIndexerOutputsFoundries is the route for getting foundries filtered by the given parameters.
	// GET with query parameter returns all outputIDs that fit these filter criteria.
	// Query parameters: "aliasAddress", "createdBefore", "createdAfter"
	// Returns an empty list if no results are found.
	RouteIndexerOutputsFoundries = IndexerRoute + "/outputs/foundry"

	// RouteIndexerOutputsFoundryByID is the route for getting foundries by their foundryID.
	// GET returns the outputIDs or 404 if no record is found.
	RouteIndexerOutputsFoundryByID = IndexerRoute + "/outputs/foundry/:" + ParameterFoundryID
)

const (
	// RouteDashboardNodeInfoExtended is the route to get additional info about the node.
	// GET returns the extended info of the node.
	RouteDashboardNodeInfoExtended = DashboardMetricsRoute + "/info"

	// RouteDashboardDatabaseSizes is the route to get the size of the databases.
	// GET returns the sizes of the databases.
	RouteDashboardDatabaseSizes = DashboardMetricsRoute + "/database/sizes"

	// RouteDashboardGossipMetrics is the route to get metrics about gossip.
	// GET returns the gossip metrics.
	RouteDashboardGossipMetrics = DashboardMetricsRoute + "/gossip"
)

const (
	// ParameterParticipationEventID is used to identify an event by its ID.
	ParameterParticipationEventID = "eventID"

	// RouteParticipationEvents is the route to list all events, returning their ID, the event name and status.
	// GET returns a list of all events known to the node. Optional query parameter returns filters events by type (query parameters: "type").
	RouteParticipationEvents = ParticipationRoute + "/events"

	// RouteParticipationEvent is the route to access a single participation by its ID.
	// GET gives a quick overview of the participation. This does not include the current standings.
	RouteParticipationEvent = ParticipationRoute + "/events/:" + ParameterParticipationEventID

	// RouteParticipationEventStatus is the route to access the status of a single participation by its ID.
	// GET returns the amount of tokens participating and accumulated votes for the ballot if the event contains a ballot. Optional query parameter returns the status for the given milestone index (query parameters: "milestoneIndex").
	RouteParticipationEventStatus = ParticipationRoute + "/events/:" + ParameterParticipationEventID + "/status"

	// RouteParticipationAdminCreateEvent is the route the node operator can use to add events.
	// POST creates a new event to track
	RouteParticipationAdminCreateEvent = ParticipationRoute + "/admin/events"

	// RouteParticipationAdminDeleteEvent is the route the node operator can use to remove events.
	// DELETE removes a tracked participation.
	RouteParticipationAdminDeleteEvent = ParticipationRoute + "/admin/events/:" + ParameterParticipationEventID
)

const (
	// RouteSpammerStatus is the route to get the status of the spammer.
	// GET the current status of the spammer.
	RouteSpammerStatus = SpammerRoute + "/status"

	// RouteSpammerStart is the route to start the spammer (with optional changing the settings).
	// POST the settings to change and start the spammer.
	RouteSpammerStart = SpammerRoute + "/start"

	// RouteSpammerStop is the route to stop the spammer.
	// POST to stop the spammer.
	RouteSpammerStop = SpammerRoute + "/stop"
)

func (d *Dashboard) setupAPIRoutes(routeGroup *echo.Group) error {

	routeGroup.GET(RouteRoutes, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteCoreInfo, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteCoreBlockMetadata, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteCoreBlock, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteCoreTransactionsIncludedBlock, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteCoreMilestoneByID, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteCoreMilestoneByIndex, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteCoreOutput, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.DELETE(RouteCorePeer, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.POST(RouteCorePeers, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	// dashboard metrics
	routeGroup.GET(RouteDashboardNodeInfoExtended, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteDashboardDatabaseSizes, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteDashboardGossipMetrics, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	// indexer
	routeGroup.GET(RouteIndexerOutputsBasic, func(c echo.Context) error {
		return d.forwardRequest(c)
	})
	routeGroup.GET(RouteIndexerOutputsAliases, func(c echo.Context) error {
		return d.forwardRequest(c)
	})
	routeGroup.GET(RouteIndexerOutputsAliasByID, func(c echo.Context) error {
		return d.forwardRequest(c)
	})
	routeGroup.GET(RouteIndexerOutputsNFTs, func(c echo.Context) error {
		return d.forwardRequest(c)
	})
	routeGroup.GET(RouteIndexerOutputsNFTByID, func(c echo.Context) error {
		return d.forwardRequest(c)
	})
	routeGroup.GET(RouteIndexerOutputsFoundries, func(c echo.Context) error {
		return d.forwardRequest(c)
	})
	routeGroup.GET(RouteIndexerOutputsFoundryByID, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	// participation
	routeGroup.GET(RouteParticipationEvents, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteParticipationEvent, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.GET(RouteParticipationEventStatus, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.POST(RouteParticipationAdminCreateEvent, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.DELETE(RouteParticipationAdminDeleteEvent, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	// spammmer
	routeGroup.GET(RouteSpammerStatus, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.POST(RouteSpammerStart, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

	routeGroup.POST(RouteSpammerStop, func(c echo.Context) error {
		return d.forwardRequest(c)
	})

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
	url := fmt.Sprintf("%s%s", d.nodeClient.BaseURL, strings.Replace(c.Request().RequestURI, "/dashboard", "", 1))

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
