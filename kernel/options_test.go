package kernel

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func timeFunc(ts time.Time) func() time.Time {
	return func() time.Time {
		return ts
	}
}

func TestSetTime(t *testing.T) {
	for _, test := range []struct {
		desc    string
		input   string
		options Options
		time    time.Time
		err     error
	}{
		{
			desc:  "timestamp",
			input: "1970-01-01T00:00:00Z",
			options: Options{
				NowFunc: timeFunc(time.Unix(0, 0).UTC()),
			},
			time: time.Unix(0, 0).UTC(),
		},
		{
			desc:  "now",
			input: "now",
			options: Options{
				NowFunc: timeFunc(time.Unix(0, 0).UTC()),
			},
			time: time.Unix(0, 0).UTC(),
		},
		{
			desc:    "invalid",
			input:   "invalid",
			options: Options{},
			err:     errors.New("not a valid timestamp: invalid"),
		},
		{
			desc:  "add to start",
			input: "start+24h",
			options: Options{
				TimeStart: time.Unix(0, 0).UTC(),
				NowFunc:   time.Now,
			},
			time: time.Date(1970, 01, 02, 0, 0, 0, 0, time.UTC),
		},
		{
			desc:  "subtract from end",
			input: "end-12h30m",
			options: Options{
				TimeEnd: time.Date(1970, 01, 02, 0, 0, 0, 0, time.UTC),
				NowFunc: time.Now,
			},
			time: time.Date(1970, 01, 01, 11, 30, 0, 0, time.UTC),
		},
		{
			desc:  "invalid duration",
			input: "end+infinity",
			options: Options{
				NowFunc: time.Now,
			},
			err: errors.New("can not parse duration: time: invalid duration +infinity"),
		},
		{
			desc:  "whitespace before op",
			input: "start  +24h",
			options: Options{
				TimeStart: time.Unix(0, 0).UTC(),
				NowFunc:   time.Now,
			},
			time: time.Date(1970, 01, 02, 0, 0, 0, 0, time.UTC),
		},
		{
			desc:  "whitespace after op",
			input: "start+  24h",
			options: Options{
				TimeStart: time.Unix(0, 0).UTC(),
				NowFunc:   time.Now,
			},
			time: time.Date(1970, 01, 02, 0, 0, 0, 0, time.UTC),
		},
		{
			desc:  "whitespace on both sides of op",
			input: "start + 24h",
			options: Options{
				TimeStart: time.Unix(0, 0).UTC(),
				NowFunc:   time.Now,
			},
			time: time.Date(1970, 01, 02, 0, 0, 0, 0, time.UTC),
		},
	} {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var ts time.Time
			err := setTime(&ts, test.input, test.options)

			if !reflect.DeepEqual(err, test.err) {
				t.Errorf("got error %q, want %q", err, test.err)
			}

			if err != nil {
				return
			}

			if ts != test.time {
				t.Errorf("got %q, want %q", ts, test.time)
			}
		})
	}
}
