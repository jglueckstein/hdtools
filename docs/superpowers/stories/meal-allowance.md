---
spec: docs/superpowers/specs/2026-11-20-meal-allowance.md
date: 2026-11-21
mode: spec
cartographer_model: grok-4.7
stories:
  - id: 1
    lens: [forces, alternatives, patterns]
    title: Keep each date's own allowance
    disposition: accepted
    disposition_rationale: >-
      Each meal date keeps only the allowance typed for it, in
      the same SQLite file as the meals. `[` and `]` show the
      date they land on. The number is not a `config.toml` key,
      not a `daily_log` column, and not the chart's Daily
      Deficit. Saving it does not insert a weight-log row and
      does not change a chart caption. The meal screen does not
      read `display_unit`.
  - id: 2
    lens: [forces, alternatives, patterns]
    title: Blank deletes and zero stays set
    disposition: accepted
    disposition_rationale: >-
      After trim, a blank field deletes the allowance and the
      screen shows `no allowance`. It does not show
      `want a number`. `0`, `0.0`, and `+0` stay set, and the
      field then shows `0`. A leading minus, including `-0` and
      `-0.0`, is refused and is not stored as zero. Foods still
      save while the allowance is unset.
  - id: 3
    lens: [defaults, alternatives, consequences]
    title: Round the unrounded difference once
    disposition: accepted
    disposition_rationale: >-
      Remaining is the allowance minus the unrounded calorie
      total, then the calorie line's one-decimal half-even rule
      once. `2.5` against an unrounded total of `1.25` shows
      `remaining 1.2`. A printed `-0.0` becomes `0.0`. The
      how-to states that rule. This line does not subtract the
      painted calorie text.
  - id: 4
    lens: [patterns, defaults, consequences]
    title: Show the value, not the keystrokes
    disposition: accepted
    disposition_rationale: >-
      The editor shows the stored number. Unset opens empty.
      `1.25` opens as `1.25`. `0`, `0.0`, and `+0` open as `0`.
      `.5` opens as `0.5`. `500` opens as `500`. Enter on that
      text unchanged keeps the stored number.
  - id: 5
    lens: [coherence, forces, consequences]
    title: Never show both terms at once
    disposition: accepted
    disposition_rationale: >-
      The idle screen shows the calorie total and the
      comparison, and does not print the allowance. The editor
      replaces the idle body. The line list, the calorie total,
      the macro totals, and the comparison line are absent
      while the field is open. The heading keeps the date.
  - id: 6
    lens: [forces, alternatives, consequences]
    title: Budget calories, not the macro totals
    disposition: accepted
    disposition_rationale: >-
      Only calories have an allowance. The idle screen shows
      the calorie total either way. Protein, carbohydrate, and
      fat day totals appear only when `track_macros` is true,
      and those three have no allowance. Kilojoules are not
      shown. The comparison is the printed sign and the words
      `no allowance`.
  - id: 7
    lens: [patterns, defaults, alternatives]
    title: Enter saves and Tab does not
    disposition: accepted
    disposition_rationale: >-
      `b` on the idle meal screen opens the allowance. `f` on
      the daily list remains the only door. Enter saves without
      Tab. Tab does not save, does not cancel, and does not move
      to another field. Esc writes nothing. Letters, `[`, and
      `]` are text while the field is open. `ctrl+c` still
      quits. The editor help line is `enter save   esc cancel`.
      The food forms keep Tab.
---

## Story #1 — Keep each date's own allowance

**Source:** `docs/superpowers/specs/2026-11-20-meal-allowance.md`
(Decisions 1 and 7, US3, US4, S12–S14, FR5, FR6, Out of scope)
**Lens:** forces / alternatives / patterns
**Refs:** —

**Context.** The allowance is a number the person types for the meal
date on screen. It is stored with that date in the same SQLite file
as the meals. The weight log already has a carried figure, trend, and
a calorie figure derived from it, Daily Deficit. This slice refuses
both as sources for the meal target.

**Forces.** An authored target for one day, against the convenience
of a standing default, against a chart deficit that moves when the
weight trend moves and is not a meal. A meal date must be able to
hold the number even when that calendar day has no weight-log row.
The file that already holds meals can do that without giving
`daily_log` a meal column.

**Options not taken.** Carry the previous meal date's allowance
forward until cleared. Use a `config.toml` default when the date has
none. Prefill the field from the chart's Daily Deficit.

**Choice as written.** Each meal date keeps only the allowance typed
for it. `[` and `]` show the date they land on, including that date's
own allowance or `no allowance`. The number is not a config key, not
a `daily_log` column, and not the deficit. Saving it does not insert
a weight-log row, does not change either chart caption, and does not
cause the meal screen to read `display_unit`. Grams, the calorie
total, the allowance, and remaining stay unscaled when the weight
display is pounds.

**Consequences.** There is no standing default in either place a
default could have hidden. Preferences already stay out of SQLite,
and this spec also refuses a `config.toml` key, so a usual allowance
has nowhere to live until a later slice adds one on purpose. A meal
date and a weight-log row stay independently optional. Daily Deficit
can change formula without rewriting meal targets. `display_unit`
remains a weight-log switch. It does not become a presentation switch
for food mass or calories. Quit and open of the same file is the
persistence check. The file stays mode `0600`, and meals do not
create a second database.

**Pattern.** This looks like an aggregate boundary (Evans,
*Domain-Driven Design*, 2003). The allowance changes with the meal
date. It is not a field of the daily log, not a preference, and not
the chart deficit. That deficit is a different quantity that happens
to be denominated in calories.

## Story #2 — Blank deletes and zero stays set

**Source:** `docs/superpowers/specs/2026-11-20-meal-allowance.md`
(Decisions 3 and 5, S7, S8, S10, S11, S15, FR3, FR4)
**Lens:** forces / alternatives / patterns
**Refs:** —

**Context.** On a food nutrient, empty after trim is a missing
number. It is refused when the field is mandatory, and stored as zero
when `track_macros` is off. Grams refuse zero as well. The allowance
field uses that trimmed numeric shape and then gives empty the
opposite meaning. A stored zero stays a target.

**Forces.** Absence against a magnitude of zero. Zero is a real
target: remaining is minus the calorie total, including
`remaining 0.0` on a day with no foods. Empty means the person has
not set one. That must not look like a zero target, like
`remaining 0.0`, or like the chart deficit. The person also needs a
way back to absence after a number was stored.

**Options not taken.** Treat blank as missing and refuse it with
`want a number`, as a mandatory nutrient does. Store a blank as zero,
as an optional macro does. Show an unset day as `remaining 0.0` and
use the words `no allowance` for nothing.

**Choice as written.** After trim, empty text, including a field of
only spaces, deletes the allowance and returns to `no allowance`
without `want a number`. `0`, `0.0`, and `+0` stay set. The field
then shows `0`, not empty. A leading `-`, including `-0` and `-0.0`,
is refused and is not stored as zero. Foods still save while the
allowance is unset, and a day with no foods can hold one.

**Consequences.** The meal screen has three meanings of empty on a
numeric field: refuse, store zero, and delete. They share a trim and
a numeric shape, so a later shared parser has to take the meaning as
a parameter or it will apply the wrong one. Clearing is Enter on an
empty field, not a separate key. A refused negative leaves the
previous state. Esc after `-0` on an unset day is still
`no allowance`. Esc after a refusal on a stored `500` is still `500`.

**Pattern.** The refused option is Null Object (Woolf, *Pattern
Languages of Program Design* 3, 1997): an unset allowance that
behaves as zero. The spec keeps a distinguished absence, painted as
the words `no allowance`, and uses zero only for a stored zero.

## Story #3 — Round the unrounded difference once

**Source:** `docs/superpowers/specs/2026-11-20-meal-allowance.md`
(Decision 4, S5, S6, S21, FR2, Documentation)
**Lens:** defaults / alternatives / consequences
**Refs:** O1, O2

**Context.** Remaining is the allowance minus the day's calorie
total. The calorie line already paints that kind of sum at one
decimal, half-even. The spec also applies the weight column's
rewrite of a printed `-0.0` to `0.0`. The allowance may hold a
hundredth the idle calorie line does not show.

**Forces.** One displayed precision for every calorie tenth on the
meal screen, against keeping the hundredths of the subtraction,
against agreeing with arithmetic done from the painted calorie line.
The weight delta already split those choices. Subtracting true
kilograms and then converting disagreed with the two painted cells,
and half-away disagreed with `%.1f` at a halfway tenth. Remaining
has the opposite operand problem. The allowance is not itself an
idle tenth.

**Options not taken.** Subtract the painted calorie text from the
allowance. `2.5` minus a painted `1.2` is `1.3`, which S6 forbids.
Round half away from zero, which already disagrees with the weight
column's `%.1f` at a halfway tenth. Show the unrounded difference,
or round the allowance and the total separately and then subtract.

**Choice as written.** The subtraction uses the unrounded total, the
same sum the calorie line is drawn from, then that line's one-decimal
half-even rule once. `1.25` prints `1.2`. `0.25` prints `0.2`. An
allowance of `2.5` against an unrounded total of `1.25` prints
`remaining 1.2`. A printed `-0.0` becomes `0.0`. The how-to states
that rule and the `2.5` against `1.25` case, and it does not tell
the reader to subtract the painted calorie text.

**Consequences.** Idle remaining can never show a hundredth, so it
will not always equal the allowance minus the painted calorie tenth.
That is accepted. O1 records the companion of the `-0.0` clamp: a
difference whose tenth rounds to zero prints `remaining 0.0`,
including a sub-tenth overage, and "over is negative" means the
printed tenth has a minus. This story does not reopen that band.
What it forecloses is a second rounding policy for remaining. The
weight delta subtracts painted tenths on purpose, because both of
its operands are already on screen. This spec does not paint the
allowance as an idle tenth, which is what makes the unrounded
subtraction the coherent rule. A later idle allowance tenth would
reopen it.

**Pattern.** —

**Notes.** Bounded read of `REFLECTION_LOG.md`, last 90 days. The
aggregate has 27 entries, all dated 2026-09-08 through 2026-11-14,
so that window is the whole file. The 2026-09-25 delta-column
fragment is why half-away is named as an option. The spec itself
only names half-even.

## Story #4 — Show the value, not the keystrokes

**Source:** `docs/superpowers/specs/2026-11-20-meal-allowance.md`
(Decisions 4, 5, and 6, S3, S6, S7, S11, S17, FR2)
**Lens:** patterns / defaults / consequences
**Refs:** #3

**Context.** ` 500 `, `+500`, and `500` are the same allowance.
`.5` is valid. `1.25` must survive a later open. The idle tenth of
that value is `1.2` under #3. The field is specified by what it
shows on the way back in, not by echoing the keys.

**Forces.** Numeric identity, so remaining has one value to
subtract, against textual fidelity, so the person sees what they
typed. The idle printer must not be the editor printer. If the
field opened on `1.2`, Enter with no edits would have to choose
between keeping `1.25` and storing the tenth.

**Options not taken.** Store the trimmed text and show it back, so
`+500`, `.5`, and `500.0` remain distinct spellings. Normalize on
save to one decimal, so the editor and the idle line share a
spelling. Always reopen with a fixed one-decimal form, including
`500.0` and `0.0`.

**Choice as written.** The editor shows the stored number, not the
idle tenth and not the original keystrokes. Unset opens empty, not
`0`. `1.25` opens as `1.25`. `0`, `0.0`, and `+0` open as `0`.
`.5` opens as `0.5`. An allowance entered as `500` opens as `500`,
not `500.0`. Enter on that text unchanged keeps the stored number.

**Consequences.** The spelling is not part of the record. A trailing
`.0` or a leading `+` cannot be kept as a reminder. Unchanged Enter
is safe only because the spelling that reopens still parses to the
same number, which is why the field must show `1.25` and not the
idle `1.2`. A later amount parser that needs the original token,
such as a unit word, cannot recover it from this field.

**Pattern.** This looks like a value object (Evans, *Domain-Driven
Design*, 2003). The stored allowance is a number, and the field is
a canonical spelling of that number.

## Story #5 — Never show both terms at once

**Source:** `docs/superpowers/specs/2026-11-20-meal-allowance.md`
(Decisions 3 and 6, S6, S17, FR2, FR8)
**Lens:** coherence / forces / consequences
**Refs:** O2, O4, #3, #4

**Context.** Decision 3 keeps the allowance off the idle screen. The
comparison line is `no allowance`, or `remaining` and its number,
after the totals that are shown. Decision 6, as settled in O4,
replaces the idle body while the field is open. The line list, the
calorie total, the macro totals, and the comparison line are not
shown. The heading keeps the date.

**Forces.** One job per view, matching the add, grams, and food
editors, which already replace the idle body. Against that, a
calculator layout would leave the calorie total up while the target
is typed, and would leave the target up beside remaining. The
allowance is not a form field the way grams are, so the total it is
judged against is not inside the editor. Copying the replace-body
layout therefore hides that total.

**Options not taken.** Leave `calories`, and possibly `remaining` or
`no allowance`, on screen under the field. Print the stored
allowance as its own idle total next to remaining. Edit the
comparison line in place and keep the food list.

**Choice as written.** Idle shows the calorie total and the
comparison, and does not print the allowance. The editor shows the
allowance field and does not show the total or the comparison. Set
opens on the stored number from #4, not on the idle tenth from #3.
The two views are mutually exclusive.

**Consequences.** The three terms of the subtraction are never on
screen together. Checking a stored `2.5` against an unrounded total
of `1.25` means opening the field, which hides both `calories 1.2`
and `remaining 1.2`. O2's how-to states that subtraction and forbids
doing it from the painted calorie text. The idle screen is also
missing the allowance, so the sentence describes the program's
arithmetic, not a sum the person can re-read from the glass. That
is the layout this spec accepts. What it forecloses is an editor
that shows the figure the allowance will be subtracted from.

**Pattern.** Same replace-body mode as the other meal editors: a
full-screen form, not an inline cell. The two decisions cohere
while the idle screen is a result and the editor is the source
number. They pull apart if a later change shows the target and the
total together.

## Story #6 — Budget calories, not the macro totals

**Source:** `docs/superpowers/specs/2026-11-20-meal-allowance.md`
(Decision 2, US1, S1, S2, FR1, Out of scope)
**Lens:** forces / alternatives / consequences
**Refs:** —

**Context.** Protein, carbohydrate, and fat already have day totals,
and only when `track_macros` is true. Omitted or false is the
default, and that default hides those three lines. The food form is
unchanged. It still asks for the three, and a blank one is still
stored as zero when tracking is off. This slice adds a target for
calories only.

**Forces.** A budget the default screen can show, against macro
targets this slice keeps out of scope. Remaining has to sit on the
idle screen whether or not the three totals are there. A macro
allowance would either force those totals on when tracking is off,
or hide a target the person had set.

**Options not taken.** An allowance per macro, shown only when
`track_macros` is true. One rule that every visible nutrient total
grows a remaining line. A kilojoule figure beside calories.

**Choice as written.** Only calories have an allowance. The idle
screen shows the calorie total either way, then the three day totals
only when tracking is on, then the comparison line. There is no
protein, carbohydrate, or fat allowance. Kilojoules are not shown.
The comparison is the printed sign and the words `no allowance`,
not a new color role.

**Consequences.** The comparison line exists under both config
states, and its place on the screen moves when the three totals
appear or disappear. A later macro target cannot reuse "the line
after calories" without a visibility rule this slice never needed.
Color stays a weight-log and chrome concern. Remaining does not
join `[colors]`. The food form's macro fields stay a data-entry
choice, not a budget.

**Pattern.** —

## Story #7 — Enter saves and Tab does not

**Source:** `docs/superpowers/specs/2026-11-20-meal-allowance.md`
(Decision 6, S3, S17, S18, S19, FR8, FR9)
**Lens:** patterns / defaults / alternatives
**Refs:** O3, #2, #5

**Context.** The add form, the grams edit, and the food edit share
a help line that names Tab. On the add form, Tab is how the name
resolves, and Enter will not save a name that has not been Tabbed.
O3 settled `b` as the allowance key on the idle meal screen. The
allowance editor is one field, and #5 replaces the idle body while
it is open.

**Forces.** A one-field edit should commit with Enter, against the
meal-form habit that Tab is part of saving. The key must not become
a second door onto the meal screen. Keys that move a date or quit
must not do that while the field is open. The help line is the only
chrome those rules fit in.

**Options not taken.** Keep `tab next field` on the help line and
make Tab a no-op that still looks like the other forms. Make Tab
the commit key. Require Tab before Enter, as the add form does
after a name change. Put `b` on the daily list as another way in.

**Choice as written.** Idle `b` opens one allowance field for the
current meal date. It is not a second door. `f` on the daily list
remains the only door. The editor help line is
`enter save   esc cancel`. Enter saves without Tab. Tab does not
save, does not cancel, and does not move to another field. Esc
writes nothing. Letters, `[`, and `]` are text and do not quit or
change the date. `ctrl+c` still quits. Idle `a`, Enter, `e`, `x`,
`[`, `]`, Esc, `q`, and `f` stay as they are. The key reference
gains one meal row for `b` and does not gain a list row for `[` or
`]`.

**Consequences.** Tab then Esc from this field has not saved,
including a blank that would have cleared the allowance under #2.
Muscle memory from the add form does not transfer, and the shorter
help line is the notice. Those two keys are the whole mode, because
the editor has replaced the idle body. A later second field would
have to put Tab back on a help line S17 locks as the two-key line.
The letter is not a parameter. The scenarios, the help line, the
key reference, and the how-to all name `b`.

**Pattern.** This looks like State (Gamma, Helm, Johnson, and
Vlissides, *Design Patterns*, 1994). Idle and allowance-edit are
two modes of one screen, and the same keys are commands in one and
characters in the other. Tab is the deliberate gap in that copy of
the form chrome.
