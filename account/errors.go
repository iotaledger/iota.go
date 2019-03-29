package account

import (
	"github.com/pkg/errors"
)

// ErrEmptyRecipients is returned when the recipient slice is empty.
var ErrEmptyRecipients = errors.New("recipients slice must be of size > 0")
// ErrTimeoutNotSpecified is returned when a deposit request defines no timeout.
var ErrTimeoutNotSpecified = errors.New("deposit requests must have a timeout")
// ErrTimeoutTooLow is returned when the defined timeout of the deposit request is too low.
var ErrTimeoutTooLow = errors.New("deposit requests must at least have a timeout of >2 minutes")
// ErrAccountNotRunning is returned when the account isn't running but methods on it are called.
var ErrAccountNotRunning = errors.New("the account is not running")
// ErrInvalidAccountSettings is returned when the settings are inconsistent.
var ErrInvalidAccountSettings = errors.New("invalid account settings")
