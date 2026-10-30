// Package update finds newer releases of Claude Switcher on GitHub and replaces the running
// executable with the latest one. It uses the standard library only, and it never runs unless
// the user asks for an update or has the daily check turned on.
package update

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Releases is where the releases live. Tests point it at a local server.
var Releases = "https://github.com/JittoJoseph/claude-switcher/releases"

// Latest returns the newest released version, such as "3.0.1". It reads the tag from where the
// latest-release page redirects to, which needs no API call and has no rate limit.
func Latest(ctx context.Context, client *http.Client) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	c := *client
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, Releases+"/latest", nil)
	if err != nil {
		return "", err
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	loc := resp.Header.Get("Location")
	if resp.StatusCode/100 != 3 || !strings.Contains(loc, "/releases/tag/") {
		return "", errors.New("no release found")
	}
	version := strings.TrimPrefix(path.Base(loc), "v")
	if _, _, ok := parse(version); !ok {
		return "", errors.New("release tag is not a semantic version")
	}
	return version, nil
}

// Newer reports whether version a is newer than b, by semantic versioning ("v" prefix optional).
// A version that is not semver, such as "dev", is never newer nor older.
func Newer(a, b string) bool {
	ca, pa, okA := parse(a)
	cb, pb, okB := parse(b)
	if !okA || !okB {
		return false
	}
	for i := range ca {
		if ca[i] != cb[i] {
			return ca[i] > cb[i]
		}
	}
	if len(pa) == 0 || len(pb) == 0 { // a release is newer than its prereleases
		return len(pa) == 0 && len(pb) > 0
	}
	for i := 0; i < min(len(pa), len(pb)); i++ {
		if c := comparePre(pa[i], pb[i]); c != 0 {
			return c > 0
		}
	}
	return len(pa) > len(pb)
}

func parse(v string) (core [3]int, pre []string, ok bool) {
	v, build, hasBuild := strings.Cut(strings.TrimPrefix(v, "v"), "+")
	if hasBuild && !identifiers(build, false) {
		return core, nil, false
	}
	v, p, hasPre := strings.Cut(v, "-")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return core, nil, false
	}
	for i, s := range parts {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 || !digits(s) || len(s) > 1 && s[0] == '0' {
			return core, nil, false
		}
		core[i] = n
	}
	if hasPre {
		if !identifiers(p, true) {
			return core, nil, false
		}
		pre = strings.Split(p, ".")
	}
	return core, pre, true
}

// comparePre orders two prerelease identifiers: numbers by value and below words.
func comparePre(a, b string) int {
	numericA, numericB := digits(a), digits(b)
	switch {
	case numericA && numericB:
		if len(a) != len(b) {
			if len(a) > len(b) {
				return 1
			}
			return -1
		}
	case numericA:
		return -1
	case numericB:
		return 1
	}
	return strings.Compare(a, b)
}

func digits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func identifiers(s string, rejectLeadingZero bool) bool {
	for _, part := range strings.Split(s, ".") {
		if part == "" || rejectLeadingZero && digits(part) && len(part) > 1 && part[0] == '0' {
			return false
		}
		for _, c := range part {
			if c != '-' && !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') {
				return false
			}
		}
	}
	return true
}

// Disabled reports whether the environment turns update checks off: CLAUDE_SWITCHER_NO_UPDATE_CHECK,
// DO_NOT_TRACK or CI.
func Disabled(getenv func(string) string) bool {
	for _, name := range []string{"CLAUDE_SWITCHER_NO_UPDATE_CHECK", "DO_NOT_TRACK", "CI"} {
		if getenv(name) != "" {
			return true
		}
	}
	return false
}

// Notifier tells whether a newer release is out, asking GitHub at most once a day.
type Notifier struct {
	Current string // the running version; a build that is not semver ("dev") never checks
	Cache   string // file that remembers the last check, <vault>/cache/update.json
	Off     bool   // the user turned the check off
}

type lastCheck struct {
	Checked time.Time `json:"checked"`
	Latest  string    `json:"latest"`
}

// Newer returns the newest release when it is newer than Current, else "".
func (n Notifier) Newer(ctx context.Context) string {
	if n.Off {
		return ""
	}
	if _, _, ok := parse(n.Current); !ok {
		return ""
	}
	var last lastCheck
	if data, err := os.ReadFile(n.Cache); err == nil {
		_ = json.Unmarshal(data, &last) // a broken cache only means checking again
	}
	if time.Since(last.Checked) >= 24*time.Hour {
		if v, err := Latest(ctx, http.DefaultClient); err == nil {
			last.Latest = v
		}
		// A failed check counts too, so an offline computer is not slowed down on every run.
		last.Checked = time.Now()
		if data, err := json.Marshal(last); err == nil && os.MkdirAll(filepath.Dir(n.Cache), 0o700) == nil {
			_ = os.WriteFile(n.Cache, data, 0o600)
		}
	}
	if Newer(last.Latest, n.Current) {
		return last.Latest
	}
	return ""
}
