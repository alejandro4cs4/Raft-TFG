package globals

import (
	"raft-tfg.com/alejandroc/pfslib/settings"
)

const (
	OpenfdsMaxSize int    = 128
	MinioBucket    string = "testbucket"
)

var PfsSettings settings.PfsSettings
var Openfds [OpenfdsMaxSize]*Openfd
