package policy

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const script = "scripts/commit-policy.sh"

func root(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	if _, err := os.Stat(filepath.Join(abs, script)); err != nil {
		t.Fatalf("stat %s: %v", script, err)
	}
	return abs
}

func run(t *testing.T, stdin string, args ...string) (string, bool) {
	t.Helper()
	cmd := exec.Command("sh", append([]string{script}, args...)...)
	cmd.Dir = root(t)
	cmd.Stdin = strings.NewReader(stdin)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err == nil
}

func TestMessageMode(t *testing.T) {
	cases := []struct {
		message string
		ok      bool
	}{
		{"ID-91 add commit policy check", true},
		{"ID-91", true},
		{"ID-91 add check\n\nbody text\n", true},
		{"ID-91 and ID-92 in one commit", false},
		{"add commit policy check", false},
		{"ID-40..ID-93 plan the audit gaps", false},
		{"id-91 lowercase id", false},
		{"ID-91 add check\n\nalso ID-92\n", false},
		{"", false},
	}
	for _, c := range cases {
		out, ok := run(t, c.message, "-")
		if ok != c.ok {
			t.Fatalf("message %q: ok = %v, want %v (%s)", c.message, ok, c.ok, out)
		}
	}
}

func TestRangeModeAcceptsCompliantHistory(t *testing.T) {
	out, ok := run(t, "", "1af72b6e4d1461375be754c5dc7aa06ca2d64cdc..HEAD")
	if !ok {
		t.Fatalf("compliant range rejected: %s", out)
	}
}

func TestRangeModeRejectsLegacyCommits(t *testing.T) {
	out, ok := run(t, "", "1993b4d42c0caa4d345328e8dd58e4102079375e..HEAD")
	if ok {
		t.Fatalf("legacy range accepted: %s", out)
	}
	if !strings.Contains(out, "0cf8f42") {
		t.Fatalf("offending commit not named: %s", out)
	}
}

func TestDefaultRangeIsBaselineToHead(t *testing.T) {
	out, ok := run(t, "")
	if !ok {
		t.Fatalf("default range rejected: %s", out)
	}
}

func TestHookDelegatesToTheCheck(t *testing.T) {
	dir := filepath.Join(root(t), "scratch")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, "commit-msg-fixture")
	if err := os.WriteFile(path, []byte("two ids ID-91 ID-92\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
	cmd := exec.Command("sh", "scripts/hooks/commit-msg", path)
	cmd.Dir = root(t)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("hook accepted two ids: %s", out)
	}
}

func TestRangeModeWalksEveryCommit(t *testing.T) {
	rev := exec.Command("git", "rev-list", "--count",
		"1af72b6e4d1461375be754c5dc7aa06ca2d64cdc..HEAD")
	rev.Dir = root(t)
	want, err := rev.Output()
	if err != nil {
		t.Fatalf("rev-list: %v", err)
	}
	out, ok := run(t, "", "1af72b6e4d1461375be754c5dc7aa06ca2d64cdc..HEAD")
	if !ok {
		t.Fatalf("compliant range rejected: %s", out)
	}
	n := strings.TrimSpace(string(want))
	if !strings.Contains(out, n+" commits") {
		t.Fatalf("reported count is not %s: %s", n, out)
	}
}

const audit = "scripts/evidence-audit.sh"

func runAudit(t *testing.T, args ...string) (string, bool) {
	t.Helper()
	cmd := exec.Command("sh", append([]string{audit}, args...)...)
	cmd.Dir = root(t)
	out, err := cmd.CombinedOutput()
	return string(out), err == nil
}

func TestEvidenceAuditAcceptsTheTracker(t *testing.T) {
	out, ok := runAudit(t)
	if !ok {
		t.Fatalf("tracker rejected: %s", out)
	}
}

func TestEvidenceAuditRejectsAnIdWithoutReview(t *testing.T) {
	dir := filepath.Join(root(t), "scratch")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, "tracker-fixture.adoc")
	body := "| Id | M | Status | Id | M | Status\n\n| ID-9999 | M99 | x | ID-9998 | M99 | -\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
	out, ok := runAudit(t, "scratch/tracker-fixture.adoc")
	if ok {
		t.Fatalf("unreviewed id accepted: %s", out)
	}
	if !strings.Contains(out, "ID-9999") {
		t.Fatalf("offending id not named: %s", out)
	}
}

func TestEvidenceAuditRequiresAReviewPerSprint(t *testing.T) {
	out, ok := runAudit(t)
	if !ok {
		t.Fatalf("audit rejected the tracker: %s", out)
	}
	sprints, err := os.ReadFile(filepath.Join(root(t), "doc/Sprints.adoc"))
	if err != nil {
		t.Fatalf("read sprints: %v", err)
	}
	declared := regexp.MustCompile(`(?m)^== (S[0-9]+)`).FindAllStringSubmatch(string(sprints), -1)
	if len(declared) == 0 {
		t.Fatal("no sprint declared")
	}
	for _, m := range declared {
		path := filepath.Join(root(t), "doc/review", m[1]+".adoc")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s has no review: %v", m[1], err)
		}
	}
}

func TestEvidenceAuditRejectsADoneReviewOfABlockedId(t *testing.T) {
	dir := filepath.Join(root(t), "scratch")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, "tracker-blocked.adoc")
	body := "| Id | M | Status | Id | M | Status\n\n| ID-01 | M0 | ! | ID-02 | M0 | x\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
	out, ok := runAudit(t, "scratch/tracker-blocked.adoc")
	if ok {
		t.Fatalf("done review of a blocked id accepted: %s", out)
	}
	if !strings.Contains(out, "ID-01") {
		t.Fatalf("offending id not named: %s", out)
	}
}

func TestEvidenceAuditReadsASingleIdRow(t *testing.T) {
	dir := filepath.Join(root(t), "scratch")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, "tracker-single.adoc")
	body := "| Id | M | Status | Id | M | Status\n\n| ID-9997 | M99 | x |  |  | \n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
	out, ok := runAudit(t, "scratch/tracker-single.adoc")
	if ok {
		t.Fatalf("unreviewed single-id row accepted: %s", out)
	}
	if !strings.Contains(out, "ID-9997") {
		t.Fatalf("single-id row not read: %s", out)
	}
}
