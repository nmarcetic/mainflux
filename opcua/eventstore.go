// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package opcua

import "context"

// EventStore represents event source for things and channels provisioning.
type EventStore interface {
	// Subscribes to a given subject and receives events.
	Subscribe(context.Context, string) error
}
