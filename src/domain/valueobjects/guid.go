package valueobjects

import (
	"fmt"
	"regexp"
)

type Guid struct {
	value string
}

var guidRegex = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")

func ParseGuidFromString(value string) (Guid, error) {
	if !guidRegex.MatchString(value) {
		return Guid{}, fmt.Errorf("invalid offer id: %s", value)
	}

	return Guid{value}, nil
}

func (o Guid) String() string {
	if !guidRegex.MatchString(o.value) {
		return ""
	}
	return o.value
}
