// Copyright 2014 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License included
// in the file licenses/BSL.txt and at www.mariadb.com/bsl11.
//
// Change Date: 2022-10-01
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt and at
// https://www.apache.org/licenses/LICENSE-2.0

/*
Package server implements the Cockroach storage node. A node
corresponds to a single instance of the cockroach binary, running on a
single physical machine, which exports the "Node" Go RPC service. Each
node multiplexes RPC requests to one or more stores, associated with
physical storage devices.

Package server also provides access to administrative tools via the
command line and also through a REST API.
*/
package server
