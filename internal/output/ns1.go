package output

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nginxinc/nginx-ns1-gslb/internal"
	api "gopkg.in/ns1/ns1-go.v2/rest"
)

// Feed contains all the information related one single Feed for the NS1 API call
type Feed struct {
	Name     string `yaml:"name"`
	FeedName string `yaml:"feed_name"`
}

// Cfg stores the configuration parameters for NS1
type Cfg struct {
	APIKey        string `yaml:"api_key"`
	ClientTimeout int    `yaml:"client_timeout"`
	SourceID      string `yaml:"source_id"`
}

// NS1 stores the NSONE API client and some internal configuration to send the data to NSONE
type NS1 struct {
	Cfg    *Cfg
	client *api.Client
}

// Push will send the data to the NSONE API
func (ns1 *NS1) Push(data map[string]*internal.FeedData) error {
	if len(data) == 0 {
		return fmt.Errorf("there is no data to send")
	}

	log.Printf("Pushing data to NS1")

	// _ is the http.Response object. We don't need it here as the API does not return anything meaningful
	_, err := ns1.client.DataSources.Publish(ns1.Cfg.SourceID, data)

	return err
}

// Configure sets the configuration for the NS1 client
func (ns1 *NS1) Configure(cfg *Cfg) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("The NS1 Pusher requires an API KEY to be defined")
	}

	if cfg.SourceID == "" {
		return fmt.Errorf("The NS1 Pusher requires a SourceID to be defined")
	}

	ns1.Cfg = cfg
	httpClient := &http.Client{Timeout: time.Duration(ns1.Cfg.ClientTimeout) * time.Second}
	ns1.client = api.NewClient(httpClient, api.SetAPIKey(ns1.Cfg.APIKey))
	return nil
}

// GetFeedsForSourceID returns a map with all the feed names as keys for future checks
func (ns1 NS1) GetFeedsForSourceID(sourceID string) (map[string]bool, error) {
	feeds, _, err := ns1.client.DataFeeds.List(sourceID)
	if err != nil {
		return nil, err
	}
	feedsMap := make(map[string]bool, len(feeds))
	for _, f := range feeds {
		feedsMap[f.Name] = true
	}

	return feedsMap, nil
}
