package kernel

import "testing"

func TestLastIdentifier(t *testing.T) {
	for _, test := range []struct {
		desc  string
		input string
		pos   int
		out   string
		start int
		end   int
	}{
		{
			desc:  "empty",
			input: "",
			pos:   0,
			out:   "",
			start: 0,
			end:   0,
		},
		{
			desc:  "simple",
			input: "ident",
			pos:   5,
			out:   "ident",
			start: 0,
			end:   5,
		},
		{
			desc:  "whitespace",
			input: "abc + def",
			pos:   9,
			out:   "def",
			start: 6,
			end:   9,
		},
		{
			desc:  "whitespace front",
			input: "abc + def",
			pos:   3,
			out:   "abc",
			start: 0,
			end:   3,
		},
		{
			desc:  "whitespace middle",
			input: "abc + def * ghi",
			pos:   9,
			out:   "def",
			start: 6,
			end:   9,
		},
	} {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			out, start, end := lastIdentifier(test.input, test.pos)

			if out != test.out {
				t.Errorf("got ident %q, want %q", out, test.out)
			}

			if start != test.start {
				t.Errorf("got start %d, want %d", start, test.start)
			}

			if end != test.end {
				t.Errorf("got end %d, want %d", end, test.end)
			}
		})
	}
}
