# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-09-27

First release, extracted from [wright](https://github.com/richardwooding/wright)
where it has decided every `bash` call since the project began.

### Added

- `Analyze(script, Workspace) Analysis` — classify a shell script from a real
  bash AST: the damage `Class`, whether the script is `Unknown` (contains
  something the analyser cannot see through), which programs were
  `Unrecognised`, whether it needs the network or installs software, a
  `HardDeny` for the catastrophic patterns, and a per-command breakdown with
  the reads and writes each command declared.
- `Workspace`, the six-method interface a caller implements over its own
  notion of the project directory, and the optional `BranchProtector` for
  overriding which branches a force-push may not touch.
- `InjectingEnvVars()`, the environment variable names that change what a
  command does, for callers that also strip them before execution.

### Changed from the version inside wright

- `Command.RuleWords` is now `Command.Prefix`. The old name described what
  wright's permission engine did with it rather than what it is: the argv
  prefix identifying the command, where the program's handler knows its own
  shape.
