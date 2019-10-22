package trix

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrParse represents a generic error when parsing
	ErrParse = errors.New("bad value")
	// ErrParseDuration represents an error when parsing a duration
	ErrParseDuration = errors.New("bad duration")

	// useful regular expressions
	durationRegex    = regexp.MustCompile(`^(?:\s*(\d+)\s*d(?:ays?)?)?(?:\s*(\d+)\s*h(?:ours?)?)?(?:\s*(\d+)\s*m(?:in(?:ute)?s?)?)?(?:\s*(\d+)\s*s(?:econds?)?)?$`)
	durationRegexHMS = regexp.MustCompile(`^([0-9]{2,10}):([0-9]{2})(?::([0-9]{2}))?$`)
	reDateAgo        = regexp.MustCompile(`^(\d+) (second|minute|hour|day|week|month|semester|year)s? ago$`)
	reDateFromNow    = regexp.MustCompile(`^(\d+) (second|minute|hour|day|week|month|semester|year)s? from (now|today)$`)
	reDateUnit       = regexp.MustCompile(`^(next|prev(?:ious)?) (second|minute|hour|day|week|month|semester|year)$`)

	reParseIgnore  = regexp.MustCompile(`^\s*(#.*)?$`)              // ignore comments and empty lines
	reParseInclude = regexp.MustCompile(`^\s*include ([^\s]+)\s*$`) // include other files

	// regular key/value, optionally typed
	reParseEntry = regexp.MustCompile(`^\s*([^=\s][^=]*?)(?:[:]((?:\[\])?(?:string|int|float|bool|duration|date|time)))?\s*=\s*(.*?)\s*$`)

	knownTimeLayouts = []string{
		time.RFC3339Nano,
		time.RFC3339,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
)

// parseBool parse a string as a bool value, accepting variants like "1", "t" or "on" as true
func parseBool(v interface{}) (bool, error) {
	switch strings.ToLower(fmt.Sprint(v)) {
	case "1":
		return true, nil
	case "t":
		return true, nil
	case "true":
		return true, nil
	case "on":
		return true, nil
	case "0":
		return false, nil
	case "f":
		return false, nil
	case "false":
		return false, nil
	case "off":
		return false, nil
	}
	return false, ErrParse
}

// parseInt parse a string as an int value.
func parseInt(v interface{}) (int, error) {
	i, err := strconv.ParseInt(fmt.Sprint(v), 10, 0)
	return int(i), err
}

// parseDuration parse durations in the form `<days>d<hours>h<minutes>m<seconds>s`,
// "HH:MM" or "HH:MM:SS". This is similar to time.ParseDuration, but accepts
// days for convenience, assuming "normal" 24 hours days.
// Each of the parts may be omitted, but at lease one must be present.
func parseDuration(v interface{}) (time.Duration, error) {
	s := fmt.Sprint(v)
	if s == "" {
		return time.Duration(0), ErrParseDuration
	}

	matches := durationRegexHMS.FindStringSubmatch(s)
	if matches != nil {
		hours, _ := strconv.ParseInt(matches[1], 10, 64)
		minutes, _ := strconv.ParseInt(matches[2], 10, 64)
		seconds, _ := strconv.ParseInt(matches[3], 10, 64)
		return time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes) + time.Second*time.Duration(seconds), nil
	}

	if matches = durationRegex.FindStringSubmatch(s); matches == nil {
		return time.Duration(0), ErrParseDuration
	}

	prs := func(s string) int64 {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0
		}
		return i
	}

	days := prs(matches[1])
	hours := prs(matches[2])
	minutes := prs(matches[3])
	seconds := prs(matches[4])

	hour := int64(time.Hour)
	minute := int64(time.Minute)
	second := int64(time.Second)
	return time.Duration((days*24+hours)*hour + minutes*minute + seconds*second), nil
}

// parseTime parse timestamps in various formats.
// Assume UTC and truncate precision to seconds.
// If none of them work, return an error.
func parseTime(v interface{}) (time.Time, error) {
	s := fmt.Sprint(v)
	for _, layout := range knownTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC().Truncate(time.Second), nil
		}
	}
	return time.Time{}, fmt.Errorf("Bad time format: %s", s)
}

// UnmarshalJSON will parse the JSON data into the node, creating child nodes
// as necessary.
func (node *Node) UnmarshalJSON(b []byte) error {
	var values map[string]interface{}
	if err := json.Unmarshal(b, &values); err != nil {
		return err
	}

	var set func([]string, interface{})
	set = func(keys []string, value interface{}) {
		if asMap, ok := value.(map[string]interface{}); ok {
			for key, value := range asMap {
				set(append(keys, key), value)
			}
		} else if asArray, ok := value.([]interface{}); ok {
			for i, value := range asArray {
				set(append(keys, fmt.Sprint(i+1)), value)
			}
		} else {
			node.SetKey(strings.Join(keys, "."), value)
		}
	}

	set([]string{}, values)
	return nil
}

// MergeReader will read lines entries from the reader, parse them and merge
// entries under the current node. If stopOnErrors is true, whevener a line is
// found that isn't recognized as whitespace (empty lines, comments) or
// a key-value, the parsing stops and an error is returned. If it is false,
// bad lines are simply ignored.
func (node *Node) MergeReader(reader io.Reader, stopOnErrors bool) error {
	scanner := bufio.NewScanner(reader)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		if line := scanner.Text(); reParseIgnore.MatchString(line) {
			continue
		} else if matches := reParseEntry.FindStringSubmatch(line); matches != nil && len(matches) == 4 {
			// regular entry
			value, err := parseValueType(matches[2], matches[3])
			if err != nil {
				return err
			}
			node.SetKey(matches[1], value)
		} else if stopOnErrors {
			// unknown/syntax error
			return fmt.Errorf(`line %d: bad format: "%s"`, lineNumber, line)
		}
	}
	return nil
}

// MergeArgs merge the arguments with the node.
func (node *Node) MergeArgs(args Args) *Node {
	for key, value := range args {
		node.SetKey(key, value)
	}
	return node
}

// tRegularFS implements tfileSystem using the local disk. This is needed
// only to make internalMergeFile testable.
type tRegularFS struct{}
type tFile interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	Stat() (os.FileInfo, error)
}

func (tRegularFS) Open(name string) (tFile, error) { return os.Open(name) }

var regularFS tfileSystem = tRegularFS{}

type tfileSystem interface {
	Open(name string) (tFile, error)
}

func parseValueType(valueType, value string) (Value, error) {
	switch valueType {
	case "string", "":
		return value, nil
	case "[]string":
		return splitEsc(value, ",", `\`), nil

	case "int":
		return parseInt(value)
	case "[]int":
		values := splitEsc(value, ",", `\`)
		slice := make([]int, len(values))
		var err error
		for i, v := range values {
			if slice[i], err = parseInt(v); err != nil {
				return nil, err
			}
		}
		return slice, nil

	case "float":
		return strconv.ParseFloat(value, 64)
	case "[]float":
		values := splitEsc(value, ",", `\`)
		slice := make([]float64, len(values))
		var err error
		for i, v := range values {
			if slice[i], err = strconv.ParseFloat(v, 64); err != nil {
				return nil, err
			}
		}
		return slice, nil

	case "bool":
		return parseBool(value)
	case "[]bool":
		values := splitEsc(value, ",", `\`)
		slice := make([]bool, len(values))
		var err error
		for i, v := range values {
			if slice[i], err = parseBool(v); err != nil {
				return nil, err
			}
		}
		return slice, nil

	case "duration":
		return parseDuration(value)
	case "[]duration":
		values := splitEsc(value, ",", `\`)
		slice := make([]time.Duration, len(values))
		var err error
		for i, v := range values {
			if slice[i], err = parseDuration(v); err != nil {
				return nil, err
			}
		}
		return slice, nil

	case "time", "date":
		return parseTime(value)
	case "[]time", "[]date":
		values := splitEsc(value, ",", `\`)
		slice := make([]time.Time, len(values))
		var err error
		for i, v := range values {
			if slice[i], err = parseTime(v); err != nil {
				return nil, err
			}
		}
		return slice, nil

	default:
		return fmt.Errorf(`Bad type: "%s"`, valueType), nil
	}
}

func internalMergeFile(os tfileSystem, node *Node, filename string) error {
	numFiles := 0

	// load initial file, handle includes
	seenFiles := map[string]bool{}
	var loadFile func(string) error
	loadFile = func(filename string) error {
		// avoid recursive parsing
		fullPath, err := filepath.Abs(filename)
		if err != nil {
			return err
		}
		if seenFiles[fullPath] {
			return nil
		}
		seenFiles[fullPath] = true

		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		// parse the file, add entries to a queue
		numFiles++
		lineNumber := 0
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lineNumber++
			if line := scanner.Text(); reParseIgnore.MatchString(line) {
				// comment/empty lines?
			} else if matches := reParseInclude.FindStringSubmatch(line); matches != nil && len(matches) == 2 {
				// include?
				includeFilename := path.Join(path.Dir(filename), matches[1])
				if err := loadFile(includeFilename); err != nil {
					return fmt.Errorf(`%s:%d: including "%s": %v`, filename, lineNumber, includeFilename, err)
				}
			} else if matches := reParseEntry.FindStringSubmatch(line); matches != nil && len(matches) == 4 {
				// regular entry
				value, err := parseValueType(matches[2], matches[3])
				if err != nil {
					return err
				}

				node.SetKey(matches[1], value)
			} else {
				// unknown/syntax error
				return fmt.Errorf(`%s:%d: bad format: "%s"`, filename, lineNumber, line)
			}
		}
		return nil
	}
	if err := loadFile(filename); err != nil {
		return err
	}

	return nil
}

// MergeFile will load/parsethe specified filename, following these rules:
// - lines started with "#" and lines containing only whitespace are ignored.
// - lines with the format "include filename" will recursively parsethe
//   specified filename; relative paths can be used.
// - lines that have at least one "=" are split into a "key=value" pair.
// - leading and trailing spaces are trimmed from keys and values.
// - remaining lines are considered syntax errors.
// All entries found are added under the current node. This operation is not
// atomic, that is, if an error occurs in the middle of the process the
// original node will be partially updated.
func (node *Node) MergeFile(filename string) error {
	return internalMergeFile(regularFS, node, filename)
}
