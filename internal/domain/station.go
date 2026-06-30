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

// QualityPref controls which StreamVariant SelectVariant returns.
type QualityPref string

const (
	QualityHighest  QualityPref = "highest"
	QualityLowest   QualityPref = "lowest"
	QualityBalanced QualityPref = "balanced" // highest at/below 128 kbps, else lowest
)

// SelectVariant returns the preferred variant. ok is false when there are none.
func (s Station) SelectVariant(p QualityPref) (StreamVariant, bool) {
	if len(s.Variants) == 0 {
		return StreamVariant{}, false
	}
	best := s.Variants[0]
	for _, v := range s.Variants[1:] {
		switch p {
		case QualityLowest:
			if v.Bitrate < best.Bitrate {
				best = v
			}
		case QualityBalanced:
			best = balancedPick(best, v)
		default: // QualityHighest
			if v.Bitrate > best.Bitrate {
				best = v
			}
		}
	}
	return best, true
}

func balancedPick(best, v StreamVariant) StreamVariant {
	const cap = 128
	bestOK := best.Bitrate <= cap
	vOK := v.Bitrate <= cap
	switch {
	case bestOK && vOK:
		if v.Bitrate > best.Bitrate { // highest under the cap
			return v
		}
	case vOK && !bestOK:
		return v // prefer anything within cap over above-cap
	case !bestOK && !vOK:
		if v.Bitrate < best.Bitrate { // both above cap → take the lowest
			return v
		}
	}
	return best
}
