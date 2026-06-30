// Package domain defines the core station model shared across the app.
package domain

// StreamVariant is a single playable stream for a station at one bitrate/codec.
type StreamVariant struct {
	URL     string
	Codec   string
	Bitrate int // kbps; 0 means unknown
	HLS     bool
}

// Station is a logical station that may expose multiple stream variants
// (Radio Browser stores one URL per record; the directory layer groups them).
type Station struct {
	Name     string
	Country  string
	Lat, Lon float64
	Tags     []string
	Homepage string
	Variants []StreamVariant
}
