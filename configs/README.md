# Agent Configuration

The agent needs to be configured in order to work. Configuration parameters are set using a YAML file that the agent will parse on start. Therefore, these parameters can't be changed in runtime, once the agent is running it will be using them until is restarted.

Non-required parameters with a default value will use the default if left blank.

## Agent
| Name | Definition | Default | Required |
|------|------------|:-------:|:--------:|
| interval | Time in seconds to perform a call to the NS1 API with new data | `60` | No |
| interval_max_random_delay | Max delay in seconds that will be used as a jitter in the main loop. For example, if `interval` is 60 and `interval_max_random_delay` is set to 20, the loop will last 60 seconds plus a random amount of seconds between 0 and 20. By default, no delay is added. | 0 | No |
| retry_time | Time in seconds to retry fetch/push of the data after an error | `5` | No |

**Note**: The `interval_max_random_delay` is used in order to add some jitter to the agent in the main loop. This is done in the case there are more than 1 instance
of the agent running, and to prevent all the agents sending data to the API at the same time.

## NGINX Plus

| Name | Definition | Default | Required |
|------|------------|:-------:|:--------:|
| hosts | List of 1 or more NGINX Plus instances. | - | Yes |
| api_endpoint | NGINX Plus API endpoint configured in all the instances | `/api` | No |
| client_timeout | The timeout in seconds for the NGINX Plus http client | `10` | No |
| resolver | Use a custom resolver to get the `hosts` addresses. The format is `ip:port`. This parameter is optional | - | No |
| resolver_timeout | The timeout in seconds for the lookup of the NGINX Plus hosts | `10` | No |

**Note:** If not resolver is configured, the local resolver will be used.

### Hosts
NGINX Hosts are defined using the following parameters

```yaml
  host: "127.0.0.1"
  port: 8443
  resolve: true
  host_header: "example.com"
```

* Host is the host of the NGINX Plus instance
* Port to use in order to connect to the Host. If no port defined `80` will be used
* Resolve. Whether to resolve the `host` using the resolver and get all the addresses resolved by lookup or use the host as it is
* Host Header is the `Host` http header that will be used when connecting to the host or resolved addresses. This parameter is not required.

## NSONE API

| Name | Definition | Default | Required |
|------|------------|:-------:|:--------:|
| api_key | The NS1 API Key | - | Yes |
| client_timeout | The timeout in seconds for the NS1 API http client | `10` | No |
| source_id | Datasource ID in NS1 Dashboard  | - | Yes |


## Services

The following parameters are used to create relations between upstream/zones and NS1 Feeds.

| Name | Definition | Default | Required |
|------|------------|:-------:|:--------:|
| method | Select the type of the agent and how it will fetch the metrics from NGINX Plus. Valid types are "global", "upstream_groups" or "status_zones" | - | Yes |
| threshold | **Note:** Only for `upstream_groups`. Minimum number of available peers per upstream to consider the NGINX Plus instance `up` | 0 | No |
| sampling_type | **Note:** Only for `upstream_groups`. How to merge the metrics from the peers. Only two values are valid: "count" or "avg" | "count" | No |
| feeds | List of feeds or PoP locations in NS1 Dashboard. Each feed requires both the name (in NGINX) and the feed name (except for `global` type, that only requires feed name) | - | Yes |

### Methods 

There are 3 different types of agent (methods). Only 1 type can be used at the same time. The method will determine how and what metrics are collected from NGINX Plus:
1. Global: Fetch global active connections from NGINX Plus, without any other filter.
2. Upstream Groups: Select from what upstreams collect the data from. Only defined upstreams will be fetched. This method has the 2 following extra settings:
     * Threshold. A number of peers greater or equal to the threshold must be available for the upstream to be considered up.
     * Sampling Type. By default "count" will sum all the active connections in the peers of the defined upstreams. If "avg" is set, the value will be divided by the number of available peers.
3. Status Zones: Select from what status zones collect the data from. Only defined zones will be fetched.

### Feeds
Feeds are the way to create a relation between upstream/zones and NS1 Feeds in a more controlled way. Depending on the chosen `method`. 

```yaml
services:
  feeds:
    - name: "my-service"
      feed_name: "region01"
    - name: "other-service"
      feed_name: "region02"
```

## Working examples of configuration

For more information check the following examples, depending on the type of agent:

* [Global connections configuration](example_global.yaml)
* [Upstream Groups configuration](example_upstreams.yaml)
* [Status Zones configuration](example_zones.yaml)
