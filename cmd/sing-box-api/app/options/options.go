package options

import (
	"fmt"
	"net"

	"github.com/spf13/pflag"
)

// Options contains everything necessary to create and run api.
type Options struct {
	BindAddress net.IP
	Port        int
}

// NewOptions returns initialized Options.
func NewOptions() *Options {
	return &Options{}
}

// Validate validates the options.
func (o *Options) Validate() error {
	if o == nil {
		return fmt.Errorf("options cannot be nil")
	}
	if o.Port <= 0 || o.Port > 65535 {
		return fmt.Errorf("invalid port: %d, must be between 1 and 65535", o.Port)
	}
	if o.BindAddress == nil {
		return fmt.Errorf("bind address cannot be nil")
	}
	return nil
}

// AddFlags adds flags of api to the specified FlagSet
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.IPVar(&o.BindAddress, "bind-address", net.IPv4(127, 0, 0, 1), "IP address on which to serve the --port, set to 0.0.0.0 for all interfaces")
	fs.IntVar(&o.Port, "port", 8001, "secure port to listen to for incoming HTTPS requests")
}
