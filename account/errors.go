package account

import (
	"github.com/pkg/errors"
)

// ErrEmptyRecipients is returned when the recipient slice is empty.
var ErrEmptyRecipients = errors.New("recipients slice must be of size > 0")
// ErrTimeoutNotSpecified is returned when conditions for a deposit address define no timeout.
var ErrTimeoutNotSpecified = errors.New("conditions must define a timeout")
// ErrTimeoutTooLow is returned when the defined timeout of the conditions for a deposit address is too low.
var ErrTimeoutTooLow = errors.New("conditions must at least define a timeout of over 2 minutes")
// ErrAccountNotRunning is returned when the account isn't running but methods on it are called.
var ErrAccountNotRunning = errors.New("the account is not running")
// ErrInvalidAccountSettings is returned when the settings are inconsistent.
var ErrInvalidAccountSettings = errors.New("invalid account settings")
// ErrTargetAddressIsSpent is returned when an address in the transfer object for a send operation is already spent.
var ErrTargetAddressIsSpent = errors.New("target address is already spent")
