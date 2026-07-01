package domain

// Axis is a browse dimension.
type Axis int

const (
	AxisCountry Axis = iota
	AxisTag
	AxisLanguage
)

func (a Axis) String() string {
	switch a {
	case AxisTag:
		return "tag"
	case AxisLanguage:
		return "language"
	default:
		return "country"
	}
}

// Facet is one browsable value on an Axis with its station count.
type Facet struct {
	Name  string
	Count int
}

// SortKey is a browse-list sort dimension.
type SortKey int

const (
	SortVotes SortKey = iota
	SortName
	SortTrend
)

const sortKeyCount = 3

// Sort is a chosen sort key plus a user reverse toggle.
type Sort struct {
	Key  SortKey
	Flip bool
}

// Descending reports the effective direction after the reverse toggle.
// Defaults: votes and trending descending, name ascending.
func (s Sort) Descending() bool {
	def := s.Key != SortName
	if s.Flip {
		return !def
	}
	return def
}

// Label is a short crumb label, e.g. "votes ↓" or "name ↑".
func (s Sort) Label() string {
	var name string
	switch s.Key {
	case SortName:
		name = "name"
	case SortTrend:
		name = "trending"
	default:
		name = "votes"
	}
	arrow := "↓"
	if !s.Descending() {
		arrow = "↑"
	}
	return name + " " + arrow
}

// Next advances to the next sort key, resetting the reverse toggle.
func (s Sort) Next() Sort { return Sort{Key: (s.Key + 1) % sortKeyCount} }
