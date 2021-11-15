package input

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/nginxinc/nginx-plus-go-client/client"
	nginx "github.com/nginxinc/nginx-plus-go-client/client"
)

var httpProtocol = "http://"

// Cfg stores the configuration parameters for all the NGINX Plus instances to get the data from
type Cfg struct {
	Hosts           []NginxHost `yaml:"hosts"`
	ClientTimeout   int         `yaml:"client_timeout"`
	APIEndpoint     string      `yaml:"api_endpoint"`
	Resolver        string      `yaml:"resolver"`
	ResolverTimeout int         `yaml:"resolver_timeout"`
}

// NginxPlus stores the NGINX Plus API client and some internal configuration to fetch data from NGINX
type NginxPlus struct {
	Cfg         *Cfg
	ClientsPool []*nginx.NginxClient
}

// Task is a wrapper to store results of fetching multiple NGINX Plus instances
type Task struct {
	result *client.Stats
	err    error
}

// NginxHost stores the information about a remote host of an NGINX Plus instance
type NginxHost struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Resolve    bool   `yaml:"resolve"`
	HostHeader string `yaml:"host_header"`
}

func (nh NginxHost) String() string {
	h := nh.Host
	if nh.Port != 0 {
		h = fmt.Sprintf("%v:%d", nh.Host, nh.Port)
	}
	return fmt.Sprintf("%v (resolved: %v)", h, nh.Resolve)
}

func (n *NginxPlus) asyncFetchGlobalStats() []Task {
	var wg sync.WaitGroup
	finishedTasks := make([]Task, len(n.ClientsPool))

	for i, nginxClient := range n.ClientsPool {
		wg.Add(1)
		go func(index int, nginxClient *client.NginxClient) {
			defer wg.Done()
			result, err := nginxClient.GetStats()
			t := Task{
				result: result,
				err:    err,
			}
			finishedTasks[index] = t
		}(i, nginxClient)
	}
	wg.Wait()

	return finishedTasks
}

// Fetch gets the stats of n NGINX Plus instances
func (n *NginxPlus) Fetch() []*client.Stats {
	finishedTasks := n.asyncFetchGlobalStats()
	var statsSlice []*client.Stats
	for _, task := range finishedTasks {
		if task.err != nil {
			log.Printf("Error fetching from NGINX Plus instance: %v", task.err)
		} else {
			statsSlice = append(statsSlice, task.result)
		}
	}
	return statsSlice
}

// Configure sets the configuration of the NginxPlus clients
func (n *NginxPlus) Configure(cfg *Cfg) error {
	if len(cfg.Hosts) == 0 {
		return fmt.Errorf("The NGINX Plus Fetcher requires at least 1 host to be defined")
	}
	n.Cfg = cfg
	var resolvedHosts []NginxHost

	resolver := NewResolver(cfg.Resolver, cfg.ResolverTimeout)
	for _, nHost := range n.Cfg.Hosts {
		var addrs []string
		var err error
		if nHost.Resolve {
			addrs, err = resolver.Lookup(nHost.Host)
			if err != nil {
				return fmt.Errorf("error trying to resolve address for [%v]: %w", nHost.Host, err)
			}
		} else {
			addrs = append(addrs, nHost.Host)
		}

		for _, addr := range addrs {
			newHost := NginxHost{
				Host:       addr,
				Resolve:    nHost.Resolve,
				HostHeader: nHost.HostHeader,
				Port:       nHost.Port,
			}
			resolvedHosts = append(resolvedHosts, newHost)
		}
	}

	log.Printf("Creating clients for NGINX Plus hosts: %v", resolvedHosts)

	for _, nHost := range resolvedHosts {
		hostHeader := nHost.Host
		if nHost.Resolve {
			hostHeader = nHost.HostHeader
		}
		newHost := nHost.Host
		if nHost.Port != 0 {
			newHost = fmt.Sprintf("%v:%d", nHost.Host, nHost.Port)
		}
		httpClient := &http.Client{
			Timeout:   time.Duration(n.Cfg.ClientTimeout) * time.Second,
			Transport: newHostHeaderEnforcerTransport(hostHeader),
		}

		nginxClient, err := nginx.NewNginxClient(httpClient, constructFullEndpoint(httpProtocol, newHost, n.Cfg.APIEndpoint))
		if err != nil {
			return err
		}
		log.Printf("New NGINX Plus host configured: [%v] %v", hostHeader, newHost)
		n.ClientsPool = append(n.ClientsPool, nginxClient)
	}

	return nil
}

// hostHeaderEnforcerTransport is an implementation of http.Transport that will define a custom RoundTrip method.
type hostHeaderEnforcerTransport struct {
	http.Transport
	host string
}

func newHostHeaderEnforcerTransport(host string) *hostHeaderEnforcerTransport {
	return &hostHeaderEnforcerTransport{
		host: host,
	}
}

// RoundTrip overrides the RoundTrip method in the hostHeaderEnforcerTransport to update the Host header with the configured value by the user.
func (ct *hostHeaderEnforcerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Host = ct.host
	return ct.Transport.RoundTrip(req)
}

func constructFullEndpoint(protocol, host, endpoint string) string {
	return fmt.Sprintf("%s%s%s", protocol, host, endpoint)
}
