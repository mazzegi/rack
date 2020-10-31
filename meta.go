package rack

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Meta struct {
	Limit   int
	Skip    int
	Filters []Filter
}

type FilterComparator string

const (
	FilterComparatorEqual   FilterComparator = "eq"
	FilterComparatorLess    FilterComparator = "ls"
	FilterComparatorGreater FilterComparator = "gt"
)

func parseFilterComparator(c string) (FilterComparator, error) {
	switch FilterComparator(c) {
	case FilterComparatorEqual:
		return FilterComparatorEqual, nil
	case FilterComparatorLess:
		return FilterComparatorLess, nil
	case FilterComparatorGreater:
		return FilterComparatorGreater, nil
	default:
		return "", errors.Errorf("invalid comparator")
	}
}

type Filter struct {
	Name       string
	Comparator FilterComparator
	Value      string
}

func parseFilter(s string) (Filter, error) {
	sl := strings.Split(s, ",")
	if len(sl) != 3 {
		return Filter{}, errors.Errorf("invalid filter format")
	}
	c, err := parseFilterComparator(sl[1])
	if err != nil {
		return Filter{}, err
	}
	return Filter{
		Name:       sl[0],
		Comparator: c,
		Value:      sl[2],
	}, nil
}

func extractNumber(vs []string) (int, error) {
	if len(vs) == 0 {
		return 0, errors.Errorf("no values")
	}
	n, err := strconv.ParseInt(vs[0], 10, 64)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func ParseMeta(r *http.Request) (Meta, error) {
	m := Meta{}
	var err error
	for key, values := range r.URL.Query() {
		switch key {
		case "limit":
			if m.Limit, err = extractNumber(values); err != nil {
				return m, err
			}
		case "skip":
			if m.Skip, err = extractNumber(values); err != nil {
				return m, err
			}
		case "filter":
			for _, s := range values {
				f, err := parseFilter(s)
				if err != nil {
					return m, err
				}
				m.Filters = append(m.Filters, f)
			}
		}
	}
	return m, nil
}
