Simple Go statsd client.

Key features:
 - Can do the sampling before building the measurement key, which can save
   Sprintf()s.
 - Fills UDP packets before sending them.


See example_test.go for usage.

Installation:

    `go get github.com/realzeitmedia/g2s`

