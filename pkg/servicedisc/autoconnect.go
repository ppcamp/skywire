package servicedisc

import (
	"context"
	"time"

	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/skycoin/skywire/internal/netutil"
	"github.com/skycoin/skywire/pkg/transport"
	"github.com/skycoin/skywire/pkg/transport/network"
)

const (
	// PublicServiceDelay defines a delay before adding transports to public services.
	PublicServiceDelay = 10 * time.Second

	fetchServicesDelay           = 10 * time.Second
	maxFailedAddressRetryAttempt = 2
)

// ConnectFn provides a way to connect to remote service
type ConnectFn func(context.Context, cipher.PubKey) error

// Autoconnector continuously tries to connect to services
type Autoconnector interface {
	Run(context.Context) error
}

type autoconnector struct {
	client   *HTTPClient
	maxConns int
	log      *logging.Logger
	tm       *transport.Manager
}

// MakeConnector returns a new connector that will try to connect to at most maxConns
// services
func MakeConnector(conf Config, maxConns int, tm *transport.Manager, log *logging.Logger) Autoconnector {
	connector := &autoconnector{}
	connector.client = NewClient(log, conf)
	connector.maxConns = maxConns
	connector.log = log
	connector.tm = tm
	return connector
}

// Run implements Autoconnector interface
func (a *autoconnector) Run(ctx context.Context) (err error) {
	// failed addresses will be populated everytime any failed attempt at establishing transport occurs.
	failedAddresses := map[cipher.PubKey]int{}

	for {
		time.Sleep(PublicServiceDelay)

		// successfully established transports
		tps := a.tm.GetTransportsByLabel(transport.LabelAutomatic)

		// don't fetch public addresses if there are more or equal to the number of maximum transport defined.
		if len(tps) >= a.maxConns {
			a.log.Debugln("autoconnect: maximum number of established transports reached: ", a.maxConns)
			return err
		}

		a.log.Infoln("Fetching public visors")
		addrs, err := a.fetchPubAddresses(ctx)
		if err != nil {
			a.log.Errorf("Cannot fetch public services: %s", err)
		}

		// filter out any established transports
		absent := a.filterDuplicates(addrs, tps)

		for _, pk := range absent {
			val, ok := failedAddresses[pk]
			if !ok || val < maxFailedAddressRetryAttempt {
				a.log.WithField("pk", pk).WithField("attempt", val).Debugln("Trying to add transport to public visor")
				logger := a.log.WithField("pk", pk).WithField("type", string(network.STCPR))
				if _, err := a.tm.SaveTransport(ctx, pk, network.STCPR, transport.LabelAutomatic); err != nil {
					logger.WithError(err).Warnln("Failed to add transport to public visor")
					failedAddresses[pk]++
					continue
				}
				logger.Infoln("Added transport to public visor")
			}
		}
	}
}

func (a *autoconnector) fetchPubAddresses(ctx context.Context) ([]cipher.PubKey, error) {
	retrier := netutil.NewRetrier(fetchServicesDelay, 5, 3, a.log)
	var services []Service
	fetch := func() (err error) {
		// "return" services up from the closure
		services, err = a.client.Services(ctx, a.maxConns)
		if err != nil {
			return err
		}
		return nil
	}
	if err := retrier.Do(fetch); err != nil {
		return nil, err
	}
	pks := make([]cipher.PubKey, len(services))
	for i, service := range services {
		pks[i] = service.Addr.PubKey()
	}
	return pks, nil
}

// return public keys from pks that are absent in given list of transports
func (a *autoconnector) filterDuplicates(pks []cipher.PubKey, trs []*transport.ManagedTransport) []cipher.PubKey {
	var absent []cipher.PubKey
	for _, pk := range pks {
		found := false
		for _, tr := range trs {
			if tr.Entry.HasEdge(pk) {
				found = true
				break
			}
		}
		if !found {
			absent = append(absent, pk)
		}
	}
	return absent
}
