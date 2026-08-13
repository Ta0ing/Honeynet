package sense

import "context"

type packetCapture interface {
	Run(context.Context, func(Probe)) error
	Close() error
}
