package account

import (
	"github.com/pkg/errors"
)

var ErrEmptyRecipients = errors.New("recipients slice must be of size > 0")
var ErrTimeoutNotSpecified = errors.New("deposit requests must have a timeout")
var ErrTimeoutTooLow = errors.New("deposit requests must at least have a timeout of >2 minutes")
var ErrAccountNotRunning = errors.New("the account is not running")
var ErrInvalidAccountSettings = errors.New("invalid account settings")
