package g2s2

import (
	"bytes"
	"io"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	// MaxPacketUDP is the default payload size when DialUDP() is used.
	MaxPacketUDP = 1432 // Fast Ethernet minus overhead
)

// T is returned by New().
type T struct {
	wg      sync.WaitGroup
	w       io.Writer
	p       chan []byte
	maxSize int
}

// DialUDP connects to addr and gives the connection to New(). Stop() the
// returned value when done. It uses MaxPacketUDP for the maximum packet payload.
func DialUDP(addr string) (*T, error) {
	c, err := net.DialTimeout("udp", addr, time.Second)
	if err != nil {
		return nil, err
	}
	return New(c, MaxPacketUDP), nil
}

// New starts a Go routine writing full packets to w.
func New(w io.Writer, packetSize int) *T {
	t := T{
		wg:      sync.WaitGroup{},
		w:       w,
		p:       make(chan []byte),
		maxSize: packetSize,
	}
	t.wg.Add(1)
	go func() {
		t.process()
		t.wg.Done()
	}()
	return &t
}

// Stop the go routine and flushes the last packets. Don't use any method
// anymore after calling Stop.
func (t *T) Stop() {
	close(t.p)
	t.wg.Wait()
}

func (t *T) process() {
	b := bytes.Buffer{}
	for msg := range t.p {
		if b.Len() > 0 {
			if b.Len()+1+len(msg) >= t.maxSize {
				// no room for this msg.
				if _, err := t.w.Write(b.Bytes()); err != nil {
					// ... ?
				}
				b.Reset()
			} else {
				b.WriteRune('\n')
			}
		}
		b.Write(msg)
	}

	if b.Len() > 0 {
		if _, err := t.w.Write(b.Bytes()); err != nil {
			// ... ?
		}
	}
}

// Sample is used with the *Smpl methods to only build keys when the
// measurement will be sampled. See the example for usage.
func (t *T) Sample(rate float64) bool {
	return rate > rand.Float64()
}

// CounterSmpl is a statsd counter, after sampling by Sample().
func (t *T) CounterSmpl(rate float64, k string, v int64) {
	b := make([]byte, 0, len(k)+20)
	b = append(b, []byte(k)...)
	b = append(b, ':')
	b = strconv.AppendInt(b, v, 10)
	b = append(b, []byte("|c|@")...)
	b = strconv.AppendFloat(b, rate, 'f', 4, 64)
	t.p <- b
}

// Counter counts k.
func (t *T) Counter(rate float64, k string, v int64) {
	if t.Sample(rate) {
		t.CounterSmpl(rate, k, v)
	}
}

// TimingSmpl is a statsd timer, after sampling by Sample().
func (t *T) TimingSmpl(k string, v time.Duration) {
	b := make([]byte, 0, len(k)+10)
	b = append(b, []byte(k)...)
	b = append(b, ':')
	b = strconv.AppendInt(b, v.Nanoseconds()/1e6, 10)
	b = append(b, '|', 'm', 's')
	t.p <- b
}

// Timing times k
func (t *T) Timing(rate float64, k string, v time.Duration) {
	if t.Sample(rate) {
		t.TimingSmpl(k, v)
	}
}

// GaugeSmpl is a statsd gauge, after sampling by Sample().
func (t *T) GaugeSmpl(k string, v float64) {
	b := make([]byte, 0, len(k)+20)
	b = append(b, []byte(k)...)
	b = append(b, ':')
	b = strconv.AppendFloat(b, v, 'f', 6, 64)
	b = append(b, '|', 'g')
	t.p <- b
}

// Gauge gauges k
func (t *T) Gauge(rate float64, k string, v float64) {
	if t.Sample(rate) {
		t.GaugeSmpl(k, v)
	}
}
