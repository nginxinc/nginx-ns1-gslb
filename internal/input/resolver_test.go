package input

import (
	"testing"
)

func TestResolverLookup(t *testing.T) {
	// localhost with be resolved using the local resolver instead, so no dependency on 8.8.8.8:53
	resolver := NewResolver("8.8.8.8:53", 5)
	host := "localhost"
	addrs, err := resolver.Lookup(host)

	if len(addrs) == 0 {
		t.Errorf("resolver didn't return any address for host %v", host)
	}

	if err != nil {
		t.Errorf("error trying to resolve %v: %v", host, err)
	}
}

func TestResolverTimeout(t *testing.T) {
	resolver := NewResolver("8.8.8.8:53", 0)
	_, err := resolver.Lookup("")

	if err == nil {
		t.Errorf("NewResolver err returned nil but expected a timeout error")
	}
}

func TestResolverExternalWrongAddress(t *testing.T) {
	resolverAddress := "8.8.8.8"
	resolver := NewResolver(resolverAddress, 5)
	_, err := resolver.Lookup("nginx.org")
	if err == nil {
		t.Errorf("NewResolver err returned nil but expected an error because resolver address (%v) has no port", resolverAddress)
	}
}
