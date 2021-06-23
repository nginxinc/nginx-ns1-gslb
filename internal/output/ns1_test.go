package output

import (
	"testing"
)

func TestConfigureNs1(t *testing.T) {
	testsCfgFail := []struct {
		input *Cfg
		msg   string
	}{
		{
			input: &Cfg{},
			msg:   "Missing API KEY",
		},
		{
			input: &Cfg{
				APIKey: "apikey",
			},
			msg: "Missing SourceID",
		},
	}
	ns1 := NS1{}
	for _, cfg := range testsCfgFail {
		err := ns1.Configure(cfg.input)
		if err == nil {
			t.Errorf("NS1 configuration err returned %+v but expected an error for case: %v", err, cfg.msg)
		}
	}
}
