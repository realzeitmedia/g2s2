package g2s2_test

import (
	"fmt"
	"time"

	"github.com/realzeitmedia/g2s2"
)

func Example() {
	d, err := g2s2.DialUDP("1.2.3.4:5678")
	_ = err // ...
	defer d.Stop()

	sampleRate := 0.1

	// Same as PB's g2s
	d.Counter(sampleRate, "my.count", 1)
	d.Timing(sampleRate, "tick.tack", time.Second)
	d.Gauge(sampleRate, "pi", 3.14)

	// Or take the sample out
	if d.Sample(sampleRate) {
		d.CounterSmpl(sampleRate, fmt.Sprintf("my.%s.total", "count"), 1)
		d.TimingSmpl("tick.tack", time.Second)
		d.GaugeSmpl("pi", 3.14)
	}

}
