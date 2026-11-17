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

package hub

import (
	"errors"
	"fmt"
	"time"
)

// ShareTimeout is how long a commissioning window opened to share a node
// stays open: long enough to open another app and type the code in.
const ShareTimeout = 10 * time.Minute

var (
	// ErrShareBusy is a node with a commissioning window open already,
	// or being commissioned by another administrator.
	ErrShareBusy = errors.New("the device has a commissioning window open already")
	// ErrShareFailed is a node which refused, or did not answer, the
	// opening of a commissioning window.
	ErrShareFailed = errors.New("the device did not open a commissioning window")
)

// CommissioningWindow is what a node's commissioning window is opened
// with: the codes another administrator adds the node with.
type CommissioningWindow struct {
	// ManualCode is the 11-digit manual pairing code, as "1234-567-8901".
	ManualCode string
	// QRCode is the QR code payload, as "MT:…".
	QRCode string
}

// Share is a commissioning window open on a node, for another app or hub
// to add the node to its own fabric: multi-admin. Its codes are as secret
// as the window is open; the hub neither logs nor keeps them.
type Share struct {
	ManualCode string    `json:"manualCode"`
	QRCode     string    `json:"qrCode"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

// ShareNode opens a commissioning window on a node, for ShareTimeout, and
// returns its codes. The node stays on the hub's fabric; another app adds
// it to its own with the codes. 5.6.3. Enhanced Commissioning Method.
func (h *Hub) ShareNode(nodeID uint64) (Share, error) {
	w, err := h.onlineWorker(nodeID)
	if err != nil {
		return Share{}, err
	}
	sess := w.currentSession()
	if sess == nil {
		return Share{}, ErrNodeOffline
	}
	opened := time.Now()
	window, err := sess.OpenCommissioningWindow(w.rec.VendorID, w.rec.ProductID, ShareTimeout)
	if err != nil {
		if errors.Is(err, ErrShareBusy) {
			return Share{}, err
		}
		return Share{}, fmt.Errorf("%w: %w", ErrShareFailed, err)
	}
	h.log.Info("commissioning window opened", "node", hexID(nodeID), "timeout", ShareTimeout)
	return Share{ManualCode: window.ManualCode, QRCode: window.QRCode, ExpiresAt: opened.Add(ShareTimeout)}, nil
}

// StopSharing closes the commissioning window of a node. A node with no
// window open is not an error.
func (h *Hub) StopSharing(nodeID uint64) error {
	w, err := h.onlineWorker(nodeID)
	if err != nil {
		return err
	}
	sess := w.currentSession()
	if sess == nil {
		return ErrNodeOffline
	}
	if err := sess.RevokeCommissioning(); err != nil {
		return fmt.Errorf("%w: %w", ErrShareFailed, err)
	}
	h.log.Info("commissioning window closed", "node", hexID(nodeID))
	return nil
}

// onlineWorker returns the worker of a node of the running hub.
func (h *Hub) onlineWorker(nodeID uint64) (*nodeWorker, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.started {
		return nil, ErrNotStarted
	}
	w, ok := h.nodes[nodeID]
	if !ok {
		return nil, ErrUnknownNode
	}
	return w, nil
}
