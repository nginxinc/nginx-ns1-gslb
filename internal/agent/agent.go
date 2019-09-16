package agent

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nginxinc/nginx-ns1-gslb/internal"
	"github.com/nginxinc/nginx-ns1-gslb/internal/input"
	"github.com/nginxinc/nginx-ns1-gslb/internal/output"
	"github.com/nginxinc/nginx-plus-go-sdk/client"
)

const (
	globalMethod         = "global"
	upstreamGroupsMethod = "upstream_groups"
	statusZonesMethod    = "status_zones"
	peerUpState          = "up"
	mergeAvg             = "avg"
	mergeCount           = "count"
)

// Agent handles all the configuration, I/O and processing of the application
type Agent struct {
	fetcher       *input.NginxPlus
	pusher        *output.NS1
	cfg           *Cfg
	services      Services
	namedServices map[string]string
}

// Cfg stores the configuration parameters for the agent
type Cfg struct {
	Interval               uint32 `yaml:"interval"`
	IntervalMaxRandomDelay uint32 `yaml:"interval_max_random_delay"`
	RetryTime              uint32 `yaml:"retry_time"`
}

// configureAll will call configure() methods of fetcher, agent and pusher
func (agent *Agent) configureAll(fetcherCfg *input.Cfg, pusherCfg *output.Cfg) error {
	err := agent.fetcher.Configure(fetcherCfg)
	if err != nil {
		return fmt.Errorf("fetcher configuration error: %v", err)
	}

	err = agent.pusher.Configure(pusherCfg)
	if err != nil {
		return fmt.Errorf("pusher configuration error: %v", err)
	}

	err = agent.configure()
	if err != nil {
		return fmt.Errorf("agent configuration error: %v", err)
	}

	return nil
}

func (agent *Agent) configure() error {
	feedNames, err := agent.pusher.GetFeedsForSourceID(agent.pusher.Cfg.SourceID)
	if err != nil {
		return fmt.Errorf("error trying to get Feeds from NS1 for validation: %v", err)
	}

	agent.namedServices = make(map[string]string)
	for _, svc := range agent.services.Feeds {
		if _, ok := feedNames[svc.FeedName]; !ok {
			return fmt.Errorf("Feed Name %v not found in NS1 DataFeed with source = %v. Review NS1 configuration", svc.FeedName, agent.pusher.Cfg.SourceID)
		}
		if agent.services.Method == globalMethod {
			agent.namedServices[svc.FeedName] = svc.FeedName
		} else {
			agent.namedServices[svc.Name] = svc.FeedName
		}
	}

	return nil
}

func (agent *Agent) processData(statsSlice []*client.Stats) (map[string]*internal.FeedData, error) {
	newData := make(map[string]*internal.FeedData)
	if statsSlice != nil {
		// If we have data to merge
		inputData, err := agent.mergeStats(statsSlice)
		if err != nil {
			return nil, err
		}

		for src, feed := range agent.namedServices {
			// For the type Global we replicate the same information for all the feeds.
			var feedData *internal.FeedData
			if agent.services.Method == globalMethod {
				feedData = inputData[globalMethod]
			} else {
				feedData = inputData[src]
			}

			if feedData == nil {
				log.Printf("Error: [%v] source was not found in the remote NGINX Plus instance(s). Check NGINX Plus config file or agent config file.", src)
				continue
			}
			newData[feed] = feedData
		}
	} else {
		// If we don't have data to merge (eg: all NGINX Plus instances are offline)
		for _, feed := range agent.namedServices {
			newData[feed] = &internal.FeedData{
				Up: false,
			}
		}
	}
	return newData, nil
}

func (agent *Agent) handleErrorAndSleep(err error) {
	log.Printf("Error while running the main loop: %v. No data will be sent this time, will try again in %d seconds", err, agent.cfg.RetryTime)
	time.Sleep(time.Duration(agent.cfg.RetryTime) * time.Second)
}

// Run runs the main loop of the agent forever
func (agent *Agent) Run() {
	for {
		input := agent.fetcher.Fetch()
		if input == nil {
			fmt.Printf("None of the NGINX Plus instances were available.")
		}

		data, err := agent.processData(input)
		if err != nil {
			agent.handleErrorAndSleep(err)
			continue
		}

		err = agent.pusher.Push(data)
		if err != nil {
			log.Printf("Error pushing the data: %v", err)
		}

		sleepTime := int(agent.cfg.Interval)
		if agent.cfg.IntervalMaxRandomDelay > 0 {
			sleepTime += rand.Intn(int(agent.cfg.IntervalMaxRandomDelay))
		}
		log.Printf("Loop execution end, sleeping for %v seconds.", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}

// New creates and configures a new Agent (including both, the fetcher and the pusher)
func New(globalConfig *Config) (*Agent, error) {
	agent := Agent{
		cfg:      &globalConfig.Agent,
		fetcher:  &input.NginxPlus{},
		pusher:   &output.NS1{},
		services: globalConfig.Services,
	}
	err := agent.configureAll(&globalConfig.NginxPlus, &globalConfig.Nsone)
	return &agent, err
}

// merge an array of Stats fetched from one or more NGINX Plus instances focusing on the right stats depending on the configured methods
func (agent *Agent) mergeStats(statsSlice []*client.Stats) (map[string]*internal.FeedData, error) {
	if len(statsSlice) == 0 {
		return nil, fmt.Errorf("Error merging data: no data to merge, empty response")
	}

	switch agent.services.Method {
	case globalMethod:
		return getGlobalConnectionsData(statsSlice), nil
	case upstreamGroupsMethod:
		return getUpstreamConnectionsData(statsSlice, agent.services.SamplingType, agent.namedServices, int(agent.services.Threshold)), nil
	case statusZonesMethod:
		return getStatusZonesConnectionsData(statsSlice, agent.namedServices), nil
	}

	return nil, fmt.Errorf("Error processing the data from NGINX Plus instance(s): %v is not a valid NGINX Plus type", agent.services.Method)
}

func getGlobalConnectionsData(statsSlice []*client.Stats) map[string]*internal.FeedData {
	data := make(map[string]*internal.FeedData)
	feedData := &internal.FeedData{
		Up: true,
	}
	for _, s := range statsSlice {
		feedData.Connections += s.Connections.Active
	}
	data[globalMethod] = feedData
	return data
}

// UpstreamsConnections wraps Active connections and the number of available peers for a given Upstream Server
type UpstreamsConnections struct {
	Active         uint64
	AvailablePeers int
}

func getUpstreamConnectionsData(statsSlice []*client.Stats, mergeMethod string, namedServices map[string]string, peerThreshold int) map[string]*internal.FeedData {
	data := make(map[string]*internal.FeedData)
	upstreamConnections := make(map[string]*UpstreamsConnections)

	for _, s := range statsSlice {
		for key, ups := range s.Upstreams {
			if _, ok := namedServices[key]; !ok {
				continue
			}
			uc := &UpstreamsConnections{}

			for _, p := range ups.Peers {
				if p.State == peerUpState {
					uc.Active += p.Active
					uc.AvailablePeers++
				}
			}

			upstreamConnections[key] = uc
		}
	}

	for svc, uc := range upstreamConnections {
		feedData := &internal.FeedData{}

		if uc.AvailablePeers >= peerThreshold {
			feedData.Up = true
		}

		feedData.Connections = uc.Active
		if mergeMethod == mergeAvg {
			if uc.AvailablePeers > 0 {
				feedData.Connections = feedData.Connections / uint64(uc.AvailablePeers)
			}
		}

		data[svc] = feedData
	}
	return data
}

func getStatusZonesConnectionsData(statsSlice []*client.Stats, namedServices map[string]string) map[string]*internal.FeedData {
	data := make(map[string]*internal.FeedData)

	for _, s := range statsSlice {
		for svc, zone := range s.ServerZones {
			if _, ok := namedServices[svc]; !ok {
				continue
			}
			if _, ok := data[svc]; ok {
				data[svc].Connections += zone.Processing
			} else {
				data[svc] = &internal.FeedData{
					Connections: zone.Processing,
					Up:          true,
				}
			}
		}
	}
	return data
}
