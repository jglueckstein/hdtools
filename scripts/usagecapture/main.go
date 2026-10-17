// Command usagecapture records one branch's grok usage figures.
//
// mark stores the session and the next turn before the work.
// capture sums that range into a per-PR actuals file. Both
// subcommands print instructions on stderr and exit 2 when a
// person still has to choose a number. They do not read the
// weekly SuperGrok allowance.

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		var need *NeedHuman
		if errors.As(err, &need) {
			fmt.Fprint(os.Stderr, need.Text)
			if !strings.HasSuffix(need.Text, "\n") {
				fmt.Fprintln(os.Stderr)
			}
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return &NeedHuman{Text: usageText()}
	}
	switch args[0] {
	case "mark":
		return cmdMark(args[1:])
	case "capture":
		return cmdCapture(args[1:])
	default:
		return &NeedHuman{Text: usageText()}
	}
}

func usageText() string {
	return strings.TrimSpace(`
usage:
  go run ./scripts/usagecapture mark --session ID [--from-turn N] [--usage-json FILE]
  go run ./scripts/usagecapture capture [--usage-json FILE] [--actuals FILE] [--replace]

mark stores the session id and the next turn for this branch.
capture sums grok usage for that turn range. When a number cannot
be read, the command prints what to run and writes no figures.
The weekly SuperGrok percentage is not a token count.

scripts/usage-capture is the same command. It exits 2 when a person
still has to choose a number. go run reports that case as exit 1.
`) + "\n"
}

func cmdMark(args []string) error {
	session, from, jsonPath, root, err := markFlags(args)
	if err != nil {
		return err
	}
	if session == "" {
		return &NeedHuman{Text: "mark needs the Grok session id.\n\n  go run ./scripts/usagecapture mark --session <session-id>\n\nThe id is the session you are working in. grok usage <session-id> prints it.\nDo not pass the weekly SuperGrok percentage.\n"}
	}
	follow := "go run ./scripts/usagecapture mark --session " + session + " --usage-json /tmp/usage.json"
	if from > 0 {
		follow += fmt.Sprintf(" --from-turn %d", from)
	}
	raw, err := loadUsage(session, jsonPath, follow)
	if err != nil {
		return err
	}
	branch, err := currentBranch(root)
	if err != nil {
		return err
	}
	span, err := MarkFromUsage(raw, branch, from)
	if err != nil {
		return err
	}
	path := spanPath(root, branch)
	if err := WriteSpan(path, span); err != nil {
		return err
	}
	fmt.Printf("marked session %s from turn %d\n%s\n", span.SessionID, span.FromTurn, path)
	return nil
}

func cmdCapture(args []string) error {
	jsonPath, actuals, replace, root, err := captureFlags(args)
	if err != nil {
		return err
	}
	branch, err := currentBranch(root)
	if err != nil {
		return err
	}
	path := spanPath(root, branch)
	span, err := ReadSpan(path)
	if err != nil {
		if os.IsNotExist(err) || os.IsNotExist(unwrap(err)) {
			return &NeedHuman{Text: NoSpanPrompt()}
		}
		// ReadSpan wraps the read error.
		if errors.Is(err, os.ErrNotExist) {
			return &NeedHuman{Text: NoSpanPrompt()}
		}
		return err
	}
	if span.SessionID == "" {
		return &NeedHuman{Text: "The span file has no session id.\nRun mark --session <session-id> again.\n"}
	}
	raw, err := loadUsage(span.SessionID, jsonPath, "go run ./scripts/usagecapture capture --usage-json /tmp/usage.json")
	if err != nil {
		return err
	}
	out, err := Capture(root, path, raw, actuals, replace)
	if err != nil {
		return err
	}
	if out != "" {
		fmt.Print(out)
	}
	return nil
}

func unwrap(err error) error {
	type unwrapper interface{ Unwrap() error }
	u, ok := err.(unwrapper)
	if !ok {
		return err
	}
	return u.Unwrap()
}

// loadUsage reads --usage-json when it is set. Otherwise it runs
// grok usage. followUp is the command to print when that run fails,
// so a mark failure does not tell the user to run capture.
func loadUsage(session, jsonPath, followUp string) ([]byte, error) {
	if jsonPath != "" {
		raw, err := os.ReadFile(jsonPath)
		if err != nil {
			return nil, fmt.Errorf("read usage json: %w", err)
		}
		return raw, nil
	}
	if strings.TrimSpace(session) == "" {
		return nil, &NeedHuman{Text: "A session id is required before grok usage can run.\n"}
	}
	cmd := exec.Command("grok", "usage", session)
	out, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return nil, &NeedHuman{Text: fmt.Sprintf(
			"grok usage did not return JSON for session %s.\nSave the JSON yourself, then re-run the same subcommand:\n\n  grok usage %s > /tmp/usage.json\n  %s\n\nDo not type the weekly SuperGrok percentage into the actuals file.\n",
			session, session, followUp,
		)}
	}
	return out, nil
}

func spanPath(root, branch string) string {
	name := strings.ReplaceAll(branch, "/", "-")
	return filepath.Join(root, "observability", "costs", "spans", name+".json")
}

// currentBranch uses the symbolic ref so a branch with no commits
// yet still has a span path. A detached HEAD has no branch name to
// file the span under, so the command stops and asks.
func currentBranch(root string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", &NeedHuman{Text: "This checkout is not on a named branch, so no span path was chosen.\nCheck out the feature branch and run the command from the repo root.\n"}
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "", &NeedHuman{Text: "This checkout is not on a named branch, so no span path was chosen.\nCheck out the feature branch and run the command from the repo root.\n"}
	}
	return branch, nil
}

func markFlags(args []string) (session string, from int, jsonPath, root string, err error) {
	root, err = os.Getwd()
	if err != nil {
		return "", 0, "", "", fmt.Errorf("getwd: %w", err)
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--session":
			session, i, err = needArg(args, i)
		case "--from-turn":
			var text string
			text, i, err = needArg(args, i)
			if err == nil {
				from, err = atoiPositive(text)
			}
		case "--usage-json":
			jsonPath, i, err = needArg(args, i)
		case "--root":
			root, i, err = needArg(args, i)
		default:
			err = &NeedHuman{Text: usageText()}
		}
		if err != nil {
			return "", 0, "", "", err
		}
	}
	return session, from, jsonPath, root, nil
}

func captureFlags(args []string) (jsonPath, actuals string, replace bool, root string, err error) {
	root, err = os.Getwd()
	if err != nil {
		return "", "", false, "", fmt.Errorf("getwd: %w", err)
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--usage-json":
			jsonPath, i, err = needArg(args, i)
		case "--actuals":
			actuals, i, err = needArg(args, i)
		case "--replace":
			replace = true
		case "--root":
			root, i, err = needArg(args, i)
		default:
			err = &NeedHuman{Text: usageText()}
		}
		if err != nil {
			return "", "", false, "", err
		}
	}
	return jsonPath, actuals, replace, root, nil
}

func needArg(args []string, i int) (string, int, error) {
	if i+1 >= len(args) {
		return "", i, &NeedHuman{Text: args[i] + " needs a value.\n" + usageText()}
	}
	return args[i+1], i + 1, nil
}

func atoiPositive(text string) (int, error) {
	n := 0
	if text == "" {
		return 0, &NeedHuman{Text: "--from-turn needs an integer.\n"}
	}
	for _, r := range text {
		if r < '0' || r > '9' {
			return 0, &NeedHuman{Text: "--from-turn needs an integer.\n"}
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}
