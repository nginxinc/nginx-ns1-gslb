package agent

import (
	"fmt"
	"testing"

	"github.com/nginxinc/nginx-ns1-gslb/internal/output"
)

func TestValidateServicesCfg(t *testing.T) {
	testCases := []struct {
		cfg     *Config
		wantErr bool
		msg     string
	}{
		{
			cfg: &Config{
				Services: Services{},
			},
			wantErr: true,
			msg:     "empty list of Feeds",
		},
		{
			cfg: &Config{
				Services: Services{
					Feeds: []output.Feed{
						{Name: "svc1", FeedName: ""},
						{Name: "svc1", FeedName: ""},
					},
					SamplingType: "count",
				},
			},
			wantErr: true,
			msg:     "duplicated feed resource name",
		},
		{
			cfg: &Config{
				Services: Services{
					Feeds: []output.Feed{
						{Name: "", FeedName: ""},
					},
					SamplingType: "count",
				},
			},
			wantErr: true,
			msg:     "feed missing name and feedname",
		},
		{
			cfg: &Config{
				Services: Services{
					Feeds: []output.Feed{
						{FeedName: "feed01"},
					},
					SamplingType: "count",
				},
			},
			wantErr: true,
			msg:     "feed missing name",
		},
		{
			cfg: &Config{
				Services: Services{
					Feeds: []output.Feed{
						{Name: "svc01"},
					},
					SamplingType: "count",
				},
			},
			wantErr: true,
			msg:     "feed missing feed_name",
		},
		{
			cfg: &Config{
				Services: Services{
					Method: globalMethod,
					Feeds: []output.Feed{
						{FeedName: "feed01"},
						{FeedName: "feed02"},
					},
					SamplingType: "count",
				},
			},
			wantErr: false,
			msg:     fmt.Sprintf("method [%v] does not require service name checks", globalMethod),
		},
		{
			cfg: &Config{
				Services: Services{
					Feeds: []output.Feed{
						{Name: "svc1", FeedName: "feed01"},
						{Name: "svc2", FeedName: "feed02"},
					},
					SamplingType: "count",
				},
			},
			wantErr: false,
			msg:     "valid input",
		},
		{
			cfg: &Config{
				Services: Services{
					Feeds: []output.Feed{
						{Name: "svc1", FeedName: "feed01"},
					},
					SamplingType: "testSampling",
				},
			},
			wantErr: true,
			msg:     "wrong sampling type",
		},
	}

	for _, testCase := range testCases {
		err := validateServicesCfg(testCase.cfg)
		if err == nil && testCase.wantErr {
			t.Errorf("validateServicesCfg err returned <nil>, but err expected an error for case %v", testCase.msg)
		}
		if err != nil && !testCase.wantErr {
			t.Errorf("validateServicesCfg returned an err: %v for case %v", err, testCase.msg)
		}
	}
}

func TestParseExampleConfigs(t *testing.T) {
	testCases := []struct {
		path    string
		msg     string
		wantErr bool
	}{
		{
			path:    "../../configs/example_global.yaml",
			msg:     fmt.Sprintf("valid configuration for %v method", globalMethod),
			wantErr: false,
		},
		{
			path:    "../../configs/example_upstreams.yaml",
			msg:     fmt.Sprintf("valid configuration for %v method", upstreamGroupsMethod),
			wantErr: false,
		},
		{
			path:    "../../configs/example_zones.yaml",
			msg:     fmt.Sprintf("valid configuration for %v method", statusZonesMethod),
			wantErr: false,
		},
		{
			msg:     "No path to config file",
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		_, err := ParseConfig(&testCase.path)
		if err == nil && testCase.wantErr {
			t.Errorf("ParseConfig err returned <nil>, but err expected an error for case %v", testCase.msg)
		}
		if err != nil && !testCase.wantErr {
			t.Errorf("ParseConfig returned an err: %v for case %v", err, testCase.msg)
		}
	}
}
