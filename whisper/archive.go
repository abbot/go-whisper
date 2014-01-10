package whisper

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// ArchiveInfo holds metadata about a single archive within a whisper database
type ArchiveInfo struct {
	Offset          uint32 // The byte offset of the archive within the database
	SecondsPerPoint uint32 // The number of seconds of elapsed time represented by a data point
	Points          uint32 // The number of data points
}

func (a ArchiveInfo) Retention() uint32        { return a.SecondsPerPoint * a.Points }
func (a ArchiveInfo) Size() uint32             { return a.Points * pointSize }
func (a ArchiveInfo) end() uint32              { return a.Offset + a.Size() }
func (a *ArchiveInfo) Write(w io.Writer) error { return binary.Write(w, binary.BigEndian, a) }

type ArchiveInfos []ArchiveInfo

func (a ArchiveInfos) Len() int           { return len(a) }
func (a ArchiveInfos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ArchiveInfos) Less(i, j int) bool { return a[i].SecondsPerPoint < a[j].SecondsPerPoint }

func (a ArchiveInfos) Validate() (err error) {
	if len(a) == 0 {
		return ErrNoArchives
	}
	sort.Sort(a)

	for i, current := range a {
		if i == len(a)-1 {
			break
		}
		next := a[i+1]
		if current.SecondsPerPoint >= next.SecondsPerPoint {
			return ErrDuplicateArchive
		}
		if next.SecondsPerPoint%current.SecondsPerPoint != 0 {
			return ErrUnevenPrecision
		}
		if next.Retention() <= current.Retention() {
			return ErrLowRetention
		}
		if ppc := (next.SecondsPerPoint / current.SecondsPerPoint); current.Points < ppc {
			return ErrInsufficientPoints
		}
	}
	return nil
}

func parseUnit(s string) (multiplier uint32, err error) {
	for k, v := range UnitMultipliers {
		if strings.HasPrefix(k, s) {
			return v, nil
		}
	}
	return 0, errors.New(fmt.Sprintf("Invalid unit '%s'", s))
}

func splitUnit(s string) (digits, unit string) {
	d := true
	for _, r := range s {
		if d {
			if !unicode.IsDigit(r) {
				unit += string(r)
				d = false
			} else {
				digits += string(r)
			}
		} else {
			unit += string(r)
		}
	}
	return
}

func parseUnitString(s string) (result uint32, err error) {
	p, u := splitUnit(s)
	if u == "" {
		u = "seconds"
	}
	if r, err := strconv.ParseUint(p, 10, 32); err != nil {
		return 0, err
	} else {
		m, _ := parseUnit(u)
		return m * uint32(r), nil
	}
}

func isNumber(x string) bool {
	_, err := strconv.ParseInt(x, 10, 32)
	return err == nil
}

/*
parseRetentionDef returns an ArchiveInfo represented by the string.

The string must consist of two numbers, the precision and retention,
separated by a colon (:).

Both the precision and retention strings accept a unit
suffix. Acceptable suffixes are: "s" for second, "m" for minute, "h"
for hour, "d" for day, "w" for week, and "y" for year.

The precision string specifies how large of a time interval is
represented by a single point in the archive.

The retention string specifies how long points are kept in the
archive. If no suffix is given for the retention it is taken to mean a
number of points and not a duration.
*/
func ParseRetentionDef(def string) (a ArchiveInfo, err error) {
	fields := strings.Split(strings.TrimSpace(def), ":")
	if len(fields) != 2 {
		return a, errors.New(fmt.Sprintf("Invalid retention definition: '%s'", def))
	}
	a.SecondsPerPoint, err = parseUnitString(fields[0])
	a.Points, err = parseUnitString(fields[1])
	if !isNumber(fields[1]) {
		a.Points /= a.SecondsPerPoint
	}
	return
}

var (
	UnitMultipliers = map[string]uint32{
		"seconds": 1,
		"minutes": 60,
		"hours":   3600,
		"days":    86400,
		"weeks":   86400 * 7,
		"years":   86400 * 365,
	}
	archiveInfoSize = uint32(binary.Size(ArchiveInfo{}))
)
