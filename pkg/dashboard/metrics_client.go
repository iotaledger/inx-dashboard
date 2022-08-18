package dashboard

import (
	"context"
	"net/http"

	"github.com/iotaledger/iota.go/v3/nodeclient"
)

// NewMetricsClient returns a new dashboard metrics node API instance.
func NewMetricsClient(client *nodeclient.Client) *MetricsClient {
	return &MetricsClient{Client: client}
}

// MetricsClient is an API wrapper over the dashboard metrics node API.
type MetricsClient struct {
	*nodeclient.Client
}

func (client *MetricsClient) NodeInfoExtended(ctx context.Context) (*NodeInfoExtended, error) {
	res := &NodeInfoExtended{}
	//nolint:bodyclose // false positive, it is done in the client.Do method
	if _, err := client.Do(ctx, http.MethodGet, RouteDashboardNodeInfoExtended, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *MetricsClient) DatabaseSizes(ctx context.Context) (*DatabaseSizesMetric, error) {
	res := &DatabaseSizesMetric{}
	//nolint:bodyclose // false positive, it is done in the client.Do method
	if _, err := client.Do(ctx, http.MethodGet, RouteDashboardDatabaseSizes, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *MetricsClient) GossipMetrics(ctx context.Context) (*GossipMetrics, error) {
	res := &GossipMetrics{}
	//nolint:bodyclose // false positive, it is done in the client.Do method
	if _, err := client.Do(ctx, http.MethodGet, RouteDashboardGossipMetrics, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}
