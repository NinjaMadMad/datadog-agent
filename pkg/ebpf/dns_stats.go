package ebpf

import (
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/pkg/process/util"
)

type dnsStats struct {
	// More stats like latency, error, etc. will be added here later
	successfulResponses uint32
	failedResponses     uint32
	successLatency      uint64 // Stored in µs
	failureLatency      uint64
	timeouts            uint32
}

type dnsKey struct {
	serverIP   util.Address
	clientIP   util.Address
	clientPort uint16
	// ConnectionType will be either TCP or UDP
	protocol ConnectionType
}

// DNSPacketType tells us whether the packet is a query or a reply (successful/failed)
type DNSPacketType uint8

const (
	// SuccessfulResponse indicates that the response code of the DNS reply is 0
	SuccessfulResponse DNSPacketType = iota
	// FailedResponse indicates that the response code of the DNS reply is anything other than 0
	FailedResponse
	// Query
	Query
)

type dnsPacketInfo struct {
	transactionID uint16
	key           dnsKey
	pktType       DNSPacketType
}

type stateKey struct {
	key dnsKey
	id  uint16
}

type dnsStatKeeper struct {
	mux              sync.Mutex
	stats            map[dnsKey]dnsStats
	state            map[stateKey]time.Time
	expirationPeriod time.Duration
	exit             chan struct{}
	maxSize          int // maximum size of the state map
	deleteCount      int
	deleteThreshold  int
}

func newDNSStatkeeper(timeout time.Duration) *dnsStatKeeper {
	statsKeeper := &dnsStatKeeper{
		stats:            make(map[dnsKey]dnsStats),
		state:            make(map[stateKey]time.Time),
		expirationPeriod: timeout,
		exit:             make(chan struct{}),
		maxSize:          10000,
		deleteThreshold:  5000,
	}

	ticker := time.NewTicker(statsKeeper.expirationPeriod)
	go func() {
		for {
			select {
			case now := <-ticker.C:
				statsKeeper.removeExpiredStates(now.Add(-statsKeeper.expirationPeriod))
			case <-statsKeeper.exit:
				ticker.Stop()
				return
			}
		}
	}()
	return statsKeeper
}

func (d *dnsStatKeeper) ProcessPacketInfo(info dnsPacketInfo, ts time.Time) {
	d.mux.Lock()
	defer d.mux.Unlock()
	sk := stateKey{key: info.key, id: info.transactionID}

	if info.pktType == Query {
		if len(d.state) == d.maxSize {
			return
		}

		if _, ok := d.state[sk]; !ok {
			d.state[sk] = ts
		}
		return
	}

	// If a response does not have a corresponding query entry, we discard it
	start, ok := d.state[sk]

	if !ok {
		return
	}

	delete(d.state, sk)
	d.deleteCount++

	latency := ts.Sub(start).Nanoseconds()

	stats := d.stats[info.key]

	if latency > d.expirationPeriod.Nanoseconds() {
		stats.timeouts++
	} else {
		latency /= 1000 // convert to microseconds
		if info.pktType == SuccessfulResponse {
			stats.successfulResponses++
			stats.successLatency += uint64(latency)
		} else if info.pktType == FailedResponse {
			stats.failedResponses++
			stats.failureLatency += uint64(latency)
		}
	}

	d.stats[info.key] = stats
}

func (d *dnsStatKeeper) GetAndResetAllStats() map[dnsKey]dnsStats {
	d.mux.Lock()
	defer d.mux.Unlock()
	ret := d.stats
	d.stats = make(map[dnsKey]dnsStats)
	return ret
}

func (d *dnsStatKeeper) removeExpiredStates(earliestTs time.Time) {
	d.mux.Lock()
	defer d.mux.Unlock()
	for k, v := range d.state {
		if v.Before(earliestTs) {
			delete(d.state, k)
			d.deleteCount++
			stats := d.stats[k.key]
			stats.timeouts++
			d.stats[k.key] = stats
		}
	}

	if d.deleteCount < d.deleteThreshold {
		return
	}

	// golang/go#20135 : maps do not shrink after elements removal (delete)
	copied := make(map[stateKey]time.Time, len(d.state))
	for k, v := range d.state {
		copied[k] = v
	}
	d.state = copied
	d.deleteCount = 0
}

func (d *dnsStatKeeper) Close() {
	d.exit <- struct{}{}
}
