// Copyright 2016 The Fuchsia Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

#ifndef APPS_NETSTACK_PORTS_LWIP_ETHERNETIF_H
#define APPS_NETSTACK_PORTS_LWIP_ETHERNETIF_H

#include "third_party/lwip/src/include/lwip/netif.h"

err_t ethernetif_init(struct netif *netif);

#endif /* APPS_NETSTACK_PORTS_LWIP_ETHERNETIF_H */
