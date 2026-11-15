// Copyright (C) 2026 The homewire Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWithoutACommand(t *testing.T) {
	var stderr bytes.Buffer
	if err := run(nil, &stderr); !errors.Is(err, errUsage) {
		t.Fatalf("run() error = %v, want errUsage", err)
	}
	if !strings.Contains(stderr.String(), "serve") {
		t.Errorf("usage does not list serve:\n%s", stderr.String())
	}

	stderr.Reset()
	if err := run([]string{"start"}, &stderr); !errors.Is(err, errUsage) {
		t.Fatalf("run(start) error = %v, want errUsage", err)
	}
	if !strings.Contains(stderr.String(), `unknown command "start"`) {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"-h"}, {"version"}, {"--version"}, {"serve", "-h"}} {
		if err := run(args, &bytes.Buffer{}); err != nil {
			t.Errorf("run(%v) error = %v", args, err)
		}
	}
	if err := run([]string{"serve", "-vendor-id", "0"}, &bytes.Buffer{}); err == nil || errors.Is(err, errUsage) {
		t.Errorf("run(serve -vendor-id 0) error = %v, want an option error", err)
	}
}

// The data directory is ~/.homewire. ~/.go-matter-hub, from before the
// hub was named homewire, is used when it is the only one there, so that
// an existing fabric and its nodes are kept.
func TestDefaultDataDir(t *testing.T) {
	var warned string
	saved := warnLegacyDataDir
	warnLegacyDataDir = func(legacy, _ string) { warned = legacy }
	t.Cleanup(func() { warnLegacyDataDir = saved })

	home := t.TempDir()
	current := filepath.Join(home, defaultDataDirName)
	legacy := filepath.Join(home, legacyDataDirName)

	if got := defaultDataDir(home); got != current || warned != "" {
		t.Fatalf("with neither: %q (warned %q), want %q", got, warned, current)
	}
	if err := os.Mkdir(legacy, 0o700); err != nil {
		t.Fatal(err)
	}
	if got := defaultDataDir(home); got != legacy || warned != legacy {
		t.Fatalf("with the legacy one only: %q (warned %q), want %q", got, warned, legacy)
	}
	warned = ""
	if err := os.Mkdir(current, 0o700); err != nil {
		t.Fatal(err)
	}
	if got := defaultDataDir(home); got != current || warned != "" {
		t.Fatalf("with both: %q (warned %q), want %q", got, warned, current)
	}
	if got, err := resolveDataDir("/srv/hub"); err != nil || got != "/srv/hub" {
		t.Fatalf("resolveDataDir(/srv/hub) = %q, %v", got, err)
	}
}

// Each option of serve can be given in the environment, as the systemd
// unit and a container give them; the command line wins.
func TestOptionsFromTheEnvironment(t *testing.T) {
	t.Setenv("HOMEWIRE_LISTEN", "0.0.0.0:8580")
	t.Setenv("HOMEWIRE_DATA_DIR", "/var/lib/homewire")
	t.Setenv("HOMEWIRE_ALLOWED_HOST", "hub.local, hub.example\tnuc")
	t.Setenv("HOMEWIRE_REQUIRE_AUTH", "true")
	t.Setenv("HOMEWIRE_LOG_LEVEL", "debug")
	o, err := parseServeOptions(nil)
	if err != nil {
		t.Fatal(err)
	}
	if o.listen != "0.0.0.0:8580" || o.dataDir != "/var/lib/homewire" || !o.requireAuth || o.logLevel.String() != "DEBUG" {
		t.Fatalf("options = %+v", o)
	}
	if strings.Join(o.allowedHosts, " ") != "hub.local hub.example nuc" {
		t.Fatalf("allowed hosts = %v", o.allowedHosts)
	}

	o, err = parseServeOptions([]string{"-listen", "127.0.0.1:9000", "-allowed-host", "other"})
	if err != nil || o.listen != "127.0.0.1:9000" || strings.Join(o.allowedHosts, " ") != "other" {
		t.Fatalf("the command line does not win: %+v, %v", o, err)
	}

	t.Setenv("HOMEWIRE_REQUIRE_AUTH", "maybe")
	if _, err := parseServeOptions(nil); err == nil || !strings.Contains(err.Error(), "HOMEWIRE_REQUIRE_AUTH") {
		t.Fatalf("an invalid variable: %v", err)
	}
}

// The commands which use a running hub find the data directory of the
// system service when the user has none, or when only the service runs.
func TestClientDataDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	saved := systemDataDir
	t.Cleanup(func() { systemDataDir = saved })
	systemDataDir = filepath.Join(t.TempDir(), "homewire")

	if got, _ := resolveClientDataDir(""); got != filepath.Join(home, defaultDataDirName) {
		t.Fatalf("with no data directory: %q", got)
	}
	if err := os.Mkdir(systemDataDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if got, _ := resolveClientDataDir(""); got != systemDataDir {
		t.Fatalf("with the service's only: %q", got)
	}
	if err := os.Mkdir(filepath.Join(home, defaultDataDirName), 0o700); err != nil {
		t.Fatal(err)
	}
	if got, _ := resolveClientDataDir(""); got != filepath.Join(home, defaultDataDirName) {
		t.Fatalf("with the user's too: %q", got)
	}
	// A running hub wins: the service's, while the user's is idle.
	if err := os.WriteFile(filepath.Join(systemDataDir, serveInfoFileName), []byte("{}"), 0o640); err != nil {
		t.Fatal(err)
	}
	if got, _ := resolveClientDataDir(""); got != systemDataDir {
		t.Fatalf("with the service running: %q", got)
	}
	if err := os.WriteFile(filepath.Join(home, defaultDataDirName, serveInfoFileName), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, _ := resolveClientDataDir(""); got != filepath.Join(home, defaultDataDirName) {
		t.Fatalf("with both running: %q", got)
	}
	if got, _ := resolveClientDataDir("/srv/hub"); got != "/srv/hub" {
		t.Fatalf("given: %q", got)
	}
	// serve never uses the service's.
	if got, _ := resolveDataDir(""); got != filepath.Join(home, defaultDataDirName) {
		t.Fatalf("serve: %q", got)
	}
}

// A user who may not read the service's data directory is told how to be
// given access, instead of being sent to a data directory of their own.
func TestClientDataDirPermission(t *testing.T) {
	err := permissionHint(systemDataDir, &fs.PathError{Op: "open", Path: filepath.Join(systemDataDir, serveInfoFileName), Err: fs.ErrPermission})
	if err == nil || !errors.Is(err, fs.ErrPermission) || !strings.Contains(err.Error(), "sudo adduser $USER homewire") {
		t.Fatalf("permissionHint = %v", err)
	}
	other := &fs.PathError{Op: "open", Path: "/srv/hub/serve.json", Err: fs.ErrPermission}
	if err := permissionHint("/srv/hub", other); err != error(other) {
		t.Fatalf("permissionHint for another directory = %v", err)
	}

	if os.Geteuid() == 0 {
		t.Skip("root may read any directory")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	saved := systemDataDir
	t.Cleanup(func() { systemDataDir = saved })
	systemDataDir = filepath.Join(t.TempDir(), "homewire")
	if err := os.Mkdir(filepath.Join(home, defaultDataDirName), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(systemDataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(systemDataDir, serveInfoFileName), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(systemDataDir, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(systemDataDir, 0o700) })
	if got, _ := resolveClientDataDir(""); got != systemDataDir {
		t.Fatalf("with the service's not readable: %q, want %q", got, systemDataDir)
	}
	if _, err := readServeInfo(systemDataDir); err == nil || !strings.Contains(err.Error(), "sudo adduser") {
		t.Fatalf("readServeInfo = %v", err)
	}
}
