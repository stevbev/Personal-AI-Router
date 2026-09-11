// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package mcast

import (
	"net"
	"testing"
)

// realIfaceV4 builds the same up, multicast, non-loopback IPv4 interface set that
// netmon.Enumerate reports. A Sender can only bind to an address the host actually
// holds, so the socket tests need a real interface and skip where there is none
// (e.g. a locked-down CI sandbox).
func realIfaceV4() (map[int][]net.IP, bool) {
	out := map[int][]net.IP{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, false
	}
	for _, ifi := range ifaces {
		if ifi.Flags&net.FlagUp == 0 || ifi.Flags&net.FlagMulticast == 0 || ifi.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := ifi.Addrs()
		if err != nil {
			continue
		}
		var v4 []net.IP
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				if ip4 := ipnet.IP.To4(); ip4 != nil && !ip4.IsLoopback() {
					v4 = append(v4, ip4)
				}
			}
		}
		if len(v4) > 0 {
			out[ifi.Index] = v4
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func TestSenderBuildsOneSocketPerInterface(t *testing.T) {
	ifaces, ok := realIfaceV4()
	if !ok {
		t.Skip("no up, multicast, non-loopback IPv4 interface in this environment")
	}
	s := New(ifaces)
	defer s.Close()
	if got := len(s.Ifaces()); got != len(ifaces) {
		t.Fatalf("Sender.Ifaces() = %d, want %d (one long-lived socket per interface)", got, len(ifaces))
	}
}

// TestSenderRefreshKeepsAStablePool pins the core of the fix: refreshing over an
// unchanged interface set must not churn the socket count, so the OS firewall
// keeps its flow instead of re-evaluating a fresh one.
func TestSenderRefreshKeepsAStablePool(t *testing.T) {
	ifaces, ok := realIfaceV4()
	if !ok {
		t.Skip("no up, multicast, non-loopback IPv4 interface in this environment")
	}
	s := New(ifaces)
	defer s.Close()
	before := len(s.Ifaces())
	s.Refresh(ifaces)
	if got := len(s.Ifaces()); got != before {
		t.Fatalf("Refresh over an unchanged set changed the pool from %d to %d sockets", before, got)
	}
}

// TestSenderSendMulticastReturnsPerIfaceOutcomes: one outcome per requested
// interface, and the result is whatever the socket reported. The value is not
// asserted — a host whose firewall drops multicast (the very case under fix) must
// not fail the send bookkeeping, only record the outcome.
func TestSenderSendMulticastReturnsPerIfaceOutcomes(t *testing.T) {
	ifaces, ok := realIfaceV4()
	if !ok {
		t.Skip("no up, multicast, non-loopback IPv4 interface in this environment")
	}
	s := New(ifaces)
	defer s.Close()
	target := &net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}
	if got := s.SendMulticast([]byte{0}, target, s.Ifaces()); len(got) != len(ifaces) {
		t.Fatalf("SendMulticast returned %d outcomes, want %d (one per requested interface)", len(got), len(ifaces))
	}
}

// TestSenderSendUnicastUnknownIfaceErrors: replying on an interface with no pooled
// socket is a caller bug and must surface, not silently drop the reply.
func TestSenderSendUnicastUnknownIfaceErrors(t *testing.T) {
	ifaces, ok := realIfaceV4()
	if !ok {
		t.Skip("no up, multicast, non-loopback IPv4 interface in this environment")
	}
	s := New(ifaces)
	defer s.Close()
	if err := s.SendUnicast([]byte{0}, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1}, 49999); err == nil {
		t.Fatal("SendUnicast on an interface with no pooled socket returned nil, want an error")
	}
}
