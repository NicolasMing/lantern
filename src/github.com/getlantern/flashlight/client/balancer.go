package client

import (
	"time"

	"github.com/getlantern/balancer"
)

// getBalancer waits for a message from client.balCh to arrive and then it
// writes it back to client.balCh before returning it as a value. This way we
// always have a balancer at client.balCh and, if we don't have one, it would
// block until one arrives.
func (client *Client) getBalancer() *balancer.Balancer {
	bal, ok := client.bal.Get(24 * time.Hour)
	if !ok {
		panic("No balancer!")
	}
	return bal.(*balancer.Balancer)
}

// initBalancer takes hosts from cfg.ChainedServers and it uses them to create a
// balancer.
func (client *Client) initBalancer(cfg *ClientConfig) {
	if len(cfg.ChainedServers) == 0 {
		log.Debug("No chained servers configured, not initializing balancer")
		return
	}
	// The dialers slice must be large enough to handle all chained
	// servers.
	dialers := make([]*balancer.Dialer, 0, len(cfg.ChainedServers))

	// Add chained (CONNECT proxy) servers.
	log.Debugf("Adding %d chained servers", len(cfg.ChainedServers))
	if len(cfg.ChainedServers) == 0 {
		log.Error("NO CHAINED SERVERS!")
	}
	for _, s := range cfg.ChainedServers {
		dialer, err := s.Dialer(cfg.DeviceID)
		if err == nil {
			dialers = append(dialers, dialer)
		} else {
			log.Errorf("Unable to configure chained server. Received error: %v", err)
		}
	}

	bal := balancer.New(dialers...)
	oldBal, ok := client.bal.Get(25 * time.Millisecond)
	if ok {
		// Close old balancer on a goroutine to avoid blocking here
		go func() {
			oldBal.(*balancer.Balancer).Close()
			log.Debug("Closed old balancer")
		}()
	}

	log.Trace("Publishing balancer")
	client.bal.Set(bal)
}
