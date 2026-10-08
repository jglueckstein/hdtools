# Plan — Charm v2 upgrade

**Spec**:
[`docs/superpowers/specs/2026-10-07-charm-v2-upgrade.md`](../specs/2026-10-07-charm-v2-upgrade.md)
**Slice**: S1 of
[`charm-v2-then-online-charts.md`](../slices/charm-v2-then-online-charts.md)
**Status**: approved
**Objections**:
[`charm-v2-upgrade.md`](../objections/charm-v2-upgrade.md)
(O1–O3 accepted)

No production code until the spec's scenarios exist as failing tests.

Red, for this slice, is not a second suite. The existing scenarios
stay the contract. Test edits are only the v2 types needed to compile
the same assertions: how a key is constructed, and how the visible
text is read. Do not change expected glyphs, loss numbers, key
destinations, PDF paths, or colour SGR expectations in order to go
green. Profile-dependent SGR is read from the written output
(Decision 10). The model text may contain full-fidelity colour.
`NO_COLOR`, bold, reverse, and `>` stay assertions on the model
text.

Two additions are justified, because the existing suite cannot see
them:

-   The space bar. No current test presses it on the month sheet or
    the day form. Add `TestSpaceBarStaysSpace` so it fails if that
    key is handled by the old printed name `" "` or by the word
    `"space"`. The test builds a v2 key whose code is a space, whose
    text is a single space, and whose printed name is `"space"`.
-   ntcharts. Add `TestModuleRequiresCharmV2`. It fails on the
    current module file, which still names the v1 libraries, and it
    fails if ntcharts is required or imported.

S2, S3, and S10 through S18 are already implemented. Do not copy them
into new tests. S14's ctrl+c quits from the same list branch as `q`;
`TestQuitKeys` stays the proof for `q`. Do not add a quit test. S16
is no alternate screen and no mouse. The model-only program still
enables modifyOtherKeys and Kitty disambiguation. Do not try to
turn those off. Do not add a test for the wire protocol.

If a v2 API other than the list below fails to compile, stop and
report it. Do not guess a replacement.

## Module structure

| File | Why |
| --- | --- |
| `go.mod`, `go.sum` | Fetch the current v2 releases of the three libraries. Do not change the `go 1.27.0` line. Do not add ntcharts. |
| `cmd/hdtools/main.go` | Import path only. `tea.NewProgram` still receives the model and nothing else. Do not try to disable keyboard disambiguation. |
| `cmd/hdtools/main_test.go` | `TestModuleRequiresCharmV2` only. Do not retarget PDF paths. |
| `internal/tui/app.go` | Key type, space-bar match, and `View` returning a view value whose visible text is the same string. |
| `internal/tui/form.go` | bubbles import, text-field width setter, space-bar match. Field behaviour stays. |
| `internal/tui/month.go` | bubbles import and text-field width setter. Paint stays. |
| `internal/tui/month_save.go` | Import path only. Workout space still saves in place. |
| `internal/tui/chart.go` | Import paths and key type. Do not change paint, the caption, or `p`. |
| `internal/tui/longchart.go` | Import path and key type. Do not change windows or paint. |
| `internal/tui/style.go` | Drop `ColorProfile` and `SetColorProfile`. `Render` keeps bold and reverse in the model text. Do not set `lipgloss.Writer.Profile`, and do not downsample in the app. |
| `internal/tui/color_test.go` | Profile-dependent SGR is asserted after the view content is written through `colorprofile.Writer` at the profile under test. Expected SGR stays. |
| `internal/tui/layout.go` | Lip Gloss import only. `visPad` and `lipgloss.Width` stay. |
| `internal/tui/layout_test.go` | Import path, key construction, view content. Alignment expectations stay. |
| `internal/tui/chart_test.go` | Key construction (`press` included) and view content. Indexed colour assertions read the written output at the 16-color floor. Braille and PDF expectations stay on the model text. |
| `internal/tui/longchart_test.go` | Key construction and view content. Indexed colour assertions read the written output at the 16-color floor. Window names and Braille expectations stay on the model text. |
| `internal/tui/app_test.go` | v2 key construction, read visible text from the view content, and `TestSpaceBarStaysSpace`. Indexed colour assertions read the written output at the 16-color floor. |
| `internal/tui/month_test.go` | Leave it unless a compile error says otherwise, except an indexed-colour assertion, which reads the written output at the 16-color floor. |
| `idea.md` | Change record names `charm.land/bubbles/v2` next to the other two modules. Do not rewrite the chart paragraphs. |
| `CHANGELOG.md` | Unreleased Changed bullet, no pull-request number. |

Do not edit `internal/dailylog` (trend or analysis), `internal/chartpdf`
paint, `internal/chartspan`, or the PDF directory resolver in
`internal/config`. Do not edit `internal/tui/braille.go`,
`internal/tui/delta.go`, or `internal/tui/select.go`. Do not edit
`docs/reference` or `docs/how-to`. Do not add
`github.com/NimbleMarkets/ntcharts`.

`internal/tui/form_test.go` does not import these libraries. Leave
it unless a compile error says otherwise, then stop and report it.
`internal/tui/month_test.go` does not import them either. Its
indexed-colour assertion reads the written output at the 16-color
floor. Leave the rest unless a compile error says otherwise.

## Algorithm notes

-   Imports. `github.com/charmbracelet/bubbletea` becomes
    `charm.land/bubbletea/v2`. `github.com/charmbracelet/lipgloss`
    becomes `charm.land/lipgloss/v2`.
    `github.com/charmbracelet/bubbles/textinput` becomes
    `charm.land/bubbles/v2/textinput`. Keep the `tea` alias. Fetch
    the current v2 release of each. Do not hand-pin a version in
    this plan. Do not bump `go 1.27.0`. bubbletea v2 asks for go
    1.26.0, which this module already exceeds. If a fetched release
    refuses to build on go 1.27.0, stop and report. Do not raise the
    go line.
-   `View() string` becomes `View() tea.View`. Visible text is
    `tea.View.Content`. `tea.NewView` sets it. Return a view whose
    content is the same text `View` returns today. Do not set
    `AltScreen` or a mouse mode. If `tea.NewView` does not compile
    as the call that sets content, stop and report. Tests that read
    `app.View()` or `model.View()` as a string read that content.
    `visible` still takes a string. A test-only helper that returns
    the content is fine. Do not add a production method to keep the
    old signature.
-   `tea.KeyMsg` is now an interface. Key presses are
    `tea.KeyPressMsg`. v1 `msg.Type` is `msg.Code` (a rune). v1
    `msg.Runes` is `msg.Text` (a string). v1 `tea.KeyRunes` is
    `len(msg.Text) > 0`. `msg.String()` for the space bar is
    `"space"`. `Code` is still `' '` and `Text` is still `" "`.
    Match the space bar on `Code` or `Text`, not on the printed name
    `" "` and not on the word `"space"`.
-   Do that space-bar check before the "one typed character starts
    an edit" check. A space bar's text is one character long, so a
    bare `len(msg.Text) > 0` test would start an edit on an idle
    weight cell. Exclude the space bar first. A non-workout cell
    that is not editing starts an edit only for one other character.
    The old `KeyRunes` test is "text is non-empty", narrowed to one
    character that is not the space bar.
-   Month-sheet workout, not editing: space bar still runs the
    in-place save (flip the flag, do not advance, do not change
    day). Idle space on weight, sleep, steps, or note: no edit, no
    save, no move. Space while a text cell is editing: pass the key
    through so the field gains one space, and do not save. Day-form
    text field: pass the key through so the field gains one space,
    and do not flip workout. Day-form workout: flip yes/no and do
    not write. `q` and `ctrl+c` keep today's branches. Enter, Tab,
    Esc, the arrows, `q`, and ctrl+c must still match those keys.
    If one of those printed names is no longer the name the code
    matches, stop and report. Do not guess new names. Do not try to
    turn off modifyOtherKeys or Kitty disambiguation.
-   If a text field shows the word "space", or shows nothing, and
    the code passed the key through with code `' '` and text `" "`,
    stop and report. Do not change the expected character to the
    word, and do not fork the text field.
-   Tests construct `tea.KeyPressMsg`. A character key sets `Code`
    to that rune and `Text` to that one-character string. `press`
    must build those messages. `press` of `" "` is the space bar
    (code space, text space), not the five letters. Named keys
    (enter, esc, tab, arrows) set `Code` to the v2 rune that
    replaces the old type constant. If `tea.KeyEnter` and the other
    old names are not accepted as that code, stop and report. Do
    not invent rune values. `flushMonthKey` passes a `KeyPressMsg`
    and still runs a save command. A text-field blink command is
    not a save. `tea.QuitMsg` still exists. `tea.WindowSizeMsg`
    still supplies `Width` for `termCols`. No test sends a resize
    today. Do not add one, and do not invent a replacement for
    `Width`.
-   textinput. The exported `Width` field becomes `SetWidth` /
    `Width()`. Assignments today are 32 and 12 on the day form and
    24 on the month sheet. Keep those numbers. Keep `SetValue`,
    `Focus`, `CursorEnd`, and `Position` unless the compiler
    rejects them. Also keep `Placeholder`, `CharLimit`, `Prompt`,
    `Blur`, `Update`, `View`, and `Value` unless the compiler
    rejects one of them. If it does, stop and report. Do not invent
    other textinput changes. `DefaultKeyMap` is now a function if
    any code reads the variable. This repo does not. Do not add a
    read.
-   `lipgloss.Color("5")` and `Foreground` still work, but `Color`
    is a function. `lipgloss.ColorProfile` and `SetColorProfile` are
    removed. `Style.Render` does not consult a profile, so delete
    the Ascii-to-ANSI force in `style.go`. Do not replace it with
    `lipgloss.Writer.Profile`. That field does not change `Render`.
    Bold and reverse stay in the model text. The app does not
    downsample the view string.
-   Profile-dependent SGR is asserted on the written output. Write
    the view content through `colorprofile.Writer` with `Profile`
    set to the profile under test, then assert on that buffer.
    Tests that already force truecolor use truecolor. Tests that do
    not set a profile use the 16-color floor (ANSI), which is what
    the old Ascii force made the view string. Do not change the
    expected SGR: `38;2` on truecolor, indexed downshift on
    16-color. `NO_COLOR`, bold, reverse, and `>` stay assertions on
    the model text, not on that buffer. Do not convert a sequence
    twice.
-   `lipgloss.Width` stays the display-cell width. Keep `visPad`.
-   `cmd/hdtools/main.go` stays `tea.NewProgram` with the model
    only. Do not pass an alternate-screen option or a mouse option.
    Do not try to disable modifyOtherKeys or Kitty disambiguation.
    S16 has no test.
-   `TestModuleRequiresCharmV2` reads `go.mod` from the module root
    and scans `.go` files under `cmd/` and `internal/` only. Docs
    mention ntcharts in prose; do not scan them. The `go` line is
    still `1.27.0`. The require block names
    `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, and
    `charm.land/bubbles/v2`. It does not require
    `github.com/charmbracelet/bubbletea`,
    `github.com/charmbracelet/lipgloss`,
    `github.com/charmbracelet/bubbles`, or any ntcharts module.
    No scanned `.go` file imports ntcharts or the three v1 module
    paths.
-   `TestSpaceBarStaysSpace` lives in `internal/tui/app_test.go` and
    uses the November 1990 sheet the edit-key tests use. Run a
    returned save the way those tests do. Subtests are S4 through
    S9: workout space flips the stored flag and stays on that day
    and column; idle space on a non-workout cell does not edit or
    write; the character `8` on a weight cell starts an edit whose
    text is `8`; space while editing inserts one space and does not
    write; day-form text space inserts one space and leaves workout
    alone; day-form workout space flips yes/no and does not write.
    The key's printed name in the test is `"space"`. The assertions
    are the behaviours above, not that printed name.

## FR mapping

| FR | Tests |
| --- | --- |
| FR1 | `TestModuleRequiresCharmV2` |
| FR2 | `TestViewListsTrendedLogs`, `TestViewEmptyStore`, `TestListShowsPoundsWhenConfigured`, `TestLoadSelectsToday`, `TestGotoTodayListDoesNotCreateToday`, `TestNewOpensForm`, `TestMOpensMonthSheet`, `TestMonthCellSaveWritesWeight`, `TestMonthTabWhileEditingSavesAndAdvances`, `TestMonthTabInvalidDoesNotAdvance`, `TestMonthTabWhenNotEditingMoves`, `TestMonthEnterWhileEditingSavesAndMovesDown`, `TestMonthDownWhileEditingSavesAndMovesDown`, `TestMonthUpWhileEditingSavesAndMovesUp`, `TestMonthEnterInvalidDoesNotMove`, `TestMonthDownInvalidDoesNotMove`, `TestMonthUpInvalidDoesNotMove`, `TestMonthEnterOnLastDayStays`, `TestMonthDownOnLastDayStays`, `TestMonthUpOnFirstDayStays`, `TestMonthLeftRightWhileEditingStayInCell`, `TestMonthLeftAtStartRightAtEndStayInCell`, `TestMonthEnterWhenNotEditingOpensForm`, `TestMonthDownUpWhenNotEditingMoveDay`, `TestMonthLeftRightWhenNotEditingMoveColumn`, `TestMonthEscWhileEditingCancels`, `TestMonthEscBeforeSaveCancels`, `TestMonthEnterSnapshotIgnoresLaterCursor`, `TestMonthSecondVerticalKeyBeforeSaveDoesNotSkip`, `TestMonthSecondEnterAfterEnterOpensForm`, `TestMonthEnterAfterDownOpensForm`, `TestMonthEnterAfterUpOpensForm`, `TestMonthSecondEnterOnLastDayOpensForm`, `TestChartOpensFromListAndEscapes`, `TestLongChartOpensFromListAndEscapes` |
| FR3 | `TestSpaceBarStaysSpace` (S4–S7) |
| FR4 | `TestSpaceBarStaysSpace` (S8–S9) |
| FR5 | `TestChartBrailleBothColumnsNovember`, `TestChartBrailleCloseMarksBins`, `TestChartBrailleStemIsGreen`, `TestChartBrailleEightByFour`, `TestChartBrailleStemInsideRow`, `TestChartBrailleNoColorCarryIsBold`, `TestChartBrailleNoColorKeepsDots`, `TestBrailleBinS2Marks`, `TestBrailleBinS4Stem`, `TestBrailleBinLastIsYMax`, `TestBrailleBits`, `TestChartShowsMonthlyLossAndDeficit`, `TestChartCurrentMonthDividesByElapsedDays`, `TestChartLossUsesDisplayUnit` |
| FR6 | `TestLongChartCyclesKinds`, `TestLongChartQuarterlyClipsAtToday`, `TestLongChartCompleteStartsAtFirstLog`, `TestLongChartLossAndDeficit`, `TestLongChartBrailleKeepsDayColumns`, `TestLongChartBrailleStaysBucketed`, `TestLongChartBrailleNoColorBoldTrend`, `TestLongChartTwoLinesNoMarksOrStems` |
| FR7 | `TestChartPDFDefaultWritesDataDir`, `TestChartPDFDefaultUsesStandInHome`, `TestChartPDFDirAbsolute`, `TestChartPDFOAbsoluteWins`, `TestChartPDFRelativeOStaysInCwd`, `TestChartPWritesPDF`, `TestChartPUsesPDFDir`, `TestChartPIgnoredOnList`, `TestChartPWhileEditingIsText`, `TestChartPIgnoredOnForm`, and the other PDF tests in `cmd/hdtools/main_test.go` and `internal/tui/chart_test.go` |
| FR8 | `TestViewUsesHexOnTruecolor`, `TestHexDownshiftsOnSixteenColor`, `TestNoColorStripsChromaKeepsStructure`, `TestEmptyNoColorKeepsDefaultChroma`, `TestTrendIsBoldWithAndWithoutColor`, `TestNamedAndHexMix`, `TestSelectionReverseWithOptionalForeground`, `TestLoadInvalidColorFallsBack`, `TestLoadRejectsUnknownUnit` |
| FR9 | `TestQuitKeys` for list `q`. Chart and long-term `q`, ctrl+c, and Esc stay the current branches (`TestChartOpensFromMonthAndEscapes`, `TestChartOpensFromListAndEscapes`, `TestLongChartOpensFromListAndEscapes`, `TestLongChartOpensFromMonthAndEscapes`, `TestLongChartOpensFromMonthlyChartAndEscapes`). No new quit test. S16 has no test: `NewProgram` stays model-only. |
| FR10 | `TestParseFormStoresPoundsAsKilograms`, `TestApplyTrendMatchesPencilExample`, `TestListDeltaPositive`, `TestListDeltaNegative`, `TestListDeltaZero`, `TestListDeltaNOCOLORKeepsSign`, `TestDeltaColumnAligns`, `TestMonthDeltaMatchesList`, `TestVisPadIgnoresANSI` |
| FR11 | `CHANGELOG.md` bullet below. No Go test. `docs/reference` and `docs/how-to` are not edited. |

Every other existing test stays green with the same expectations.
`go test ./...` is the net.

## Test list

Existing tests, expectations unchanged. Key construction and view
content are the only edits.

-   `TestViewListsTrendedLogs` — S2. Date and weight still in the
    visible text.
-   `TestViewEmptyStore` — S2. Empty list still offers the same
    help, including quit.
-   `TestListShowsPoundsWhenConfigured` — S2. Display unit converts;
    storage stays kilograms.
-   `TestLoadSelectsToday` — S2. List opens on the closest row.
-   `TestGotoTodayListDoesNotCreateToday` — S2. `t` does not insert
    a row.
-   `TestNewOpensForm` — S2. `n` opens the day form.
-   `TestMOpensMonthSheet` — S2. `m` shows November 1990.
-   `TestMonthCellSaveWritesWeight` — S3. Stored kilograms.
-   `TestMonthTabWhileEditingSavesAndAdvances` — S3. Tab saves and
    moves to the next cell.
-   `TestMonthTabInvalidDoesNotAdvance` — S3. Invalid text stays.
-   `TestMonthTabWhenNotEditingMoves` — S3. Tab without an edit.
-   `TestMonthEnterWhileEditingSavesAndMovesDown` — S3. Enter moves
    down.
-   `TestMonthDownWhileEditingSavesAndMovesDown` — S3. Down moves
    down.
-   `TestMonthUpWhileEditingSavesAndMovesUp` — S3. Up moves up.
-   `TestMonthEnterInvalidDoesNotMove` — S3. Invalid Enter stays.
-   `TestMonthDownInvalidDoesNotMove` — S3. Invalid Down stays.
-   `TestMonthUpInvalidDoesNotMove` — S3. Invalid Up stays.
-   `TestMonthEnterOnLastDayStays` — S3. No wrap past day 30.
-   `TestMonthDownOnLastDayStays` — S3. Same edge for Down.
-   `TestMonthUpOnFirstDayStays` — S3. No wrap before day 1.
-   `TestMonthLeftRightWhileEditingStayInCell` — S3. Caret moves;
    `Position` stays the caret.
-   `TestMonthLeftAtStartRightAtEndStayInCell` — S3. Caret does not
    leave the text.
-   `TestMonthEnterWhenNotEditingOpensForm` — S3. Enter opens the
    form.
-   `TestMonthDownUpWhenNotEditingMoveDay` — S3. Arrows move the
    day.
-   `TestMonthLeftRightWhenNotEditingMoveColumn` — S3. Arrows move
    the column.
-   `TestMonthEscWhileEditingCancels` — S3. Esc writes nothing.
-   `TestMonthEscBeforeSaveCancels` — S3. Esc drops a save already
    queued.
-   `TestMonthEnterSnapshotIgnoresLaterCursor` — S3. The save uses
    the cell from the key.
-   `TestMonthSecondVerticalKeyBeforeSaveDoesNotSkip` — S3. A second
    key does not double-move.
-   `TestMonthSecondEnterAfterEnterOpensForm` — S3. Enter after a
    finished edit opens the form.
-   `TestMonthEnterAfterDownOpensForm` — S3. Same after Down.
-   `TestMonthEnterAfterUpOpensForm` — S3. Same after Up.
-   `TestMonthSecondEnterOnLastDayOpensForm` — S3. Last day, second
    Enter opens the form and does not add a row.
-   `TestChartBrailleBothColumnsNovember` — S10. Thirty Braille
    columns.
-   `TestChartBrailleCloseMarksBins` — S10. Bins and the padded Y
    labels stay.
-   `TestChartBrailleStemIsGreen` — S10. Stem, mark, and carry
    colours stay.
-   `TestChartBrailleEightByFour` — S10. Plot cells stay Braille.
-   `TestChartBrailleStemInsideRow` — S10. Shared-cell dots stay.
-   `TestChartBrailleNoColorCarryIsBold` — S10. `NO_COLOR` keeps the
    rune and bold on the carry.
-   `TestChartBrailleNoColorKeepsDots` — S10. Dots survive
    `NO_COLOR`.
-   `TestBrailleBinS2Marks` — S10. Bin helper stays.
-   `TestBrailleBinS4Stem` — S10. Stem bins stay.
-   `TestBrailleBinLastIsYMax` — S10. Top bin stays.
-   `TestBrailleBits` — S10. Dot masks stay.
-   `TestChartShowsMonthlyLossAndDeficit` — S10. Caption words stay.
-   `TestChartCurrentMonthDividesByElapsedDays` — S10. Current month
    ends at today; the divisor is elapsed days.
-   `TestChartLossUsesDisplayUnit` — S10. Loss uses the display
    unit.
-   `TestLongChartCyclesKinds` — S11. `]` cycles Semiannual, Annual,
    Complete, Quarterly. `[` from Quarterly is Complete.
-   `TestLongChartQuarterlyClipsAtToday` — S11. Quarterly ends at
    today when the data does.
-   `TestLongChartCompleteStartsAtFirstLog` — S11. Complete starts
    at the first log.
-   `TestLongChartLossAndDeficit` — S11. Same pencil caption.
-   `TestLongChartBrailleKeepsDayColumns` — S11. Day columns stay
    Braille.
-   `TestLongChartBrailleStaysBucketed` — S11. Wide spans stay
    bucketed.
-   `TestLongChartBrailleNoColorBoldTrend` — S11. Bold trend under
    `NO_COLOR`.
-   `TestLongChartTwoLinesNoMarksOrStems` — S11. The current
    long-term series, not a new picture.
-   `TestChartPDFDefaultWritesDataDir` — S12. CLI default path.
-   `TestChartPDFDefaultUsesStandInHome` — S12. No `XDG_DATA_HOME`.
-   `TestChartPDFDirAbsolute` — S12. `pdf_dir` wins.
-   `TestChartPDFOAbsoluteWins` — S12. `-o` wins.
-   `TestChartPDFRelativeOStaysInCwd` — S12. Relative `-o`.
-   `TestChartPWritesPDF` — S12. TUI `p` path and mode `0600`.
-   `TestChartPUsesPDFDir` — S12. TUI `p` honours `pdf_dir`.
-   `TestChartPIgnoredOnList` — S12. `p` writes nothing off the
    monthly chart.
-   `TestChartPWhileEditingIsText` — S12. `p` during a month edit
    is text, not a PDF.
-   `TestChartPIgnoredOnForm` — S12. `p` on the form is text.
-   `TestViewUsesHexOnTruecolor` — S13. Written truecolor output
    has `38;2` for `#00ff00` and `#0D47A1`.
-   `TestHexDownshiftsOnSixteenColor` — S13. Written 16-color
    output has no `38;2` and has indexed green.
-   `TestNoColorStripsChromaKeepsStructure` — S13. No chroma.
    Bold, reverse, and `>` stay.
-   `TestEmptyNoColorKeepsDefaultChroma` — S13. Empty `NO_COLOR`
    keeps blue weight and red trend.
-   `TestTrendIsBoldWithAndWithoutColor` — S13. Bold both ways.
-   `TestNamedAndHexMix` — S13. Named blue and hex `38;2` together.
-   `TestSelectionReverseWithOptionalForeground` — S13. Reverse
    stays.
-   `TestLoadInvalidColorFallsBack` — S13. `[colors]` fail-open.
-   `TestLoadRejectsUnknownUnit` — S13. `display_unit` fail-closed.
-   `TestQuitKeys` — S14. `q` on the list quits. Do not add a
    ctrl+c test.
-   `TestChartOpensFromListAndEscapes` — S15. Esc from the chart
    returns to the list.
-   `TestChartOpensFromMonthAndEscapes` — S15. Esc returns to the
    month sheet.
-   `TestLongChartOpensFromListAndEscapes` — S15. Esc returns to
    the list.
-   `TestLongChartOpensFromMonthAndEscapes` — S15. Esc returns to
    the month sheet.
-   `TestLongChartOpensFromMonthlyChartAndEscapes` — S15. Esc
    returns to the monthly chart.
-   `TestParseFormStoresPoundsAsKilograms` — S17. Pounds store as
    kilograms.
-   `TestApplyTrendMatchesPencilExample` — S17. Trend stays the
    derived series. Do not edit the trend package.
-   `TestListDeltaPositive` — S18. Positive delta.
-   `TestListDeltaNegative` — S18. Negative delta.
-   `TestListDeltaZero` — S18. Zero delta.
-   `TestListDeltaNOCOLORKeepsSign` — S18. Sign survives
    `NO_COLOR`.
-   `TestDeltaColumnAligns` — S18. Delta column lines up.
-   `TestMonthDeltaMatchesList` — S18. Same delta on the sheet.
-   `TestVisPadIgnoresANSI` — S18. `lipgloss.Width` stays
    display cells. Coloured and plain pads match.

New tests:

-   `TestSpaceBarStaysSpace` — S4–S9. One test, six subtests, v2
    space bar (printed name `"space"`, code and text are a space).
    Do not weaken an assertion to match a wrong key.
-   `TestModuleRequiresCharmV2` — S1 and FR1. `go 1.27.0`, the three
    v2 modules, no v1 Charm modules, no ntcharts import. Scan
    `go.mod` plus `.go` files under `cmd/` and `internal/` only.

## Changelog

Under the existing `## Unreleased` / `### Changed` section, add this
bullet and do not invent a pull-request number. Integration adds
`(#N)` when the PR exists.

-   TUI uses Charm v2 (Bubble Tea, Lip Gloss, and bubbles). Screens,
    charts, and the monthly PDF behave as before.
