// Copyright 2024 Jigsaw Operations LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !linux

package service

import (
	"net"
	"syscall"

	"github.com/Jigsaw-Code/outline-sdk/transport"

	onet "github.com/Jigsaw-Code/outline-ss-server/net"
)

// fwmark can be used in conjunction with other Linux networking features like cgroups, network namespaces, and TC (Traffic Control) for sophisticated network management.
// Value of 0 disables fwmark (SO_MARK) (Linux Only)
func MakeValidatingTCPStreamDialer(targetIPValidator onet.TargetIPValidator, fwmark uint) transport.StreamDialer {
	if fwmark != 0 {
		panic("fwmark is linux-specific feature and should be 0")
	}
	return &transport.TCPDialer{Dialer: net.Dialer{Control: func(network, address string, c syscall.RawConn) error {
		ip, _, _ := net.SplitHostPort(address)
		return targetIPValidator(net.ParseIP(ip))
	}}}
}
