package issue1

import "fmt"

// ErrChannelNotFound is returned when the specified channel does not exist
var ErrChannelNotFound = fmt.Errorf("channel does not exist found")
