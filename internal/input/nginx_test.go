package input

import (
	"reflect"
	"testing"
)

func TestConstructFullEndpoint(t *testing.T) {
	expected := "http://nginxplus/api"
	endpoint := constructFullEndpoint("http://", "nginxplus", "/api")

	if !reflect.DeepEqual(expected, endpoint) {
		t.Errorf("constructFullEndpoint returned %v, but %v expected", endpoint, expected)
	}
}

func TestConfigureNginxPlus(t *testing.T) {
	testsCfgFail := []struct {
		input *Cfg
		msg   string
	}{
		{
			input: &Cfg{},
			msg:   "Empty list of Hosts",
		},
		{
			input: &Cfg{
				Hosts:         []NginxHost{{Host: "localhost"}},
				ClientTimeout: 1,
			},
			msg: "NGINX Plus not reachable",
		},
	}
	nginxPlus := NginxPlus{}
	for _, cfg := range testsCfgFail {
		err := nginxPlus.Configure(cfg.input)
		if err == nil {
			t.Errorf("NGINX Plus configuration err returned %+v but expected an error for case: %v", err, cfg.msg)
		}
	}
}
