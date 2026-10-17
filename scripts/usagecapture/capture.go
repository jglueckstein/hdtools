// usagecapture reads `grok usage` JSON for one branch.
//
// A Grok session in this repo outlives a single pull request, so the
// session total is never the task. A span file names the first turn
// that belongs to the branch, and the first successful capture
// freezes the last turn. The weekly SuperGrok percentage is not in
// the JSON and is never turned into a token count or a dollar rate.
// A turn does not name a pipeline stage, so those figures stay
// unavailable. When the span or a turn's tokens are missing, the
// command prints instructions and writes nothing.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// stageNames are the per-PR actuals labels. None of them can be
// filled from a usage turn.
var stageNames = []string{
	"spec-writer",
	"tdd-agent",
	"implementer",
	"code-reviewer",
	"integration-agent",
}

// NeedHuman means the command stopped because a person has to
// choose a value the JSON does not carry.
type NeedHuman struct {
	Text string
}

// Error is the instruction a person still has to follow.
func (e *NeedHuman) Error() string { return e.Text }

// Span is the committed record of which turns belong to a branch.
type Span struct {
	SessionID   string `json:"session_id"`
	Branch      string `json:"branch"`
	FromTurn    int    `json:"from_turn"`
	ThroughTurn int    `json:"through_turn"`
}

// StageFigure is one pipeline stage. Tokens is "unavailable" or a
// decimal integer string.
type StageFigure struct {
	Name   string
	Tokens string
}

// Figures is the token and cost block of one per-PR actuals file.
type Figures struct {
	TokensTotal int64
	CostUSD     string
	Source      string
	Stages      []StageFigure
}

type usageDoc struct {
	SessionID string `json:"sessionId"`
	Session   struct {
		TotalTokens int64 `json:"totalTokens"`
	} `json:"session"`
	Turns []usageTurn `json:"turns"`
}

type usageTurn struct {
	TurnNumber   int    `json:"turnNumber"`
	TotalTokens  *int64 `json:"totalTokens"`
	CostUsdTicks *int64 `json:"costUsdTicks"`
}

// NoSpanPrompt tells the user how to mark a turn range. It does not
// print figures.
func NoSpanPrompt() string {
	return strings.TrimSpace(`
Need a turn range before token figures can be written.
This Grok session is shared across tasks, so its session total is not this pull request.

Mark the start before the work, from the repo root:

  go run ./scripts/usagecapture mark --session <session-id>

That stores the session id and the next turn in
observability/costs/spans/<branch>.json.

If the work already ran, choose the first turn yourself:

  go run ./scripts/usagecapture mark --session <session-id> --from-turn N

Find N with:

  grok usage <session-id>

N is the first turn whose endedAt is during this branch.
Do not paste the weekly SuperGrok percentage. It is not a token count.
`) + "\n"
}

// MarkFromUsage builds a span. fromTurn 0 means one past the last
// turn in the JSON, or 1 when the session has no turns yet.
func MarkFromUsage(raw []byte, branch string, fromTurn int) (Span, error) {
	doc, err := parseUsage(raw)
	if err != nil {
		return Span{}, err
	}
	if doc.SessionID == "" {
		return Span{}, &NeedHuman{Text: "The usage JSON has no sessionId.\nRun: grok usage <session-id>\nPass that id to mark --session.\n"}
	}
	if fromTurn == 0 {
		max := 0
		for _, turn := range doc.Turns {
			if turn.TurnNumber > max {
				max = turn.TurnNumber
			}
		}
		fromTurn = max + 1
	}
	if fromTurn < 1 {
		return Span{}, &NeedHuman{Text: "from-turn has to be 1 or greater.\n"}
	}
	return Span{SessionID: doc.SessionID, Branch: branch, FromTurn: fromTurn}, nil
}

// SumUsage adds turns in [fromTurn, throughTurn]. The session total
// is not read.
func SumUsage(raw []byte, fromTurn, throughTurn int) (Figures, error) {
	doc, err := parseUsage(raw)
	if err != nil {
		return Figures{}, err
	}
	fig := Figures{Source: "human-supplied"}
	for _, name := range stageNames {
		fig.Stages = append(fig.Stages, StageFigure{Name: name, Tokens: "unavailable"})
	}
	seen := map[int]bool{}
	var ticks int64
	costKnown := true
	matched := 0
	for _, turn := range doc.Turns {
		if turn.TurnNumber < fromTurn || turn.TurnNumber > throughTurn {
			continue
		}
		if seen[turn.TurnNumber] {
			return Figures{}, &NeedHuman{Text: fmt.Sprintf(
				"turn %d appears twice in the usage JSON.\nA turn number has to be unique before it can be summed.\nRe-run: grok usage <session-id>\n",
				turn.TurnNumber,
			)}
		}
		seen[turn.TurnNumber] = true
		matched++
		if turn.TotalTokens == nil {
			return Figures{}, &NeedHuman{Text: fmt.Sprintf(
				"turn %d is in the capture range and has no totalTokens field.\ngrok usage did not record that turn's tokens. Do not substitute 0 or the session total.\nRe-run: grok usage <session-id>\n",
				turn.TurnNumber,
			)}
		}
		fig.TokensTotal += *turn.TotalTokens
		if turn.CostUsdTicks == nil {
			costKnown = false
			continue
		}
		ticks += *turn.CostUsdTicks
	}
	if matched == 0 {
		return Figures{}, &NeedHuman{Text: fmt.Sprintf(
			"No turns from %d through %d are in this usage JSON.\nThe session total was not used.\nRe-mark the turn that belongs to this branch:\n\n  go run ./scripts/usagecapture mark --session <session-id> --from-turn N\n\nN is the first turn whose endedAt is during this branch.\nFind it with: grok usage <session-id>\nDo not paste the weekly SuperGrok percentage.\n",
			fromTurn, throughTurn,
		)}
	}
	if costKnown {
		fig.CostUSD = formatUSD(ticks)
	} else {
		fig.CostUSD = "unavailable"
	}
	return fig, nil
}

func parseUsage(raw []byte) (usageDoc, error) {
	var doc usageDoc
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&doc); err != nil {
		return usageDoc{}, &NeedHuman{Text: "grok usage did not return readable JSON.\nRun: grok usage <session-id>\nPass that output with --usage-json, or fix the session id.\nDo not type the weekly SuperGrok percentage into the actuals file.\n"}
	}
	return doc, nil
}

func formatUSD(ticks int64) string {
	const scale int64 = 10000000000
	sign := ""
	if ticks < 0 {
		sign = "-"
		ticks = -ticks
	}
	whole := ticks / scale
	frac := ticks % scale
	if frac == 0 {
		return sign + fmt.Sprintf("%d", whole)
	}
	s := fmt.Sprintf("%s%d.%010d", sign, whole, frac)
	return strings.TrimRight(s, "0")
}

// WriteSpan stores a span as indented JSON.
func WriteSpan(path string, span Span) error {
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return fmt.Errorf("make span dir: %w", err)
	}
	raw, err := json.MarshalIndent(span, "", "  ")
	if err != nil {
		return fmt.Errorf("encode span: %w", err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("write span: %w", err)
	}
	return nil
}

// ReadSpan loads a span file.
func ReadSpan(path string) (Span, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Span{}, fmt.Errorf("read span: %w", err)
	}
	var span Span
	if err := json.Unmarshal(raw, &span); err != nil {
		return Span{}, fmt.Errorf("parse span: %w", err)
	}
	return span, nil
}

func dirOf(path string) string {
	i := strings.LastIndex(path, string(os.PathSeparator))
	if i < 0 {
		return "."
	}
	return path[:i]
}

// Capture sums the span's turns. actuals empty means return the YAML
// fragment on stdout. A missing span or a refused write returns
// NeedHuman and leaves the actuals file untouched.
func Capture(root string, spanPath string, usageJSON []byte, actuals string, replace bool) (string, error) {
	_ = root
	if _, err := os.Stat(spanPath); err != nil {
		if os.IsNotExist(err) {
			return "", &NeedHuman{Text: NoSpanPrompt()}
		}
		return "", fmt.Errorf("stat span: %w", err)
	}
	if actuals != "" {
		if _, err := os.Stat(actuals); err != nil {
			if os.IsNotExist(err) {
				return "", &NeedHuman{Text: "No actuals file exists yet.\nWrite the structural per-PR record first (branch, issue, stages, files changed), with the token fields unavailable.\nThen run:\n\n  go run ./scripts/usagecapture capture --actuals " + actuals + "\n\nDo not paste the weekly SuperGrok percentage into that file.\n"}
			}
			return "", fmt.Errorf("stat actuals: %w", err)
		}
		raw, err := os.ReadFile(actuals)
		if err != nil {
			return "", fmt.Errorf("read actuals: %w", err)
		}
		if !replace && strings.Contains(string(raw), "figures_source: human-supplied") {
			return "", &NeedHuman{Text: "This actuals file already has figures_source: human-supplied.\nIt was left unchanged.\nTo overwrite those figures from the marked turn range, run:\n\n  go run ./scripts/usagecapture capture --replace --actuals " + actuals + "\n\nDo not paste the weekly SuperGrok percentage over the existing figures.\n"}
		}
	}
	span, err := ReadSpan(spanPath)
	if err != nil {
		return "", err
	}
	doc, err := parseUsage(usageJSON)
	if err != nil {
		return "", err
	}
	through := span.ThroughTurn
	if through == 0 {
		for _, turn := range doc.Turns {
			if turn.TurnNumber >= span.FromTurn && turn.TurnNumber > through {
				through = turn.TurnNumber
			}
		}
	}
	if through < span.FromTurn {
		// through is still 0 when the marked start is past every
		// turn in the file. Saying "through 0" hides that.
		return "", &NeedHuman{Text: fmt.Sprintf(
			"The span starts at turn %d, and this usage JSON has no turn numbered %d or later.\nThe session total was not used, and no figures were written.\nIf you marked the branch before the work, run capture again after the next turn.\nIf the work already ran, re-mark its first turn:\n\n  go run ./scripts/usagecapture mark --session %s --from-turn N\n\nN is the first turn whose endedAt is during this branch.\nFind it with: grok usage %s\nDo not paste the weekly SuperGrok percentage.\n",
			span.FromTurn, span.FromTurn, span.SessionID, span.SessionID,
		)}
	}
	fig, err := SumUsage(usageJSON, span.FromTurn, through)
	if err != nil {
		return "", err
	}
	span.ThroughTurn = through
	if err := WriteSpan(spanPath, span); err != nil {
		return "", err
	}
	block := disclosure(span, fig)
	if actuals == "" {
		return fragment(fig) + "\n" + block, nil
	}
	if err := writeActuals(actuals, fig, block); err != nil {
		return "", err
	}
	return "", nil
}

func disclosure(span Span, fig Figures) string {
	// The cost sentence stands alone so a long dollar figure does
	// not run into the stage sentence.
	cost := "cost_usd is " + fig.CostUSD + ", from costUsdTicks divided by 1e10."
	if fig.CostUSD == "unavailable" {
		cost = "cost_usd is unavailable because an included turn omitted costUsdTicks."
	}
	return fmt.Sprintf(`Structural fields in the frontmatter were left as they were.
Token and cost figures were read from grok usage for session
%s, turns %d through %d inclusive.
tokens_total is the sum of those turns' totalTokens.
%s
tokens_by_stage stays unavailable because a turn does not name
a pipeline stage. The session total was not used. The weekly
SuperGrok allowance was not used.
`, span.SessionID, span.FromTurn, span.ThroughTurn, cost)
}

func fragment(fig Figures) string {
	var b strings.Builder
	b.WriteString("tokens_by_stage:\n")
	for _, stage := range fig.Stages {
		fmt.Fprintf(&b, "  - stage: %s\n    tokens: %s\n", stage.Name, stage.Tokens)
	}
	fmt.Fprintf(&b, "tokens_total: %d\ncost_usd: %s\nfigures_source: %s\n", fig.TokensTotal, fig.CostUSD, fig.Source)
	return b.String()
}

func writeActuals(path string, fig Figures, block string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read actuals: %w", err)
	}
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") {
		return &NeedHuman{Text: "The actuals file has no YAML frontmatter, so no figures were written.\n"}
	}
	rest := text[4:]
	i := strings.Index(rest, "\n---\n")
	if i < 0 {
		return &NeedHuman{Text: "The actuals file has no closing frontmatter marker, so no figures were written.\n"}
	}
	front := rest[:i]
	front = setFront(front, "tokens_total", fmt.Sprintf("%d", fig.TokensTotal))
	front = setFront(front, "cost_usd", fig.CostUSD)
	front = setFront(front, "figures_source", fig.Source)
	out := "---\n" + front + "\n---\n\n" + block
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		return fmt.Errorf("write actuals: %w", err)
	}
	return nil
}

func setFront(front, key, value string) string {
	prefix := key + ":"
	lines := strings.Split(front, "\n")
	found := false
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), prefix) {
			lines[i] = key + ": " + value
			found = true
		}
	}
	if !found {
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		lines = append(lines, key+": "+value)
	}
	return strings.Join(lines, "\n")
}
