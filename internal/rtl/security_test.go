package rtl

import (
	"strings"
	"testing"
	"unicode"
)

func TestUntrustedTextCannotSendTerminalControls(t *testing.T) {
	for _, mode := range []Mode{Off, Terminal, App, Auto} {
		for _, input := range []string{
			"title\x1b]52;c;Y29weQ==\x07 end",
			"title\x1b[2J end",
			"title\x1b]8;;https://example.invalid\x1b\\link\x1b]8;;\x1b\\ end",
			"title\x07\x00\x9b end",
		} {
			for _, output := range []string{mode.Line(input, LTR), mode.Fit(input, 80, LTR), mode.Fit(input, 6, LTR)} {
				if strings.ContainsFunc(output, unicode.IsControl) {
					t.Fatalf("mode %v retained controls: %q", mode, output)
				}
			}
		}
		if got := mode.Line("title\x1b[31m red\x1b[0m", LTR); got != "title red" {
			t.Fatalf("ANSI sequence remained: %q", got)
		}
	}
}
