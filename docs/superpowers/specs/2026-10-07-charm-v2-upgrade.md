# Charm v2 upgrade

**Date**: 2026-10-07
**Status**: approved
**Objections**:
[`charm-v2-upgrade.md`](../objections/charm-v2-upgrade.md)
(O1–O3 accepted)
**Backlog**: [`idea.md`](../../../idea.md) (TUI)
**Slice**: S1 of
[`charm-v2-then-online-charts.md`](../slices/charm-v2-then-online-charts.md)

The TUI moves to Charm v2. The three libraries it already uses move
together: Bubble Tea, Lip Gloss, and bubbles. Bubbles is in this slice
because the month sheet and the day form both use its text field.
Leaving it behind would be a file cut, not a separate decision.

The daily list, the month sheet, the day form, the monthly chart, the
long-term chart, and the monthly PDF keep the behaviour they have now.
The on-screen charts stay the Braille-cell charts already shipped.
Nothing visible is supposed to move except the libraries underneath.

The existing test suite is the regression net. Scenarios S2, S3, and
S10 through S18 name behaviour the program already has. They are not
a second product, and they do not ask for new glyphs, new loss
numbers, or new keys. The suite already locks those scenarios, except
list ctrl+c, which is the same quit as `q` (S14), chart and
long-term `q` and ctrl+c, which stay on today's quit branches (S15),
and S16, which is no alternate screen and no mouse. S16 is not a
promise that the wire protocol matches v1. None of those three needs
a new test. Profile-dependent colour (S13, and the colour half of
S10) is judged on the written output, not on the screen model's raw
text. Scenarios S1 and S4 through S9 are the only ones that need new
tests.

Later issues are accepted and out of this slice. The monthly Online
picture is
[#77](https://github.com/jglueckstein/hdtools/issues/77).
The Online caption is
[#79](https://github.com/jglueckstein/hdtools/issues/79).
The long-term Online picture is
[#78](https://github.com/jglueckstein/hdtools/issues/78).

## User stories

### US1 — Keep logging on the upgraded stack

As a person who logs weight in the terminal, I want the daily list,
the month sheet, the day form, and both charts to keep today's keys
and today's screens after the Charm upgrade, so that the app still
works on the v2 stack.

### US2 — Keep the space bar meaning a space

As a person filling a month or a day form, I want the space bar to
keep saving a workout toggle without moving, and to stay a space
inside a text field, so that the upgrade does not turn that key into
something else.

### US3 — Keep the charts and the PDF I already have

As a person who reads the on-screen charts and exports the monthly
PDF, I want the Braille-cell pictures, the pencil loss numbers, and
the PDF page and path to stay as they are, so that a library upgrade
is not a new chart.

### US4 — Leave colour and quitting alone

As a person who uses a 16-color terminal, a theme, or `NO_COLOR`, I
want the colour that reaches the terminal, the quit keys, and the
lack of a mouse or a second screen mode to stay as they are, so that
the upgrade does not change how I read the log or how I leave it.

## Decisions

1.  **Three libraries move together.** Bubble Tea, Lip Gloss, and
    bubbles land in one slice. Bubbles is required because the month
    sheet and the day form use its text field. The modules are
    `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, and
    `charm.land/bubbles/v2`. No patch release is pinned. Resolved:
    bubbles is in scope (O3).
2.  **Screens stay.** The daily list, the month sheet, the day form,
    the monthly chart, the long-term chart, and the monthly PDF keep
    their current behaviour. On-screen charts stay the Braille-cell
    pictures already shipped. Resolved: this slice does not draw a
    new chart.
3.  **ntcharts is absent.** The program does not gain that library
    and does not draw the Hacker's Diet Online picture. Issues #77,
    #79, and #78 are later work.
4.  **The monthly PDF path and page stay.** Where the file is written,
    the page, and what `p` does on each screen stay as they are.
5.  **Weight stays kilograms.** Display units stay config-only. A
    unit change does not rewrite stored history.
6.  **Trend stays the derived series (ApplyTrend),** never a stored
    column. The caption stays the pencil identity: Monthly Loss is
    the first trend minus the last trend of the plotted span, and
    Daily Deficit is that loss in pounds times 3500 divided by the
    days in the span. The current calendar month still ends at
    today. Resolved: the caption formula does not change here. Issue
    #79 changes it later.
7.  **The four long-term windows stay** Quarterly, Semiannual,
    Annual, and Complete. The long-term chart opens on Quarterly.
    `]` cycles Semiannual, Annual, Complete, then Quarterly. `[`
    from Quarterly goes to Complete.
8.  **`[colors]` stays fail-open. `display_unit` stays fail-closed.**
    An invalid or omitted colour is dropped and the built-in default
    is used. An invalid display unit rejects the config. A bad
    colour must not block the log. A bad unit must not store pounds
    as kilograms.
9.  **`NO_COLOR` strips chromatic colour only.** When it is
    non-empty, bold, reverse, and the `>` mark stay. When it is
    empty, chromatic colour stays. The program does not decide this
    by whether a terminal is attached.
10. **16-color is still the floor and the downshift target.** The
    floor is the colour written to the terminal, not the string
    inside the screen model (O1). A hex colour written for a
    truecolor profile is `38;2`. The same hex written for a 16-color
    profile is an indexed colour and is not `38;2`. The model text
    may stay full fidelity. The output writer downsamples once. The
    app does not downsample that text itself, so a sequence is not
    converted twice. `NO_COLOR`, bold, reverse, and `>` stay
    properties of the model text (Decision 9).
11. **The space bar is a space,** not the word "space", and not the
    letters s, p, a, c, e. A library may report that key under the
    name "space"; the program still treats it as the space
    character. On the month sheet, the space bar on the workout
    cell saves the flipped workout and does not move the day or the
    column. On any other month cell that is not being edited, the
    space bar does not start an edit and does not move. One other
    typed character on a non-workout cell starts an edit that shows
    that character. While a month text cell is being edited, the
    space bar inserts a space, leaves the edit open, and does not
    save or move. On the day form, the space bar in a text field
    inserts a space and does not flip workout. The space bar on the
    form's workout field flips yes/no, stays on that field, and
    does not write the log until the form is saved. Resolved: the
    month sheet saves workout in place; the day form only toggles
    until save. Those are the behaviours the screens have now.
12. **No alternate screen and no mouse.** The program is started
    with the screen model alone, as it is today. It does not switch
    the terminal to an alternate screen and it does not turn on
    mouse reporting. Charm v2 enables modifyOtherKeys and Kitty key
    disambiguation for any program that reads keys. This slice does
    not try to turn those off, and it does not promise that the wire
    protocol matches v1 (O2). S16 needs no new test.
13. **Quit keys stay.** From the daily list, `q` and ctrl+c quit.
    On the monthly chart and the long-term chart, `q` and ctrl+c
    quit, and Esc returns to the screen that opened that chart.
    Enter, Tab, Esc, the arrows, `q`, and ctrl+c still match those
    keys. No quit key is added or removed on those screens.

## Acceptance scenarios

November 1990 is the month-sheet fixture when a scenario needs a
date. Weights in the log are kilograms. The display unit is
kilograms unless the scenario says otherwise.

### S1 — Charm v2, and no ntcharts

**Given** the built program
**When** its library requirements and its source imports are read
**Then** it depends on `charm.land/bubbletea/v2`,
`charm.land/lipgloss/v2`, and `charm.land/bubbles/v2`
**And** it does not depend on ntcharts
**And** no source file imports ntcharts or the v1 paths
`github.com/charmbracelet/bubbletea`,
`github.com/charmbracelet/lipgloss`, or
`github.com/charmbracelet/bubbles`

### S2 — The daily list still drives the same screens

**Given** a log with at least one weighed day
**When** the daily list is shown
**Then** the row shows the date, the weight in the display unit, the
derived trend, the delta (displayed weight minus trend), sleep,
steps, and workout
**And** the selected row is marked with `>`
**And** the list opens on the row closest to today, not the first
row
**And** `t` selects that closest row and does not create a today row
**And** `n` opens the day form
**And** `m` opens the month sheet
**And** `c` opens the monthly chart for the selected row's month
**And** `l` opens the long-term chart

### S3 — Month-sheet edit keys stay

**Given** the November 1990 month sheet
**And** a weight cell that is not the first or last day is being
edited with a valid weight
**When** Enter or Down is pressed
**Then** the weight is stored in kilograms and focus moves down one
day in the same column
**When** Up is pressed while editing a valid weight
**Then** the weight is stored and focus moves up one day in the same
column
**When** Left or Right is pressed while editing
**Then** the caret moves inside the text, the cell is not saved, and
the day and column do not change
**When** Tab is pressed while editing
**Then** the cell is saved and focus moves to the next cell
**When** Esc is pressed while editing
**Then** the edit is cancelled and nothing is written
**And** invalid text does not move the focus
**And** Enter or Down on the last day does not leave the month
**And** Up on the first day does not leave the month
**And** when the sheet is not editing, arrows move among cells, Tab
moves to the next cell, and Enter opens the day form

### S4 — Workout space saves and does not move

**Given** the November 1990 month sheet
**And** focus is on the workout cell for a day, not editing
**When** the space bar is pressed
**Then** the stored workout flag flips
**And** focus stays on that same day and that same column
**And** the cell is not left in an edit
**And** the sheet does not move to another screen

### S5 — Space on an idle non-workout cell does not edit

**Given** the month sheet is not editing
**And** focus is on a weight, sleep, steps, or note cell
**When** the space bar is pressed
**Then** no edit starts
**And** the day and column do not change
**And** nothing is written

### S6 — One typed character starts an edit

**Given** the month sheet is not editing
**And** focus is on a non-workout cell
**When** one character other than the space bar is typed (for
example `8`)
**Then** an edit starts on that same cell
**And** the text shows that character and no other
**And** the day and column do not change
**And** nothing is written yet

### S7 — Space in a month text field is a space

**Given** a non-workout month cell is being edited
**When** the space bar is pressed
**Then** the field gains one space character
**And** it does not gain the word "space"
**And** the edit stays open
**And** the day and column do not change
**And** nothing is written yet

### S8 — Space in a day-form text field is a space

**Given** the day form is focused on a text field
**When** the space bar is pressed
**Then** that field gains one space character
**And** it does not gain the word "space"
**And** the workout yes/no does not change
**And** the log is not written

### S9 — Day-form workout space toggles and does not save

**Given** the day form is focused on workout
**When** the space bar is pressed
**Then** the on-screen yes/no flips
**And** focus stays on workout
**And** the log is not written until the form is saved

### S10 — The monthly chart stays the Braille picture

**Given** a month with weighed days
**When** the monthly chart is shown
**Then** the plot is still the Braille-cell chart already shipped,
with the glyphs and title the current chart tests expect
**And** profile-dependent colour is the written output in
Decision 10, still the colours those tests expect
**And** it is not a new picture
**And** the current calendar month ends at today
**And** a past month still runs through its last calendar day
**And** Monthly Loss is the first trend minus the last trend of the
plotted span, in the display unit
**And** Daily Deficit is that loss in pounds times 3500, divided by
the days in the span
**And** those two numbers stay the numbers the current chart tests
expect

### S11 — The long-term windows stay

**Given** the long-term chart
**When** it opens
**Then** the window is Quarterly
**When** `]` is pressed
**Then** the windows cycle Semiannual, Annual, Complete, and back to
Quarterly
**When** `[` is pressed from Quarterly
**Then** the window is Complete
**And** the plot stays the Braille-cell long-term chart already
shipped
**And** there is no fifth window

### S12 — The monthly PDF path and page stay

**Given** the monthly chart
**When** `p` is pressed, or the command writes that chart
**Then** the file is still written by the current PDF rules (data
directory, `pdf_dir`, or `-o`)
**And** the page is the page those rules already write
**And** a failure does not quit, does not claim the file was saved,
and does not write a fallback copy
**And** `p` on the list, the form, the month sheet, or the
long-term chart still writes no PDF

### S13 — Colour and the display unit stay

**Given** a hex weight colour and a truecolor profile
**When** the list is shown
**Then** the colour written for that profile uses `38;2` for the hex
**Given** the same hex and a 16-color profile
**Then** the colour written for that profile is an indexed colour
**And** that written output does not contain `38;2` for that hex
**And** the screen model's own text may still contain
full-fidelity colour
**Given** `NO_COLOR` set to a non-empty value
**Then** chromatic colour is absent from the model text
**And** bold, reverse, and `>` remain in the model text
**Given** `NO_COLOR` empty
**Then** the built-in chromatic colours are present
**Given** an invalid `[colors]` value
**Then** the log still opens and that role uses the built-in default
**Given** an invalid `display_unit`
**Then** the config is rejected
**And** pounds are not stored as kilograms

### S14 — `q` and ctrl+c quit from the list

**Given** the daily list
**When** `q` is pressed
**Then** the program quits
**When** ctrl+c is pressed
**Then** the program quits

### S15 — Chart and long-term quit keys stay

**Given** the monthly chart opened from the list or the month sheet
**When** `q` or ctrl+c is pressed
**Then** the program quits
**When** Esc is pressed
**Then** the screen returns to the one that opened the chart
**Given** the long-term chart opened from the list, the month
sheet, or the monthly chart
**When** `q` or ctrl+c is pressed
**Then** the program quits
**When** Esc is pressed
**Then** the screen returns to the one that opened the long-term
chart

### S16 — No alternate screen and no mouse

**Given** the program starts the TUI
**When** any screen is shown
**Then** the terminal is not switched to an alternate screen
**And** mouse reporting is not turned on
**And** the visible text is the same full text the screens already
show
**And** this scenario does not require the wire protocol to match
v1

### S17 — Kilograms and the derived trend stay

**Given** a weight entered in pounds or stone
**When** it is saved
**Then** the log stores kilograms
**And** the trend is still derived for display
**And** the trend is not a stored column
**And** a backdated edit still recomputes the following trend

### S18 — Delta and column alignment stay

**Given** a weighed day on the list or the month sheet
**When** the row is shown
**Then** the delta is the displayed weight minus the displayed trend
**And** a missing weight has an empty delta
**And** columns line up by display width, including when a cell is
coloured
**And** colour is not the only way to read the delta's sign

## Functional requirements

-   **FR1.** The built program depends on `charm.land/bubbletea/v2`,
    `charm.land/lipgloss/v2`, and `charm.land/bubbles/v2`. It does
    not depend on ntcharts. No source file imports ntcharts or the
    v1 paths `github.com/charmbracelet/bubbletea`,
    `github.com/charmbracelet/lipgloss`, or
    `github.com/charmbracelet/bubbles`.
-   **FR2.** The daily list, the month-sheet edit keys, and the keys
    that open the form, the month sheet, and either chart stay as
    they are. The visible text of each screen (words, `>`, Braille
    glyphs, loss numbers) is unchanged.
-   **FR3.** On the month sheet, the space bar on the workout cell
    saves the flipped workout and does not move the day or the
    column. The space bar on any other cell that is not being
    edited does not start an edit and does not move. One other typed
    character on a non-workout cell starts an edit that contains
    that character. The space bar while a text cell is being edited
    inserts one space, leaves the edit open, and does not save or
    move. The field does not show the word "space".
-   **FR4.** On the day form, the space bar in a text field inserts
    one space and does not flip workout. The space bar on workout
    flips yes/no, stays on that field, and does not write the log
    until the form is saved.
-   **FR5.** The monthly chart stays the Braille-cell chart already
    shipped. The current calendar month ends at today. Monthly Loss
    and Daily Deficit stay the pencil identity in Decision 6, and
    stay the numbers the current chart tests expect.
-   **FR6.** The long-term chart keeps Quarterly, Semiannual,
    Annual, and Complete, cycling as in Decision 7, on the
    Braille-cell picture already shipped.
-   **FR7.** The monthly PDF keeps its current page and its current
    directory rule. `p` on other screens still writes nothing. A
    failed export still does not quit or claim success.
-   **FR8.** `[colors]` stays fail-open and `display_unit` stays
    fail-closed. Non-empty `NO_COLOR` strips chromatic colour only
    from the model text; bold, reverse, and `>` remain there. Empty
    `NO_COLOR` keeps chromatic colour in the model text. 16-color
    remains the floor and the downshift target for the colour
    written out. Hex written for a truecolor profile uses `38;2`.
    The same hex written for a
    16-color profile does not use `38;2`. The app does not
    downsample the model text itself.
-   **FR9.** `q` and ctrl+c quit from the daily list. On the monthly
    chart and the long-term chart, `q` and ctrl+c quit and Esc
    returns to the screen that opened that chart. Enter, Tab, Esc,
    the arrows, `q`, and ctrl+c still match those keys. The program
    does not use an alternate screen and does not enable a mouse
    mode. It does not disable the keyboard disambiguation Charm v2
    turns on for a program that reads keys.
-   **FR10.** Weight is stored in kilograms. Trend stays the derived
    series (ApplyTrend) and is not stored. The delta column stays
    displayed weight minus displayed trend. Columns stay aligned by
    display-cell width when colour is present.
-   **FR11.** The changelog records the upgrade under Unreleased as
    a behaviour-preserving change, with no pull-request number
    invented in advance. User-facing docs under `docs/` are not
    rewritten: the keys and screens did not change.

## Out of scope

-   ntcharts, and the Hacker's Diet Online picture on either chart
    (#77, #78)
-   Replacing the pencil caption with the Online rate (#79)
-   PDF destination changes, filled log sheets, long-term PDFs, and
    blank sheets
-   New daily fields, an exercise rung, or meal planning
-   An alternate screen, mouse reporting, a grid, a legend, or zoom
-   Storing trend, or storing weight in a display unit

## Documentation

No user-facing doc change. The key reference, the config reference,
and the chart how-to already describe these screens. The backlog
change record is the Charm v2 paragraph in `idea.md`. The changelog
bullet is added when the code lands.
