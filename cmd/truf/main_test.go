package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		code       int
		exit       bool
		stdout     string
		stderr     string
		db         string
		seed       bool
		wantNoErrs bool
	}{
		{name: "no args boots the TUI", args: nil, db: defaultDBPath, wantNoErrs: true},
		{name: "seed", args: []string{"--seed"}, db: defaultDBPath, seed: true, wantNoErrs: true},
		{name: "db", args: []string{"--db", "/tmp/x.db"}, db: "/tmp/x.db", wantNoErrs: true},
		{name: "db with equals", args: []string{"--db=/tmp/y.db", "--seed"}, db: "/tmp/y.db", seed: true, wantNoErrs: true},
		{name: "help", args: []string{"--help"}, exit: true, stdout: "Usage: truf"},
		{name: "short help", args: []string{"-h"}, exit: true, stdout: "Usage: truf"},
		{name: "version", args: []string{"--version"}, exit: true, stdout: "truf dev\n"},
		{name: "unknown flag", args: []string{"--halp"}, code: 2, exit: true, stderr: "flag provided but not defined: -halp"},
		{name: "stray argument", args: []string{"extra"}, code: 2, exit: true, stderr: "unexpected argument: extra"},
		{name: "db without value", args: []string{"--db"}, code: 2, exit: true, stderr: "flag needs an argument"},
		{name: "empty db", args: []string{"--db="}, code: 2, exit: true, stderr: "--db needs a path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			opts, code, exit := parseFlags(tt.args, &stdout, &stderr)
			if code != tt.code || exit != tt.exit {
				t.Fatalf("code, exit = %d, %v; want %d, %v", code, exit, tt.code, tt.exit)
			}
			if tt.stdout != "" && !strings.Contains(stdout.String(), tt.stdout) {
				t.Errorf("stdout = %q; want it to contain %q", stdout.String(), tt.stdout)
			}
			if tt.stderr != "" {
				if !strings.Contains(stderr.String(), tt.stderr) {
					t.Errorf("stderr = %q; want it to contain %q", stderr.String(), tt.stderr)
				}
				if !strings.Contains(stderr.String(), "Usage: truf") {
					t.Errorf("stderr = %q; want the usage", stderr.String())
				}
			}
			if tt.wantNoErrs && (stdout.Len() > 0 || stderr.Len() > 0) {
				t.Errorf("unexpected output: stdout %q, stderr %q", stdout.String(), stderr.String())
			}
			if tt.exit {
				return
			}
			if opts.db != tt.db || opts.seed != tt.seed {
				t.Errorf("opts = %+v; want db %q seed %v", opts, tt.db, tt.seed)
			}
		})
	}
}
