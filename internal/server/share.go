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
	"errors"
	"net/http"

	"github.com/cybergarage/homewire/internal/hub"
)

// handleShareNode opens a commissioning window on a node for another app
// or hub to add it to its own fabric, and answers 201 with the window's
// codes and when it closes. The codes are answered once; the hub does
// not keep them.
func (s *Server) handleShareNode(w http.ResponseWriter, r *http.Request) {
	id, ok := nodeID(w, r)
	if !ok {
		return
	}
	if s.hub == nil {
		writeError(w, http.StatusServiceUnavailable, "not_ready", "the hub has not started")
		return
	}
	share, err := s.hub.ShareNode(r.Context(), id)
	if err != nil {
		writeShareError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, share)
}

// handleStopSharing closes the commissioning window of a node.
func (s *Server) handleStopSharing(w http.ResponseWriter, r *http.Request) {
	id, ok := nodeID(w, r)
	if !ok {
		return
	}
	if s.hub == nil {
		writeError(w, http.StatusServiceUnavailable, "not_ready", "the hub has not started")
		return
	}
	if err := s.hub.StopSharing(r.Context(), id); err != nil {
		writeShareError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeShareError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, hub.ErrUnknownNode):
		writeError(w, http.StatusNotFound, "not_found", "no such node")
	case errors.Is(err, hub.ErrNodeOffline):
		writeError(w, http.StatusConflict, "node_offline", "the node is offline")
	case errors.Is(err, hub.ErrShareBusy):
		writeError(w, http.StatusConflict, "busy", "the device has a commissioning window open already; wait until it closes, or close it")
	case errors.Is(err, hub.ErrShareFailed):
		writeError(w, http.StatusBadGateway, "share_failed", err.Error())
	case errors.Is(err, hub.ErrNotStarted):
		writeError(w, http.StatusServiceUnavailable, "not_ready", "the hub has not started")
	default:
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
	}
}
