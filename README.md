# NGINX Plus - NS1 Global Server Load Balancing
[![Continuous Integration](https://github.com/nginxinc/nginx-ns1-gslb/actions/workflows/ci.yml/badge.svg)](https://github.com/nginxinc/nginx-ns1-gslb/actions/workflows/ci.yml)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-ns1-gslb.svg?type=shield)](https://app.fossa.com/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-ns1-gslb?ref=badge_shield)
[![Go Report Card](https://goreportcard.com/badge/github.com/nginxinc/nginx-ns1-gslb)](https://goreportcard.com/report/github.com/nginxinc/nginx-ns1-gslb)

**nginx-ns1-gslb** allows [NGINX Plus](https://www.nginx.com/products/nginx) to connect with NS1 API to use [NS1 managed DNS](https://ns1.com/products/managed-dns) to create a Global Server Load Balancing solution that load balances connections or requests across two or more distinct data centers or points of presence (PoPs).

## Prerequisites
* At least 1 working instance of NGINX Plus reachable from the host where you are running this agent
* NS1 API credentials

## Configuration
The configuration of the agent is managed by a YAML file. Check the [configuration](configs/README.md) readme for more information.

The agent will try to open the file and configure itself, there might be 2 problems when doing this:

* The file can't be opened (due to a bad path or file permissions)
* The file is opened but the data can't be used (errors or missing required parameters)

Both cases will make the agent fail to start with an error to describe the problem.

## Running the agent

### Locally (using Go >= 1.11)

`go run cmd/agent/main.go --config-file <path/to/your_file.yaml>`

This will run an Agent that fetches stats from one or more NGINX Plus instances, and updates the remote NS1 feeds.

### Docker

1. Build the image specifying the config file:
`make CONFIG_FILE=<path/to/your_file.yaml> container`

    **Note**: By default the binary is built locally, if you need to build it inside a container append `TARGET=container` to the command above.

1. Run the container:
`docker run nginx/nginx-ns1-gslb:<version>`

**Note:** consider the following while running the agent:

* All NGINX Plus instances must be reachable and running when the agent is run for the first time.
* While running, if at least 1 Instance of NGINX Plus is working, the agent will use the data from that one.
* While running, if all NGINX Plus instances are off, the agent will send `{up: false}` to NS1 API for all the configured services.

## Tests
Run `make test` to run the tests.
