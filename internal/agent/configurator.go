package agent

import (
	"fmt"
	"io/ioutil"

	"github.com/nginxinc/nginx-ns1-gslb/internal/input"
	"github.com/nginxinc/nginx-ns1-gslb/internal/output"
	yaml "gopkg.in/yaml.v2"
)

// Services stores the configuration that relates NGINX Plus services with NS1 Data Feeds
type Services struct {
	Method       string        `yaml:"method"`
	Threshold    uint          `yaml:"threshold"`
	SamplingType string        `yaml:"sampling_type"`
	Feeds        []output.Feed `yaml:"feeds"`
}

// Config stores all the parameters from the configuration file
type Config struct {
	Agent     Cfg        `yaml:"agent"`
	NginxPlus input.Cfg  `yaml:"nginx_plus"`
	Nsone     output.Cfg `yaml:"nsone"`
	Services  Services   `yaml:"services"`
}

// ParseConfig reads the configuration file and return a Config object ready to configure agent and resources
func ParseConfig(path *string) (*Config, error) {
	data, err := ioutil.ReadFile(*path)
	if err != nil {
		return nil, fmt.Errorf("error reading file at %v: %w", path, err)
	}

	globalConfig := &Config{}
	err = yaml.Unmarshal(data, globalConfig)
	if err != nil {
		return nil, fmt.Errorf("error while parsing the configuration file: %w", err)
	}

	globalConfig = fillWithDefaults(globalConfig)

	err = validateServicesCfg(globalConfig)
	if err != nil {
		return nil, fmt.Errorf("error while validating Services configuration: %w", err)
	}

	return globalConfig, nil
}

func validateServicesCfg(cfg *Config) error {
	if len(cfg.Services.Feeds) == 0 {
		return fmt.Errorf("at least 1 Feed needs to be defined")
	}

	if cfg.Services.SamplingType != mergeAvg && cfg.Services.SamplingType != mergeCount {
		return fmt.Errorf("sampling Type [%v] is not a valid type. Valid Sampling Types are: %v, %v", cfg.Services.SamplingType, mergeAvg, mergeCount)
	}

	names := make(map[string]bool)
	for _, feed := range cfg.Services.Feeds {
		if feed.FeedName == "" {
			return fmt.Errorf("feeds must define at least a feed_name")
		}

		if cfg.Services.Method != globalMethod {
			if feed.Name == "" {
				return fmt.Errorf("feeds must define a name for method: %v", cfg.Services.Method)
			}

			if _, ok := names[feed.Name]; ok {
				return fmt.Errorf("[%v] duplicated in Feed List. NGINX resources names must be unique", feed.Name)
			}
			names[feed.Name] = true
		}
	}

	return nil
}

func fillWithDefaults(cfg *Config) *Config {
	if cfg.Agent.Interval == 0 {
		cfg.Agent.Interval = 60
	}

	if cfg.Agent.Interval == 0 {
		cfg.Agent.Interval = 5
	}

	if cfg.NginxPlus.ClientTimeout == 0 {
		cfg.NginxPlus.ClientTimeout = 10
	}

	if cfg.NginxPlus.ResolverTimeout == 0 {
		cfg.NginxPlus.ResolverTimeout = 10
	}

	if cfg.NginxPlus.APIEndpoint == "" {
		cfg.NginxPlus.APIEndpoint = "/api"
	}

	if cfg.Nsone.ClientTimeout == 0 {
		cfg.Nsone.ClientTimeout = 10
	}

	if cfg.Services.SamplingType == "" {
		cfg.Services.SamplingType = mergeCount
	}

	return cfg
}
