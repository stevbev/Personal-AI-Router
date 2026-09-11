// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package discovery

// retransmitWorkaround reports whether the per-interface unicast re-send of the
// PTR query is required on this platform. Windows refuses to send from a socket
// whose local address is a multicast group, so grandcat/zeroconf's outgoing query
// is silently dropped there and the workaround re-send is the only copy that
// reaches the wire.
func retransmitWorkaround() bool { return true }
