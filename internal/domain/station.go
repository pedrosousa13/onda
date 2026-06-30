// Package domain defines the core station model shared across the app.
package domain

import (
	"strconv"
	"strings"
)

// StreamVariant is a single playable stream for a station at one bitrate/codec.
type StreamVariant struct {
	URL      string
	Codec    string
	Bitrate  int // kbps; 0 means unknown
	HLS      bool
	Lossless bool // true for HiFi/FLAC streams (often reported with bitrate 0)
}

// Quality is a short human label for the stream's quality.
func (v StreamVariant) Quality() string {
	if v.Lossless {
		return "HiFi"
	}
	if v.Bitrate > 0 {
		return strconv.Itoa(v.Bitrate) + "k"
	}
	if c := strings.TrimSpace(v.Codec); c != "" && !strings.EqualFold(c, "unknown") {
		return strings.ToUpper(c)
	}
	return "—"
}

// effBitrate is the value SelectVariant ranks by; lossless sorts above any kbps.
func (v StreamVariant) effBitrate() int {
	if v.Lossless {
		return 9999
	}
	return v.Bitrate
}

// Station is a logical station that may expose multiple stream variants
// (Radio Browser stores one URL per record; the directory layer groups them).
type Station struct {
	Name       string
	Country    string
	Lat, Lon   float64
	Tags       []string
	Homepage   string
	Votes      int // community votes/popularity from Radio Browser
	ClickCount int // recent listens from Radio Browser
	Variants   []StreamVariant
}

// QualityPref controls which StreamVariant SelectVariant returns.
type QualityPref string

const (
	QualityHighest  QualityPref = "highest"
	QualityLowest   QualityPref = "lowest"
	QualityBalanced QualityPref = "balanced" // highest at/below 128 kbps, else lowest
)

// SelectVariant returns the preferred variant. ok is false when there are none.
// Lossless/HiFi streams rank as the highest quality.
func (s Station) SelectVariant(p QualityPref) (StreamVariant, bool) {
	if len(s.Variants) == 0 {
		return StreamVariant{}, false
	}
	best := s.Variants[0]
	for _, v := range s.Variants[1:] {
		switch p {
		case QualityLowest:
			if v.effBitrate() < best.effBitrate() {
				best = v
			}
		case QualityBalanced:
			best = balancedPick(best, v)
		default: // QualityHighest
			if v.effBitrate() > best.effBitrate() {
				best = v
			}
		}
	}
	return best, true
}

func balancedPick(best, v StreamVariant) StreamVariant {
	const cap = 128
	bb, vb := best.effBitrate(), v.effBitrate()
	bestOK := bb <= cap
	vOK := vb <= cap
	switch {
	case bestOK && vOK:
		if vb > bb { // highest under the cap
			return v
		}
	case vOK && !bestOK:
		return v // prefer anything within cap over above-cap
	case !bestOK && !vOK:
		if vb < bb { // both above cap → take the lowest
			return v
		}
	}
	return best
}
