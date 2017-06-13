// Copyright 2017 The Fuchsia Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"syscall/mx"
	"syscall/mx/mxerror"

	"apps/netstack/eth"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/buffer"
	"github.com/google/netstack/tcpip/header"
	"github.com/google/netstack/tcpip/stack"
)

const headerLength = 14

type linkEndpoint struct {
	c        *eth.Client
	linkAddr tcpip.LinkAddress

	vv    buffer.VectorisedView
	views [1]buffer.View
}

func (ep *linkEndpoint) MTU() uint32                    { return uint32(ep.c.MTU) }
func (ep *linkEndpoint) MaxHeaderLength() uint16        { return headerLength }
func (ep *linkEndpoint) LinkAddress() tcpip.LinkAddress { return ep.linkAddr }

func (ep *linkEndpoint) WritePacket(r *stack.Route, hdr *buffer.Prependable, payload buffer.View, protocol tcpip.NetworkProtocolNumber) error {
	ethHdr := hdr.Prepend(headerLength)
	if r.RemoteLinkAddress == "" && r.RemoteAddress == "\xff\xff\xff\xff" {
		r.RemoteLinkAddress = "\xff\xff\xff\xff\xff\xff"
	}
	remoteLinkAddr := r.RemoteLinkAddress
	if header.IsV4MulticastAddress(r.RemoteAddress) {
		// RFC 1112.6.4
		remoteLinkAddr = tcpip.LinkAddress([]byte{0x01, 0x00, 0x5e, r.RemoteAddress[1] & 0x7f, r.RemoteAddress[2], r.RemoteAddress[3]})
	} else if header.IsV6MulticastAddress(r.RemoteAddress) {
		// RFC 2464.7
		remoteLinkAddr = tcpip.LinkAddress([]byte{0x33, 0x33, r.RemoteAddress[12], r.RemoteAddress[13], r.RemoteAddress[14], r.RemoteAddress[15]})
	}

	copy(ethHdr[0:], remoteLinkAddr)
	copy(ethHdr[6:], ep.linkAddr)
	ethHdr[12] = uint8(protocol >> 8)
	ethHdr[13] = uint8(protocol)

	pktlen := len(hdr.UsedBytes()) + len(payload)
	if pktlen < 60 {
		pktlen = 60
	}
	var buf eth.Buffer
	for {
		buf = ep.c.AllocForSend()
		if buf != nil {
			break
		}
		if err := ep.c.WaitSend(); err != nil {
			log.Printf("link: alloc error: %v", err)
			return err
		}
	}
	buf = buf[:pktlen]
	copy(buf, hdr.UsedBytes())
	copy(buf[len(hdr.UsedBytes()):], payload)
	return ep.c.Send(buf)
}

func (ep *linkEndpoint) Attach(dispatcher stack.NetworkDispatcher) {
	go func() {
		if err := ep.dispatch(dispatcher); err != nil {
			log.Printf("dispatch error: %v", err)
		}
	}()
}

func (ep *linkEndpoint) dispatch(d stack.NetworkDispatcher) (err error) {
	for {
		var b eth.Buffer
		for {
			b, err = ep.c.Recv()
			if mxerror.Status(err) != mx.ErrShouldWait {
				break
			}
			ep.c.WaitRecv()
		}
		if err != nil {
			return err
		}
		// TODO: use the eth.Buffer as a buffer.View
		v := make(buffer.View, len(b))
		copy(v, b)
		ep.views[0] = v
		ep.vv.SetViews(ep.views[:])
		ep.vv.SetSize(len(v))
		ep.c.Free(b)

		// TODO: optimization: consider unpacking the destination link addr
		// and checking that we are a destination (including multicast support).

		remoteLinkAddr := tcpip.LinkAddress(v[6:12])
		p := tcpip.NetworkProtocolNumber(uint16(v[12])<<8 | uint16(v[13]))

		ep.vv.TrimFront(headerLength)
		d.DeliverNetworkPacket(ep, remoteLinkAddr, p, &ep.vv)
	}

	return nil
}

func (ep *linkEndpoint) init() error {
	ep.linkAddr = tcpip.LinkAddress(ep.c.MAC[:])
	log.Printf("linkaddr: %v", ep.linkAddr)
	return nil
}

func newLinkEndpoint(c *eth.Client) *linkEndpoint {
	ep := &linkEndpoint{
		c: c,
	}
	return ep
}
