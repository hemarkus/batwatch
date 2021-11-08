package source

import "github.com/distatus/battery"

// Battery wraps battery.Battery to provide percentage
type Battery struct {
	*battery.Battery
}

// Percentage returns the battery's load in percent
func (b *Battery) Percentage() float64 {
	return b.Current * 100 / b.Design
}
