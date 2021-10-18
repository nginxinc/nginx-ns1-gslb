package agent

import (
	"reflect"
	"testing"

	"github.com/nginxinc/nginx-ns1-gslb/internal"
	"github.com/nginxinc/nginx-ns1-gslb/internal/input"
	"github.com/nginxinc/nginx-ns1-gslb/internal/output"
	"github.com/nginxinc/nginx-plus-go-client/client"
)

func TestNewAgentFailureNoNGINXPlus(t *testing.T) {
	globalCfg := &Config{
		Agent: Cfg{},
		NginxPlus: input.Cfg{
			Hosts: []input.NginxHost{{Host: "nginxplushost"}},
		},
		Nsone: output.Cfg{},
	}
	_, err := New(globalCfg)
	if err == nil {
		t.Errorf("Agent creation err returned %+v but expected an error because NGINX Plus is not reachable", err)
	}
}

func TestProcessData(t *testing.T) {
	testCases := []struct {
		input    []*client.Stats
		expected map[string]*internal.FeedData
		nType    string
		msg      string
	}{
		{
			input: createExampleStatsSlice(1, false),
			nType: "global",
			expected: map[string]*internal.FeedData{
				"feed01": {
					Connections: 0,
					Up:          true,
				},
			},
			msg: "Global connections",
		},
		{
			input: createExampleStatsSlice(1, false),
			expected: map[string]*internal.FeedData{
				"feed01": {
					Connections: 3,
					Up:          true,
				},
			},
			nType: "upstream_groups",
			msg:   "Change resource with feed name for Upstreams/Status Zones",
		},
	}

	servicesFeedsMap := map[string]string{
		"service01": "feed01",
	}
	agent := &Agent{
		namedServices: servicesFeedsMap,
		services: Services{
			Threshold: 1,
		},
	}

	for _, testCase := range testCases {
		agent.services.Method = testCase.nType
		feedData, _ := agent.processData(testCase.input)
		if !reflect.DeepEqual(testCase.expected, feedData) {
			t.Errorf("Agent.processData returned %v, but %v expected for case: %v", feedData, testCase.expected, testCase.msg)
		}
	}
}

func TestGetUpstreamConnectionsData(t *testing.T) {
	namedServices := map[string]string{
		"service01": "feed01",
		"service02": "feed02",
	}

	testUpstreamConnections := []struct {
		statsSlice       []*client.Stats
		mergeMethod      string
		minPeerThreshold int
		msg              string
		expected         map[string]*internal.FeedData
	}{
		{
			statsSlice:       createExampleStatsSlice(1, false),
			mergeMethod:      "count",
			minPeerThreshold: 1,
			msg:              "1 NGINX Plus instance with 2 Upstreams with 3 available peers using method: count",
			expected: map[string]*internal.FeedData{
				"service01": {
					Connections: 3,
					Up:          true,
				},
				"service02": {
					Connections: 3,
					Up:          true,
				},
			},
		},
		{
			statsSlice:       createExampleStatsSlice(1, false),
			mergeMethod:      "avg",
			minPeerThreshold: 0,
			msg:              "1 NGINX Plus instance with 2 Upstreams with 3 available peers using method: avg",
			expected: map[string]*internal.FeedData{
				"service01": {
					Connections: 1,
					Up:          true,
				},
				"service02": {
					Connections: 1,
					Up:          true,
				},
			},
		},
		{
			statsSlice:       createExampleStatsSlice(2, false),
			mergeMethod:      "count",
			minPeerThreshold: 0,
			msg:              "2 NGINX Plus instances with 2 Upstreams with 3 available peers using method: count",
			expected: map[string]*internal.FeedData{
				"service01": {
					Connections: 6,
					Up:          true,
				},
				"service02": {
					Connections: 6,
					Up:          true,
				},
			},
		},
		{
			statsSlice:       createExampleStatsSlice(2, false),
			mergeMethod:      "avg",
			minPeerThreshold: 0,
			msg:              "2 NGINX Plus instances with 2 Upstreams with 3 available peers using method: avg",
			expected: map[string]*internal.FeedData{
				"service01": {
					Connections: 2,
					Up:          true,
				},
				"service02": {
					Connections: 2,
					Up:          true,
				},
			},
		},
		{
			statsSlice:       createExampleStatsSlice(1, true),
			mergeMethod:      "count",
			minPeerThreshold: 3,
			msg:              "1 NGINX Plus instance with 2 Upstreams with 2 available peers and 1 unavailable (min peer threshold 3) using method: count",
			expected: map[string]*internal.FeedData{
				"service01": {
					Connections: 1,
					Up:          false,
				},
				"service02": {
					Connections: 1,
					Up:          false,
				},
			},
		},
		{
			statsSlice:       createExampleStatsSlice(2, true),
			mergeMethod:      "avg",
			minPeerThreshold: 3,
			msg:              "1 NGINX Plus instance with 2 Upstreams with 2 available peers and 1 unavailable (min peer threshold 3) using method: avg",
			expected: map[string]*internal.FeedData{
				"service01": {
					Connections: 1,
					Up:          false,
				},
				"service02": {
					Connections: 1,
					Up:          false,
				},
			},
		},
	}

	for _, testCase := range testUpstreamConnections {
		feedData := getUpstreamConnectionsData(testCase.statsSlice, testCase.mergeMethod, namedServices, testCase.minPeerThreshold)
		if !reflect.DeepEqual(testCase.expected, feedData) {
			t.Errorf("getUpstreamConnectionsData returned %v, but %v expected for case: %v", feedData, testCase.expected, testCase.msg)
		}
	}
}

func TestGetGlobalConnectionsData(t *testing.T) {
	testGlobalConnections := []struct {
		statsSlice []*client.Stats
		msg        string
		expected   map[string]*internal.FeedData
	}{
		{
			statsSlice: createExampleStatsSlice(1, false),
			msg:        "1 NGINX Plus instance with 0 active connections",
			expected: map[string]*internal.FeedData{
				"global": {
					Connections: 0,
					Up:          true,
				},
			},
		},
		{
			statsSlice: createExampleStatsSlice(2, false),
			msg:        "2 NGINX Plus instances with 1 active connection",
			expected: map[string]*internal.FeedData{
				"global": {
					Connections: 1,
					Up:          true,
				},
			},
		},
	}

	for _, testCase := range testGlobalConnections {
		feedData := getGlobalConnectionsData(testCase.statsSlice)
		if !reflect.DeepEqual(testCase.expected, feedData) {
			t.Errorf("getGlobalConnectionsData returned %v, but %v expected for case: %v", feedData, testCase.expected, testCase.msg)
		}
	}
}

func TestGetStatusZonesConnectionsData(t *testing.T) {
	namedServices := map[string]string{
		"zone1.org": "feed01",
		"zone2.org": "feed02",
	}

	testStatusZonesConnections := []struct {
		statsSlice []*client.Stats
		msg        string
		expected   map[string]*internal.FeedData
	}{
		{
			statsSlice: createExampleStatsSlice(1, false),
			msg:        "1 NGINX Plus instance with 2 Server Zones",
			expected: map[string]*internal.FeedData{
				"zone1.org": {
					Connections: 0,
					Up:          true,
				},
				"zone2.org": {
					Connections: 1,
					Up:          true,
				},
			},
		},
		{
			statsSlice: createExampleStatsSlice(2, false),
			msg:        "2 NGINX Plus instances with 2 Server Zones",
			expected: map[string]*internal.FeedData{
				"zone1.org": {
					Connections: 1,
					Up:          true,
				},
				"zone2.org": {
					Connections: 3,
					Up:          true,
				},
			},
		},
	}

	for _, testCase := range testStatusZonesConnections {
		feedData := getStatusZonesConnectionsData(testCase.statsSlice, namedServices)
		if !reflect.DeepEqual(testCase.expected, feedData) {
			t.Errorf("getStatusZonesConnectionsData returned %v, but %v expected for case: %v", feedData, testCase.expected, testCase.msg)
		}
	}
}

func TestMergeStatsWrongType(t *testing.T) {
	slice := createExampleStatsSlice(1, false)
	a := createAgentWithServices("", "", 0)
	_, err := a.mergeStats(slice)
	if err == nil {
		t.Errorf("mergeStats err is nil, but error expected for the case: Wrong type")
	}
}

func TestMergeStatsEmptyStats(t *testing.T) {
	a := createAgentWithServices("", "", 0)
	_, err := a.mergeStats(nil)
	if err == nil {
		t.Errorf("mergeStats err is nil, but error expected for the case: Empty slice of stats")
	}
}

// createExampleStatsSlice is an util function that creates a fake slice of client.Stats for testing.
// the size of the slice can be set as a parameter, but the number of Upstreams, peers per upstream and zones are fixed
func createExampleStatsSlice(size uint64, unavailPeer bool) []*client.Stats {
	var stats []*client.Stats
	var i uint64
	for i = 0; i < size; i++ {
		peers := []client.Peer{
			{Active: i, State: peerUpState},
			{Active: i + 1, State: peerUpState},
		}
		state := peerUpState
		if unavailPeer {
			state = "down"
		}
		peers = append(peers, client.Peer{Active: i + 2, State: state})

		upstreams := map[string]client.Upstream{
			"service01": {Peers: peers},
			"service02": {Peers: peers},
		}
		serverZones := map[string]client.ServerZone{
			"zone1.org": {Processing: i},
			"zone2.org": {Processing: i + 1},
		}
		newStats := &client.Stats{
			Connections: client.Connections{Active: i},
			Upstreams:   upstreams,
			ServerZones: serverZones,
		}

		stats = append(stats, newStats)
	}
	return stats
}

// createAgentWithServices returns a new instance of Agent with only services configured.
func createAgentWithServices(method, sampling string, threshold uint) *Agent {
	return &Agent{
		services: Services{
			Method:       method,
			SamplingType: sampling,
			Threshold:    threshold,
			Feeds: []output.Feed{
				{Name: "svc1", FeedName: "feed1"},
				{Name: "svc2", FeedName: "feed2"},
			},
		},
	}
}
