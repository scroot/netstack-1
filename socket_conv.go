// Copyright 2017 The Fuchsia Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"log"

	"github.com/google/netstack/tcpip"
)

func (v *c_mxrio_gai_req) Unpack() (node, service string, hints *c_addrinfo) {
	if v.node_is_null == 0 {
		for i := 0; i < len(v.node); i++ {
			if v.node[i] == 0 {
				node = string(v.node[:i])
				break
			}
		}
	}
	if v.service_is_null == 0 {
		for i := 0; i < len(v.service); i++ {
			if v.service[i] == 0 {
				service = string(v.service[:i])
				break
			}
		}
	}
	if v.hints_is_null == 0 {
		hints = new(c_addrinfo)
		*hints = v.hints
	}
	return node, service, hints
}

func isZeros(buf []byte) bool {
	for i := 0; i < len(buf); i++ {
		if buf[i] != 0 {
			return false
		}
	}
	return true
}

// TODO: create a tcpip.Option type
func (v *c_mxrio_sockopt_req_reply) Unpack() interface{} {
	switch v.level {
	case SOL_SOCKET:
		switch v.optname {
		case SO_ERROR:
			return tcpip.ErrorOption{}
		case SO_REUSEADDR:
			return tcpip.ReuseAddressOption(0) // TODO extract value
		case SO_KEEPALIVE:
		case SO_BROADCAST:
		case SO_DEBUG:
		case SO_SNDBUF:
		case SO_RCVBUF:
		}
		log.Printf("convSockOpt: TODO SOL_SOCKET optname=%d", v.optname)
	case SOL_IP:
		switch v.optname {
		case IP_TOS:
		case IP_TTL:
		case IP_MULTICAST_IF:
		case IP_MULTICAST_TTL:
			if len(v.optval) < 1 {
				log.Printf("sockopt: bad argument to IP_MULTICAST_TTL")
				return nil
			}
			return tcpip.MulticastTTLOption(v.optval[0])
		case IP_MULTICAST_LOOP:
		case IP_ADD_MEMBERSHIP, IP_DROP_MEMBERSHIP:
			mreq := c_ip_mreq{}
			if err := mreq.Decode(v.optval[:]); err != nil {
				log.Printf("sockopt: bad argument to %d", v.optname)
				return nil
			}
			option := tcpip.MembershipOption{
				InterfaceAddr: tcpip.Address(mreq.imr_interface[:]),
				MulticastAddr: tcpip.Address(mreq.imr_multiaddr[:]),
			}
			if v.optname == IP_ADD_MEMBERSHIP {
				return tcpip.AddMembershipOption(option)
			} else {
				return tcpip.RemoveMembershipOption(option)
			}
		}
		log.Printf("convSockOpt: TODO IPPROTO_IP optname=%d", v.optname)
	case SOL_TCP:
		switch v.optname {
		case TCP_NODELAY:
			return tcpip.NoDelayOption(0) // TODO extract value
		case TCP_MAXSEG:
		case TCP_CORK:
		case TCP_KEEPIDLE:
		case TCP_KEEPINTVL:
		case TCP_KEEPCNT:
		case TCP_SYNCNT:
		case TCP_LINGER2:
		case TCP_DEFER_ACCEPT:
		case TCP_WINDOW_CLAMP:
		case TCP_INFO:
		case TCP_QUICKACK:
		}
		log.Printf("convSockOpt: TODO SOL_TCP optname=%d", v.optname)
	}
	return nil
}
