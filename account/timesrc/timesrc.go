package timesrc

import (
	"github.com/beevik/ntp"
	"github.com/pkg/errors"
	"time"
)

// TimeSource defines a source of time.
type TimeSource interface {
	Time() (time.Time, error)
}

// SystemClock is a TimeSource which uses the system clock.
type SystemClock struct{}

func (rc *SystemClock) Time() (time.Time, error) {
	return time.Now().UTC(), nil
}

// NTPTimeSource is a time source which uses a NTP server.
type NTPTimeSource struct {
	server string
}

// NewNTPTimeSource creates a new TimeSource using the given NTP server.
func NewNTPTimeSource(ntpServer string) *NTPTimeSource {
	return &NTPTimeSource{ntpServer}
}

func (ntpTimeSource *NTPTimeSource) Time() (time.Time, error) {
	t, err := ntp.Time(ntpTimeSource.server)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "NTP time source error")
	}
	return t.UTC(), nil
}
