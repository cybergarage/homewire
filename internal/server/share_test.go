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

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/cybergarage/homewire/internal/hub"
)

func TestShareNode(t *testing.T) {
	s, h := nodesServer(true)
	if res := do(s, call{method: "POST", path: lightPath + "/share"}); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("without the token: status %d", res.StatusCode)
	}
	res := do(s, call{method: "POST", path: lightPath + "/share", headers: bearer})
	var share hub.Share
	if err := json.NewDecoder(res.Body).Decode(&share); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusCreated || share.ManualCode != "3497-011-2332" || share.QRCode == "" || res.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("status %d, %+v, %q", res.StatusCode, share, res.Header.Get("Cache-Control"))
	}
	if res := do(s, call{method: "DELETE", path: lightPath + "/share", headers: bearer}); res.StatusCode != http.StatusNoContent {
		t.Fatalf("stop: status %d", res.StatusCode)
	}
	h.mu.Lock()
	if len(h.shared) != 1 || h.shared[0] != lightID || len(h.unshared) != 1 {
		t.Fatalf("shared %v, unshared %v", h.shared, h.unshared)
	}
	h.mu.Unlock()

	for err, want := range map[error]int{
		hub.ErrUnknownNode: http.StatusNotFound,
		hub.ErrNodeOffline: http.StatusConflict,
		hub.ErrShareBusy:   http.StatusConflict,
		fmt.Errorf("%w: timeout", hub.ErrShareFailed): http.StatusBadGateway,
		hub.ErrNotStarted: http.StatusServiceUnavailable,
	} {
		h.mu.Lock()
		h.shareErr = err
		h.mu.Unlock()
		if res := do(s, call{method: "POST", path: lightPath + "/share", headers: bearer}); res.StatusCode != want {
			t.Fatalf("%v: status %d, want %d", err, res.StatusCode, want)
		}
		if res := do(s, call{method: "DELETE", path: lightPath + "/share", headers: bearer}); res.StatusCode != want {
			t.Fatalf("stop, %v: status %d, want %d", err, res.StatusCode, want)
		}
	}
}
