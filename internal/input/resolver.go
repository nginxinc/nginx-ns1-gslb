package input

import (
	"context"
	"log"
	"net"
	"time"
)

// Resolver handles the resolution of NGINX Plus IPs using a custom DNS resolver
type Resolver struct {
	resolver *net.Resolver
	timeout  int
}

// NewResolver returns a new instance of the Resolver
func NewResolver(resolver string, timeout int) *Resolver {
	if resolver == "" {
		log.Print("Using the local resolver to resolve the hosts")
		return &Resolver{}
	}

	log.Printf("Using a custom resolver [%v] to resolve the hosts", resolver)

	return &Resolver{
		resolver: &net.Resolver{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", resolver)
			},
			PreferGo: true,
		},
		timeout: timeout,
	}
}

// Lookup returns a list of IP Addresses for a given host using the custom resolver. If not resolver defined, the local resolver is used.
func (r *Resolver) Lookup(host string) ([]string, error) {
	if r.resolver == nil {
		return localLookup(host)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()

	return r.resolver.LookupHost(ctx, host)
}

func localLookup(host string) ([]string, error) {
	return net.LookupHost(host)
}
