package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const twoTurns = `{
  "sessionId": "sess-1",
  "session": {"totalTokens": 999999, "costUsdTicks": 999999},
  "turns": [
    {"turnNumber": 4, "totalTokens": 10, "costUsdTicks": 15000000000},
    {"turnNumber": 5, "totalTokens": 25, "costUsdTicks": 5000000000}
  ]
}`

func TestSumIgnoresSessionTotal(t *testing.T) {
	fig, err := SumUsage([]byte(twoTurns), 4, 5)
	if err != nil {
		t.Fatal(err)
	}
	if fig.TokensTotal != 35 {
		t.Fatalf("tokens = %d, want 35 from the turns only", fig.TokensTotal)
	}
	if fig.CostUSD != "2" {
		t.Fatalf("cost = %q, want 2", fig.CostUSD)
	}
	if fig.Source != "human-supplied" {
		t.Fatalf("source = %q", fig.Source)
	}
	for _, stage := range fig.Stages {
		if stage.Tokens != "unavailable" {
			t.Fatalf("stage %s = %s", stage.Name, stage.Tokens)
		}
	}
}

func TestSumLeavesCostUnavailableWhenTicksMissing(t *testing.T) {
	raw := `{
	  "sessionId": "sess-1",
	  "session": {"totalTokens": 1},
	  "turns": [
	    {"turnNumber": 1, "totalTokens": 10, "costUsdTicks": 10000000000},
	    {"turnNumber": 2, "totalTokens": 5}
	  ]
	}`
	fig, err := SumUsage([]byte(raw), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if fig.TokensTotal != 15 {
		t.Fatalf("tokens = %d", fig.TokensTotal)
	}
	if fig.CostUSD != "unavailable" {
		t.Fatalf("cost = %q, want unavailable", fig.CostUSD)
	}
}

func TestSumRejectsTurnWithoutTokens(t *testing.T) {
	raw := `{
	  "sessionId": "sess-1",
	  "turns": [
	    {"turnNumber": 1, "costUsdTicks": 1},
	    {"turnNumber": 2, "totalTokens": 5, "costUsdTicks": 1}
	  ]
	}`
	_, err := SumUsage([]byte(raw), 1, 2)
	need, ok := err.(*NeedHuman)
	if !ok {
		t.Fatalf("err = %v, want NeedHuman", err)
	}
	if !strings.Contains(need.Text, "turn 1") || !strings.Contains(need.Text, "totalTokens") {
		t.Fatalf("prompt = %q", need.Text)
	}
}

func TestSumEmptyRangePrompts(t *testing.T) {
	_, err := SumUsage([]byte(twoTurns), 9, 9)
	need, ok := err.(*NeedHuman)
	if !ok {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(need.Text, "999999") {
		t.Fatal("empty range must not offer the session total")
	}
	if !strings.Contains(need.Text, "from-turn") {
		t.Fatalf("prompt = %q", need.Text)
	}
}

func TestNoSpanPromptForbidsWeeklyPercentage(t *testing.T) {
	text := NoSpanPrompt()
	for _, want := range []string{
		"go run ./scripts/usagecapture mark --session",
		"--from-turn",
		"grok usage",
		"weekly SuperGrok percentage",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
	if strings.Contains(text, "tokens_total:") {
		t.Fatal("no-span prompt must not print figures")
	}
}

func TestMarkFromUsageUsesNextTurn(t *testing.T) {
	span, err := MarkFromUsage([]byte(twoTurns), "chore/usage-capture", 0)
	if err != nil {
		t.Fatal(err)
	}
	if span.SessionID != "sess-1" || span.FromTurn != 6 || span.ThroughTurn != 0 {
		t.Fatalf("span = %+v", span)
	}
	explicit, err := MarkFromUsage([]byte(twoTurns), "chore/usage-capture", 4)
	if err != nil {
		t.Fatal(err)
	}
	if explicit.FromTurn != 4 {
		t.Fatalf("from = %d", explicit.FromTurn)
	}
}

func TestApplyFillsFiguresAndFreezesRange(t *testing.T) {
	root := t.TempDir()
	spanPath := filepath.Join(root, "observability", "costs", "spans", "chore-usage-capture.json")
	if err := os.MkdirAll(filepath.Dir(spanPath), 0o755); err != nil {
		t.Fatal(err)
	}
	span := Span{SessionID: "sess-1", Branch: "chore/usage-capture", FromTurn: 4}
	if err := WriteSpan(spanPath, span); err != nil {
		t.Fatal(err)
	}
	actuals := filepath.Join(root, "actuals.md")
	body := "---\ntokens_total: unavailable\ncost_usd: unavailable\nfigures_source: unavailable\n---\n\nFigures were not supplied.\n"
	if err := os.WriteFile(actuals, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := Capture(root, spanPath, []byte(twoTurns), actuals, false)
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Fatalf("stdout = %q, want empty when --actuals is set", out)
	}
	got, err := os.ReadFile(actuals)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	for _, want := range []string{
		"tokens_total: 35",
		"cost_usd: 2",
		"figures_source: human-supplied",
		"turns 4 through 5",
		"sess-1",
		"unavailable",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("actuals missing %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "not supplied") {
		t.Fatalf("old body kept:\n%s", text)
	}
	saved, err := ReadSpan(spanPath)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ThroughTurn != 5 {
		t.Fatalf("through = %d, want 5 frozen", saved.ThroughTurn)
	}
	// A later turn must not change the frozen sum.
	later := `{
	  "sessionId": "sess-1",
	  "session": {"totalTokens": 999999},
	  "turns": [
	    {"turnNumber": 4, "totalTokens": 10, "costUsdTicks": 15000000000},
	    {"turnNumber": 5, "totalTokens": 25, "costUsdTicks": 5000000000},
	    {"turnNumber": 6, "totalTokens": 100, "costUsdTicks": 10000000000}
	  ]
	}`
	if _, err := Capture(root, spanPath, []byte(later), actuals, true); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(actuals)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "tokens_total: 35") {
		t.Fatalf("frozen range changed:\n%s", got)
	}
}

func TestCaptureBeforeWorkPromptsAndWritesNothing(t *testing.T) {
	root := t.TempDir()
	spanPath := filepath.Join(root, "span.json")
	if err := WriteSpan(spanPath, Span{SessionID: "sess-1", Branch: "b", FromTurn: 6}); err != nil {
		t.Fatal(err)
	}
	actuals := filepath.Join(root, "actuals.md")
	body := "---\nfigures_source: unavailable\n---\n\nStay.\n"
	if err := os.WriteFile(actuals, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := Capture(root, spanPath, []byte(twoTurns), actuals, false)
	need, ok := err.(*NeedHuman)
	if !ok {
		t.Fatalf("err = %v", err)
	}
	if out != "" {
		t.Fatalf("stdout = %q", out)
	}
	if strings.Contains(need.Text, "999999") || strings.Contains(need.Text, "through 0") {
		t.Fatalf("prompt = %q", need.Text)
	}
	if !strings.Contains(need.Text, "from-turn") || !strings.Contains(need.Text, "turn 6") {
		t.Fatalf("prompt = %q", need.Text)
	}
	got, err := os.ReadFile(actuals)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("file changed:\n%s", got)
	}
	saved, err := ReadSpan(spanPath)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ThroughTurn != 0 {
		t.Fatalf("through = %d, want unset", saved.ThroughTurn)
	}
}

func TestApplyKeepsStructuralFrontmatter(t *testing.T) {
	root := t.TempDir()
	spanPath := filepath.Join(root, "span.json")
	if err := WriteSpan(spanPath, Span{SessionID: "sess-1", Branch: "b", FromTurn: 4}); err != nil {
		t.Fatal(err)
	}
	actuals := filepath.Join(root, "actuals.md")
	body := "" +
		"---\n" +
		"date: 2026-10-17\n" +
		"branch: chore/usage-capture\n" +
		"issue: \"#111\"\n" +
		"stages_run: [implementer]\n" +
		"review_cycles: 1\n" +
		"files_changed: 4\n" +
		"languages: [go, markdown]\n" +
		"tokens_by_stage:\n" +
		"  - stage: implementer\n" +
		"    tokens: unavailable\n" +
		"tokens_total: unavailable\n" +
		"cost_usd: unavailable\n" +
		"figures_source: unavailable\n" +
		"---\n\n" +
		"Figures were not supplied.\n"
	if err := os.WriteFile(actuals, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Capture(root, spanPath, []byte(twoTurns), actuals, false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(actuals)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	for _, want := range []string{
		"date: 2026-10-17",
		"issue: \"#111\"",
		"stages_run: [implementer]",
		"    tokens: unavailable",
		"tokens_total: 35",
		"cost_usd: 2",
		"figures_source: human-supplied",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("actuals missing %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "not supplied") {
		t.Fatalf("old body kept:\n%s", text)
	}
}

func TestApplyRefusesWhenAlreadySupplied(t *testing.T) {
	root := t.TempDir()
	spanPath := filepath.Join(root, "span.json")
	if err := WriteSpan(spanPath, Span{SessionID: "sess-1", Branch: "b", FromTurn: 4}); err != nil {
		t.Fatal(err)
	}
	actuals := filepath.Join(root, "actuals.md")
	body := "---\ntokens_total: 7\ncost_usd: 1\nfigures_source: human-supplied\n---\n\nKept.\n"
	if err := os.WriteFile(actuals, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Capture(root, spanPath, []byte(twoTurns), actuals, false)
	need, ok := err.(*NeedHuman)
	if !ok {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(need.Text, "--replace") {
		t.Fatalf("prompt = %q", need.Text)
	}
	got, err := os.ReadFile(actuals)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("file changed:\n%s", got)
	}
}

func TestCaptureWithoutSpanWritesNothing(t *testing.T) {
	root := t.TempDir()
	actuals := filepath.Join(root, "actuals.md")
	body := "---\nfigures_source: unavailable\n---\n\nStay.\n"
	if err := os.WriteFile(actuals, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := Capture(root, filepath.Join(root, "missing.json"), []byte(twoTurns), actuals, false)
	need, ok := err.(*NeedHuman)
	if !ok {
		t.Fatalf("err = %v", err)
	}
	if out != "" {
		t.Fatalf("stdout = %q", out)
	}
	if !strings.Contains(need.Text, "weekly SuperGrok percentage") {
		t.Fatalf("prompt = %q", need.Text)
	}
	got, err := os.ReadFile(actuals)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("file changed:\n%s", got)
	}
}
