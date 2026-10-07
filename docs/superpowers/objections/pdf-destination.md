---
spec: docs/superpowers/specs/2026-10-03-pdf-destination.md
date: 2026-10-03
mode: spec
diaboli_model: grok-4.7
objections:
  - id: O1
    category: scope
    severity: medium
    claim: "CLI success still exits 0 without reporting the path, so once the file leaves the working directory the command gives no indication of where it was written, unlike TUI p."
    evidence: "FR1: exits 0, does not start the TUI, and does not write that filename in the working directory. Decision 8 requires the TUI status line to contain the directory and YYYY-MM-chart.pdf, not only the filename. No functional requirement says the CLI prints the path it wrote."
    disposition: accepted
    disposition_rationale: "A successful -chart-pdf prints the path it wrote, one line on stdout, the same path string TUI p puts on the status line. That covers the default directory, pdf_dir, and -o. Failure stays on stderr and prints no path."
  - id: O2
    category: implementation
    severity: high
    claim: "A relative pdf_dir, including the spec's ~/charts example, follows the process working directory and is not shell-expanded, so the config key does not give a stable directory and a leading tilde silently names a directory ~."
    evidence: "Decision 5: 'A relative pdf_dir is relative to the process working directory, the same way -o and -db are.' 'pdf_dir = \"~/charts\" is a directory whose name starts with ~, under the working directory, not the home directory.' Decision 3: 'A leading ~ or a $ name is not shell-expanded.' S14 locks that result. US2 wants the config directory so a path is not passed on every export. US1's reason is that the PDF 'is not left in whatever directory the program was started from.'"
    disposition: accepted
    disposition_rationale: "A stored directory must not follow the launch directory. pdf_dir is absolute, or it begins with ~/ and that prefix alone expands to the home directory. $ stays literal. Any other relative value, including . and charts, fails the export. The TUI still opens, nothing is written, and there is no working-directory fallback. If the home directory is unavailable, ~/… fails the export the same way."
  - id: O3
    category: implementation
    severity: high
    claim: "A non-string pdf_dir fails the config read for the whole process, so the TUI does not start and -o writes nothing, even though an unwritable directory must not block the TUI and a bad pdf_dir cannot corrupt the log."
    evidence: "Decision 4: 'Non-string pdf_dir fails when the config is read (fail-closed, like display_unit).' 'A missing or unwritable directory does not stop the TUI from opening. Only the export fails. -o does not bypass this read.' FR9: 'No PDF is written and the TUI does not start, including when -o is set.'"
    disposition: accepted
    disposition_rationale: "A non-string pdf_dir does not fail the config read and does not keep the log from opening. -o still writes, because that flag ignores pdf_dir. An export that would use the key fails, names pdf_dir, and writes nothing. A bad directory cannot store pounds as kilograms, so this key is not fail-closed like display_unit."
  - id: O4
    category: specification quality
    severity: high
    claim: "S3 and S5 require exit 0 when the home directory is unavailable and pdf_dir or -o is set, but they never supply -config or -db, while default database resolution fails in that environment and Decision 9 only keeps the log open for an explicit database path."
    evidence: "S3: 'XDG_DATA_HOME is empty and the home directory is unavailable' / '-chart-pdf 1990-11' / 'the process exits 0 and the PDF is in that pdf_dir.' S5: 'no usable home or XDG_DATA_HOME' and an absolute -o / 'the process exits 0.' Decision 9: 'An explicit -db (or $HDTOOLS_DB) still lets the log open; only the export needs the data directory when pdf_dir and -o are empty.' FR2: 'Export does not require a resolvable data directory when pdf_dir is set.' S15 is the only scenario that says the database path is given explicitly."
    disposition: accepted
    disposition_rationale: "S3 and S5 keep exit 0 when the home directory is unavailable, and they name -db and -config as temporary paths, as S15 already does. Default database and config resolution stay as they are."
  - id: O5
    category: specification quality
    severity: medium
    claim: "S12 and FR10 only forbid a PDF in the working directory and the data directory, so p on another screen can write into pdf_dir and still satisfy the acceptance text."
    evidence: "S12: 'no PDF is created in the working directory or under the data directory' and then 'no PDF is created in either directory.' FR10: 'writes no PDF in the working directory or the data directory.' Decision 8: 'p still writes no PDF on the list, the form, the month sheet, or the long-term chart, including nothing under the data directory.' The scenarios never mention pdf_dir."
    disposition: accepted
    disposition_rationale: "On the list, form, month sheet, and long-term chart, p writes no PDF in the working directory, the data directory, or pdf_dir. S12 sets pdf_dir to a third directory and checks that directory too. FR10 matches Decision 8."
  - id: O6
    category: specification quality
    severity: low
    claim: "When pdf_dir is '.', the file path is only the filename, which cannot also satisfy the status rule that the line include the directory and not be only the filename unless that relative directory is rewritten."
    evidence: "S14: 'When pdf_dir is . Then the PDF is 1990-11-chart.pdf in the working directory.' Decision 8: 'the status line contains the path written, including its directory and YYYY-MM-chart.pdf, not only the filename' and 'A relative pdf_dir is not required to be turned into an absolute path.' FR6 repeats that a relative pdf_dir need not be absolutized."
    disposition: accepted
    disposition_rationale: "pdf_dir = \".\" is a failed export under O2, so there is no success status to word. S14 drops that success case. The status rule stays: the line contains the path written, and a ~/… path need not be made absolute."
---

## O1 — scope — medium

### Claim

CLI success still exits 0 without reporting the path, so once the file
leaves the working directory the command gives no indication of where
it was written, unlike TUI `p`.

### Evidence

FR1 requires `-chart-pdf` with no `-o` and no `pdf_dir` to exit 0, not
start the TUI, and not write `YYYY-MM-chart.pdf` in the working
directory. It does not require the process to say where the file went.

Decision 8, for the same directory rule, requires the TUI status line
to contain the directory and `YYYY-MM-chart.pdf`, "not only the
filename." Nothing in the functional requirements or the documentation
section puts that cue on the CLI. The how-to and reference updates
describe the rule in general. They cannot name the directory actually
chosen for this run when `pdf_dir` is set.

### Why this matters

Today the file appears in the directory the command was run from, so a
silent exit is still discoverable with `ls`. After this slice that
`ls` is empty and exit 0 looks like a no-op, especially for someone
who still expects the old cwd behaviour. TUI `p` was given an on-screen
path for this reason. The CLI, which US1 treats as the other entry
point, was not. Two shippable readings both pass the scenarios: keep
stdout empty, or print the path. Only the second tells the person who
just exported the chart which directory received a personal weight
chart.

## O2 — implementation — high

### Claim

A relative `pdf_dir`, including the spec's `~/charts` example, follows
the process working directory and is not shell-expanded, so the config
key does not give a stable directory and a leading tilde silently names
a directory `~`.

### Evidence

US1's reason for moving the file is that it "is not left in whatever
directory the program was started from." US2 is a config directory "so
I do not pass a path on every export."

Decision 3:

> A leading `~` or a `$` name is not shell-expanded.

Decision 5:

> A relative `pdf_dir` is relative to the process working directory,
> the same way `-o` and `-db` are. It is not resolved against the home
> directory or the data directory. … `pdf_dir = "~/charts"` is a
> directory whose name starts with `~`, under the working directory,
> not the home directory.

S14 then requires `pdf_dir` of `charts` to follow the working
directory, and `~/charts` to be a literal directory there, not under
the home directory.

### Why this matters

The default data directory does fix the cwd problem. The new config
key puts it back for every relative value, and it does so with exit 0.
`pdf_dir = "charts"` writes a different folder for each directory
`hdtools` is started from, which is the behaviour US1 is removing.
`pdf_dir = "~/charts"` and `pdf_dir = "$XDG_DATA_HOME/charts"` are the
strings a config file invites. Both succeed by creating a directory
whose name starts with `~` or `$` under the working directory. The
chart is a personal weight record, mode `0600`, in a place the user
will not think to look. The spec rejects the home directory and the
data directory as bases and treats the value like `-o` and `-db`.
Those are one-shot flags. A value stored in `config.toml` is read
again from whatever cwd the next launch happens to have. Executed
exactly as written, US2 is true only for an absolute path the user
already has to get right with no `~` and no `$`.

## O3 — implementation — high

### Claim

A non-string `pdf_dir` fails the config read for the whole process, so
the TUI does not start and `-o` writes nothing, even though an
unwritable directory must not block the TUI and a bad `pdf_dir` cannot
corrupt the log.

### Evidence

Decision 4:

> Non-string `pdf_dir` fails when the config is read (fail-closed, like
> `display_unit`). … A missing or unwritable directory does not stop
> the TUI from opening. Only the export fails. `-o` does not bypass
> this read: a non-string `pdf_dir` still fails before any PDF is
> written.

FR9:

> No PDF is written and the TUI does not start, including when `-o` is
> set.

S11 covers `-chart-pdf` including when `-o` names a file. The TUI
lockout is only in FR9.

### Why this matters

The same decision uses two failure policies. A directory that cannot
be written must leave the log usable and fail only the export. A type
mismatch on the same key makes the log unreachable, including a one-off
`-o` that Decision 6 says ignores `pdf_dir` entirely. The
`display_unit` analogy does not carry the reason that key is
fail-closed. Project `ARCH_DECISIONS` fail-closes `display_unit`
because a bad unit would store pounds as kilograms, and fail-opens
`[colors]` because a bad colour cannot corrupt the series. `pdf_dir`
is in the second class: the page, the filename, and the stored
kilograms do not change. A bare path without quotes is already illegal
TOML and already fails the read. The new blast radius is numbers,
booleans, and arrays (`pdf_dir = 3`, `pdf_dir = true`), which now
block daily logging until `config.toml` is hand-edited. That is a
larger outcome than an export-path typo, and it is the opposite of the
unwritable-directory rule in the same bullet.

## O4 — specification quality — high

### Claim

S3 and S5 require exit 0 when the home directory is unavailable and
`pdf_dir` or `-o` is set, but they never supply `-config` or `-db`,
while default database resolution fails in that environment and
Decision 9 only keeps the log open for an explicit database path.

### Evidence

S3, second case:

> **Given** the same log, an absolute `pdf_dir`, and no `-o`
> **And** `XDG_DATA_HOME` is empty and the home directory is
> unavailable
> **When** `-chart-pdf 1990-11` runs
> **Then** the process exits 0 and the PDF is in that `pdf_dir`

S5, second case:

> **Given** the same log and no usable home or `XDG_DATA_HOME`
> **And** `pdf_dir` is an absolute directory that does not exist
> **When** `-chart-pdf 1990-11 -o` is an absolute file path
> **Then** the process exits 0 and the PDF exists only at that path

Decision 9:

> An explicit `-db` (or `$HDTOOLS_DB`) still lets the log open; only
> the export needs the data directory when `pdf_dir` and `-o` are
> empty.

FR2 says the export does not need a resolvable data directory when
`pdf_dir` is set. It does not say startup does not need one. S15 is
the only scenario that states the database path is explicit so opening
the log did not need the data directory.

That startup is the rule this spec says it reuses. `config.DataDir`
returns `home directory: …` when `XDG_DATA_HOME` is empty and
`os.UserHomeDir` fails. `resolveDBPath` calls `DefaultDBPath` unless
`-db` or `$HDTOOLS_DB` is set. `run` does that, and resolves the
config path, before `exportChartPDF`. On Unix, `os.UserConfigDir`
also fails when both `XDG_CONFIG_HOME` and `HOME` are unset. This
slice does not change either resolver.

### Why this matters

Read literally, S3 and S5 cannot pass. With home unavailable and no
explicit database path, the process exits on the existing
home-directory failure while opening the log, before the PDF rule
runs. A test that injects `-config` and `-db`, as S15 describes and
as current chart tests do, can exit 0. A test that only clears `HOME`
and `XDG_DATA_HOME` cannot. A third reading changes default config and
database resolution so a home-less process still exports, which this
spec never asks for and which would blur the S15 failure it does ask
for. "Export does not need the data directory" and "the process exits
0" are not the same requirement until the scenario says how the log
and the config file are found.

## O5 — specification quality — medium

### Claim

S12 and FR10 only forbid a PDF in the working directory and the data
directory, so `p` on another screen can write into `pdf_dir` and still
satisfy the acceptance text.

### Evidence

Decision 8:

> `p` still writes no PDF on the list, the form, the month sheet, or
> the long-term chart, including nothing under the data directory.

S12's check is narrower. On the list: "no PDF is created in the
working directory or under the data directory." On the month sheet,
the long-term chart, and the form: "no PDF is created in either
directory." FR10 repeats those two locations only. None of those Thens
mention `pdf_dir`, and the list case does not set `pdf_dir` either
way.

### Why this matters

This repo writes the failing test from the acceptance scenario. A `p`
handler that exports on the list or the long-term chart into the
configured directory, and not into the cwd or the data directory,
turns S12 and FR10 green while violating Decision 8. The old S4 said
"no PDF is created" with no location limit. This amendment adds the
data directory and, by naming only those two directories, drops the
rest. The false green shows up only when `pdf_dir` is a third path,
which is the new place a mistaken export would go.

## O6 — specification quality — low

### Claim

When `pdf_dir` is `.`, the file path is only the filename, which
cannot also satisfy the status rule that the line include the
directory and not be only the filename unless that relative directory
is rewritten.

### Evidence

S14:

> **When** `pdf_dir` is `.`
> **Then** the PDF is `1990-11-chart.pdf` in the working directory

Decision 8:

> On success the status line contains the path written, including its
> directory and `YYYY-MM-chart.pdf`, not only the filename. … A
> relative `pdf_dir` is not required to be turned into an absolute
> path.

FR6 repeats both the "directory and `YYYY-MM-chart.pdf`" requirement
and that a relative `pdf_dir` need not be absolutized. S7 and S8 only
check the status line for absolute directories.

### Why this matters

The file location is clear: the working directory, under the bare
filename. The status text is not. One implementer sets the status to
the path that was written, `1990-11-chart.pdf`, which is only the
filename and fails Decision 8's "not only the filename" sentence.
Another displays `./1990-11-chart.pdf` or the absolute cwd so a
directory is present, which Decision 8 does not require for a relative
`pdf_dir` and which is not the path S14 names. The chart still lands
in the right place. The status line for this one config value will
differ, and no scenario picks which string is right.

## Explicitly not objecting to

- **Default XDG data directory:** `idea.md` asks for
  `$XDG_DATA_HOME/hdtools` (otherwise `~/.local/share/hdtools`), and
  S1 and S2 lock that default, beside `hdtools.db` rather than in a
  subdirectory.
- **`-db` and `$HDTOOLS_DB` not moving the PDF:** Decision 2 and S10
  state this on purpose. The backlog names the XDG data directory, not
  the directory of a database opened from somewhere else.
- **Silent overwrite with no backup:** S13 and the out-of-scope list
  keep the shipped rule. Confirm-before-overwrite was already
  adjudicated on the 2026-09-12 spec and this slice does not reopen
  it.
- **Leaving `2026-09-12-pdf-monthly-charts.md` unedited:** Decision 11
  names the cwd sentences it amends (Decision 2, Decision 10, S3,
  FR1, and the XDG leftover). The older file stays the picture record.
- **Not building filled sheets, long-term PDFs, or blank sheets:**
  Decision 12 and `idea.md` defer those kinds and already say they
  reuse this directory rule.
- **No PDF-specific environment variable:** Decision 1 and the
  out-of-scope list. `$XDG_DATA_HOME` still selects the default data
  directory. The sentence is not a ban on that variable.
- **`-o` staying a full file path:** US3, Decision 6, and S6. Joining
  `-o` onto `pdf_dir` would change the flag the 2026-09-12 spec
  already shipped.
- **FR7's "Covered failures" list:** Decisions 6 through 10 already
  forbid a working-directory copy on any failed write. The list reads
  as the scenario set, not as permission to fall back on EACCES or a
  full disk.
- **Directory mode `0755`:** Decision 7 matches today's parent
  creation for this PDF, umask included. The scenarios lock the file
  at `0600`, which is the mode that matters for the chart bytes.
- **In-app help remaining `p pdf`:** That chord does not claim the
  current directory. FR12 updates the key reference, which is the
  line that does.
