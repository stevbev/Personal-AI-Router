// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package discovery

// retransmitWorkaround reports whether the per-interface unicast re-send of the
// PTR query is required on this platform. On every non-Windows platform
// grandcat/zeroconf's multicast-bound socket transmits fine, so the re-send would
// be a redundant second copy of the same query and is skipped.
func retransmitWorkaround() bool { return false }
