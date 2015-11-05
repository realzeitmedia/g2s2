package g2s

import (
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

type chunkBuf struct {
	msg []string
}

func (c *chunkBuf) Write(msg []byte) (int, error) {
	c.msg = append(c.msg, string(msg))
	return len(msg), nil
}

const (
	s10  = "1234567890"
	s25  = s10 + s10 + "12345"
	s50  = s10 + s10 + s10 + s10 + s10
	s100 = s50 + s50
)

func Test(t *testing.T) {
	type C struct {
		name string
		cmds func(*T)
		want []string
	}
	for _, cas := range []C{
		{
			name: "basic counter",
			cmds: func(g *T) {
				g.Counter(1.0, "key", 42)
			},
			want: []string{
				"key:42|c|@1.0000",
			},
		},
		{
			name: "triple counter",
			cmds: func(g *T) {
				g.Counter(1.0, "key", 42)
				g.Counter(1.0, "key", 43)
				g.Counter(1.0, "key", 44)
			},
			want: []string{
				"key:42|c|@1.0000\n" +
					"key:43|c|@1.0000\n" +
					"key:44|c|@1.0000",
			},
		},
		{
			name: "big counter",
			cmds: func(g *T) {
				g.Counter(1.0, s25, 42)
				g.Counter(1.0, s25, 43)
				g.Counter(1.0, s25, 44)
				g.Counter(1.0, s25, 45) // doesn't fit
			},
			want: []string{
				s25 + ":42|c|@1.0000\n" +
					s25 + ":43|c|@1.0000",
				s25 + ":44|c|@1.0000\n" +
					s25 + ":45|c|@1.0000",
			},
		},

		{
			name: "basic timer",
			cmds: func(g *T) {
				g.Timing(1.0, "key", 23*time.Millisecond)
			},
			want: []string{
				"key:23|ms",
			},
		},

		{
			name: "basic gauge",
			cmds: func(g *T) {
				g.Gauge(1.0, "key", 42.0)
			},
			want: []string{
				"key:42.000000|g",
			},
		},

		{
			name: "big packet",
			cmds: func(g *T) {
				g.Counter(1.0, s100, 42)
			},
			want: []string{
				s100 + ":42|c|@1.0000",
			},
		},

		{
			name: "just fits",
			cmds: func(g *T) {
				g.Counter(1.0, s50, 42)
				g.Counter(1.0, "123123123123123123", 42)
				g.Counter(1.0, "1", 1)
			},
			want: []string{
				s50 + ":42|c|@1.0000\n" +
					"123123123123123123:42|c|@1.0000",
				"1:1|c|@1.0000",
			},
		},

		{
			name: "just not fits",
			cmds: func(g *T) {
				g.Counter(1.0, s50, 42)
				g.Counter(1.0, "12312312312312312312312", 42)
				g.Counter(1.0, "1", 1)
			},
			want: []string{
				s50 + ":42|c|@1.0000",
				"12312312312312312312312:42|c|@1.0000\n" +
					"1:1|c|@1.0000",
			},
		},
	} {
		b := &chunkBuf{}
		g := New(b, 100)
		cas.cmds(g)
		g.Stop()

		t.Logf("%s: have %#v", cas.name, b.msg)
		if have, want := b.msg, cas.want; !reflect.DeepEqual(have, want) {
			t.Errorf("%s: have %#v, want %#v", cas.name, have, want)
		}
	}
}

func BenchmarkCounter(b *testing.B) {
	g := New(ioutil.Discard, 100)
	defer g.Stop()
	for i := 0; i < b.N; i++ {
		g.Counter(1.0, s10, 42)
	}
}
