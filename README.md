# shellclass

[![Go Reference](https://pkg.go.dev/badge/github.com/richardwooding/shellclass.svg)](https://pkg.go.dev/github.com/richardwooding/shellclass)
[![Go](https://github.com/richardwooding/shellclass/actions/workflows/go.yml/badge.svg)](https://github.com/richardwooding/shellclass/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**What can this shell script do?** `shellclass` answers that from a real bash
AST — never from a regex over the command string — so a tool that runs shell
commands on someone's behalf can decide whether to just run one, ask first, or
refuse outright.

It was built for a coding agent's permission engine, where the question is
asked hundreds of times a session and getting it wrong means either running
something destructive or asking about `ls`.

```go
a := shellclass.Analyze("rm -rf build && npm install", ws)

a.Class        // destructive — the worst any one command can do
a.NeedsNetwork // true  — npm install fetches
a.Unknown      // false — every command was legible
a.HardDeny     // ""    — nothing catastrophic

for _, c := range a.Commands {
	c.Argv   // ["rm" "-rf" "build"]      then ["npm" "install"]
	c.Class  // destructive               then network
	c.Writes // ["/src/app/build"]        then []
}
```

## What makes it different from matching on strings

**It parses.** Every simple command reached through pipes, `&&`/`||`,
subshells, blocks, loops, `case`, function bodies, `$(…)`, `<(…)`, `find -exec`
and `xargs` payloads is found and classified. Wrappers are peeled — `env`,
`timeout`, `nice`, `nohup`, `stdbuf`, `busybox`, `xargs` — so `timeout 120
./deploy.sh` is classified as `./deploy.sh`, not as `timeout`.

**It knows what it cannot see.** This is the part that matters. Three states,
not two:

| | meaning | a caller should |
|---|---|---|
| `Unknown` | `eval`, `sh -c "$X"`, `$(…)` in command position, a `PATH=` override — the analyser cannot see through it | never auto-approve it |
| `Unrecognised` | the argv is fully legible, the program just is not in the table | let a rule naming it cover it |
| neither | classified from the table | act on `Class` |

Collapsing those two into "unsafe" is what makes a permission system either
useless or a lie. A script containing `eval` can never match a saved rule; a
script running your in-house `./bin/deploy` can, because you can read it.

`Installs` marks a command that writes outside the project — `brew install jq`,
`pip install`, `cargo install` — so a caller can widen exactly one call's
filesystem access rather than leaving a package prefix writable all session.

**It has a floor.** A small set of patterns set `HardDeny` — privilege
escalation, `curl | sh`, `base64 -d | sh`, `rm -rf` of `$HOME` or a filesystem
root, force-pushing a protected branch, reading a credential with `gh auth
token`, moving `.env` — and those are reported separately from `Class`
precisely so that no caller configuration can permit them.

**The tables are the security surface, and they are readable.** More than 500 programs
with subcommand granularity: `git` by verb (including which options name a
program to run, and `-C`/`--git-dir` tainting), `gh`/`glab` by noun and verb,
`docker`/`kubectl`/`terraform`/`helm` and 30+ cloud CLIs, package managers and
build tools, and `awk` and `sed` in detail — the two text tools whose
*arguments are programs*.

## Install

```sh
go get github.com/richardwooding/shellclass
```

## Use

Implement `Workspace` over whatever you already use to resolve paths and read
git state — six methods, no types of ours:

```go
type Workspace interface {
	Resolve(string) (abs string, inside bool, err error) // expands ~, resolves symlinks, reports containment
	IsSecretFile(string) bool                            // looks like a credential file
	IsProtected(string) bool                             // may never be written
	IsTracked(string) bool                               // tracked by git
	Home() string
	Roots() []string
}
```

Then call `Analyze`. The result carries the whole script's verdict plus a
per-command breakdown in `Commands`, each with its `Argv`, the peeled
`Program`, a `Prefix` naming what a saved rule should match where the handler
knows its own shape (`gh pr create` → `["gh", "pr"]`), and the `Reads`/`Writes`
its redirects and operands declared.

See [the package documentation](https://pkg.go.dev/github.com/richardwooding/shellclass)
and `example_test.go`.

## Scope

It answers one question and decides nothing. Whether to run a script, prompt
for it, sandbox it or refuse it is the caller's; this package exists to make
that decision an informed one.

It is a **leaf**: one dependency, [`mvdan.cc/sh`](https://github.com/mvdan/sh),
for the parser. No build tags, no syscalls, no filesystem access of its own —
every path question goes back to your `Workspace`.

**It is not a sandbox.** Static analysis tells you what a script *says* it will
do. Anything that can reach a shell can still surprise you, which is why
`Unknown` exists and why the intended deployment is analysis *plus* OS-level
confinement, not analysis alone.

## Status

Pre-1.0, `v0.x`: the API may still move. It is extracted from
[wright](https://github.com/richardwooding/wright), where it decides every
`bash` call, and it carries its tests with it — a behaviour table of 300+
scripts, targeted tests for each documented trap, and a fuzz target whose
invariant is that a hard deny survives whatever the fuzzer wraps around it,
with its crasher corpus committed.

## License

[MIT](LICENSE)
