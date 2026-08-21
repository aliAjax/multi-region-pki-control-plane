package clock

import "time"

type Clock interface{ Now() time.Time }
type Real struct{}

func (Real) Now() time.Time { return time.Now().UTC() }

type Fixed struct{ T time.Time }

func (f Fixed) Now() time.Time { return f.T }

func (f Fixed) Advance(d time.Duration) Fixed {
	return Fixed{T: f.T.Add(d)}
}
