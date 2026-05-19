// ===================
// © AngelaMos | 2026
// app.go
// ===================

package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CarterPerez-dev/test-audit/internal/audit"
	"github.com/CarterPerez-dev/test-audit/internal/parse"
)

const (
	exitOK    = 0
	exitFail  = 1
	exitUsage = 2
)

var version = "dev"

const usage = `test-audit — deterministic certification practice-test auditor

Usage:
  test-audit FILE [FILE...] [flags]

Reads a db.tests.insertOne({...}) test file and writes {stem}_audit.json
next to it (auditor.md schema). Deterministic only — semantic checks are
intentionally out of scope and disclosed in the report summary.

Flags:
  -o, --out DIR        write audit JSON to DIR (default: same dir as input)
      --targets FILE   JSON file overriding distribution targets (default: auditor.md spec)
      --stdout         write report to stdout instead of a file
  -h, --help           show this help
  -v, --version        show version

Examples:
  test-audit cissp_test_6.js
  test-audit *.js -o ./audit/
  test-audit cissp_test_6.js --stdout | jq .overallPass
`

type flags struct {
	out     string
	targets string
	stdout  bool
	help    bool
	version bool
}

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("test-audit", flag.ContinueOnError)
	fl.SetOutput(stderr)
	fl.Usage = func() { fmt.Fprint(stderr, usage) }

	var f flags
	fl.StringVar(&f.out, "o", "", "")
	fl.StringVar(&f.out, "out", "", "")
	fl.StringVar(&f.targets, "targets", "", "")
	fl.BoolVar(&f.stdout, "stdout", false, "")
	fl.BoolVar(&f.help, "h", false, "")
	fl.BoolVar(&f.help, "help", false, "")
	fl.BoolVar(&f.version, "v", false, "")
	fl.BoolVar(&f.version, "version", false, "")

	if err := fl.Parse(reorderFlagsFirst(args)); err != nil {
		return exitUsage
	}
	if f.help {
		fmt.Fprint(stdout, usage)
		return exitOK
	}
	if f.version {
		fmt.Fprintf(stdout, "test-audit %s\n", version)
		return exitOK
	}

	inputs := fl.Args()
	if len(inputs) == 0 {
		fmt.Fprint(stderr, usage)
		return exitUsage
	}

	tg := audit.DefaultTargets()
	if f.targets != "" {
		t, err := loadTargets(f.targets)
		if err != nil {
			fmt.Fprintf(stderr, "test-audit: targets: %v\n", err)
			return exitFail
		}
		tg = t
	}

	hadFail := false
	for _, in := range inputs {
		if err := auditOne(in, f, tg, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "test-audit: %s: %v\n", in, err)
			hadFail = true
		}
	}
	if hadFail {
		return exitFail
	}
	return exitOK
}

func auditOne(
	path string,
	f flags,
	tg audit.Targets,
	stdout, stderr io.Writer,
) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	tst, err := parse.ParseFile(raw)
	if err != nil {
		return err
	}

	rep := audit.Audit(filepath.Base(path), tst, tg)

	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	b = append(b, '\n')

	if f.stdout {
		_, err := stdout.Write(b)
		return err
	}

	outPath := auditPath(path, f.out)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(outPath, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	fmt.Fprintf(
		stderr,
		"test-audit: wrote %s (%d high-priority, %d flagged questions)\n",
		outPath,
		len(rep.CriticalFlags),
		len(rep.QuestionFlags),
	)
	return nil
}

func loadTargets(path string) (audit.Targets, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return audit.Targets{}, err
	}
	t := audit.DefaultTargets()
	if err := json.Unmarshal(raw, &t); err != nil {
		return audit.Targets{}, fmt.Errorf("parse targets json: %w", err)
	}
	return t, nil
}

func auditPath(input, outDir string) string {
	base := filepath.Base(input)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	name := stem + "_audit.json"
	if outDir != "" {
		return filepath.Join(outDir, name)
	}
	return filepath.Join(filepath.Dir(input), name)
}

func reorderFlagsFirst(args []string) []string {
	valueFlags := map[string]bool{
		"-o": true, "--o": true, "-out": true, "--out": true,
		"-targets": true, "--targets": true,
	}
	var fl, positional []string
	i := 0
	for i < len(args) {
		a := args[i]
		if a == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if a == "-" || !strings.HasPrefix(a, "-") {
			positional = append(positional, a)
			i++
			continue
		}
		fl = append(fl, a)
		eq := strings.Contains(a, "=")
		key := a
		if eq {
			key = a[:strings.Index(a, "=")]
		}
		if !eq && valueFlags[key] && i+1 < len(args) {
			fl = append(fl, args[i+1])
			i += 2
			continue
		}
		i++
	}
	return append(fl, positional...)
}
