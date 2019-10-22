package trix

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"testing"
	"time"
)

func TestParseBool(t *testing.T) {
	ck := func(v interface{}, expected bool, expectedError string) {
		t.Helper()
		actual, err := parseBool(v)
		testError(t, err, expectedError)
		testDeepEqual(t, actual, expected)
	}

	ck("1", true, "")
	ck(1, true, "")
	ck("t", true, "")
	ck("T", true, "")
	ck("true", true, "")
	ck("TRUE", true, "")
	ck(true, true, "")
	ck("on", true, "")
	ck("ON", true, "")
	ck("0", false, "")
	ck(0, false, "")
	ck("f", false, "")
	ck("F", false, "")
	ck("false", false, "")
	ck("FALSE", false, "")
	ck("off", false, "")
	ck("OFF", false, "")

	ck(nil, false, "bad value")
	ck("", false, "bad value")
	ck("untrue", false, "bad value")
}

func TestParseInt(t *testing.T) {
	ck := func(v interface{}, expected int, expectedError string) {
		t.Helper()
		actual, err := parseInt(v)
		testError(t, err, expectedError)
		testDeepEqual(t, actual, expected)
	}

	// from strconv/atoi_test.go
	ck("", 0, `strconv.ParseInt: parsing "": invalid syntax`)
	ck("0", 0, "")
	ck("-0", 0, "")
	ck("1", 1, "")
	ck("-1", -1, "")
	ck("12345", 12345, "")
	ck("-12345", -12345, "")
	ck("012345", 12345, "")
	ck("-012345", -12345, "")
	ck("98765432100", 98765432100, "")
	ck("-98765432100", -98765432100, "")
	ck("9223372036854775807", 1<<63-1, "")
	ck("-9223372036854775807", -(1<<63 - 1), "")
	ck("9223372036854775808", 1<<63-1, `strconv.ParseInt: parsing "9223372036854775808": value out of range`)
	ck("-9223372036854775808", -1<<63, "")
	ck("9223372036854775809", 1<<63-1, `strconv.ParseInt: parsing "9223372036854775809": value out of range`)
	ck("-9223372036854775809", -1<<63, `strconv.ParseInt: parsing "-9223372036854775809": value out of range`)

	ck(nil, 0, `strconv.ParseInt: parsing "<nil>": invalid syntax`)
	ck(0, 0, "")
	ck(999, 999, "")
	ck(math.Pi, 0, `strconv.ParseInt: parsing "3.141592653589793": invalid syntax`)
}

// bunch of classes to mock the filesystem
type tMockFS map[string]*bytes.Buffer
type tMockFile struct{ r io.Reader }

func (f tMockFile) Read(p []byte) (n int, err error) { return f.r.Read(p) }

func (f tMockFile) Close() error                                  { return nil }
func (f tMockFile) ReadAt(p []byte, off int64) (n int, err error) { return 0, nil }
func (f tMockFile) Seek(offset int64, whence int) (int64, error)  { return 0, nil }
func (f tMockFile) Stat() (os.FileInfo, error)                    { return nil, nil }

func (mock tMockFS) Open(name string) (tFile, error) {
	buf, found := mock[name]
	if !found {
		return nil, os.ErrNotExist
	}
	return tMockFile{bufio.NewReader(buf)}, nil
}

func TestInternalMergeFile(t *testing.T) {
	emptyFS := tMockFS{}
	testError(t,
		internalMergeFile(emptyFS, NewNode(""), "missing-file"),
		"file does not exist",
	)

	badIncludeFS := tMockFS{
		"main.conf": bytes.NewBufferString("include missing-file.conf"),
	}
	testError(t,
		internalMergeFile(badIncludeFS, NewNode(""), "main.conf"),
		`main.conf:1: including "missing-file.conf": file does not exist`,
	)

	niceFS := tMockFS{
		"main.conf": bytes.NewBufferString(`
			a=2
			b.c=3
			include other.conf
		`),
		"other.conf": bytes.NewBufferString(`

			# comment
			a=3
		`),
	}
	node := NewNode("")
	testError(t, internalMergeFile(niceFS, node, "main.conf"), "")
	testEqualString(t, node, `{a=3,b={c=3}}`)

	typedFS := tMockFS{
		"main.conf": bytes.NewBufferString(`
			v.s:string=a
			v.i:int=1
			v.f:float=3.14
			v.b:bool=true
			v.d:duration=1h
			v.t:time=1979-12-07
			a.s:[]string=a,b,c
			a.i:[]int=1,2,3
			a.f:[]float=3.14,3.15,3.16
			a.b:[]bool=1,t,true,on,0,f,false,off
			a.d:[]duration=1h,1m,3d,1h2m3s
			a.t:[]time=1979-12-07T00:00:00Z,1979-12-07T00:00:00Z,Fri Dec  7 00:00:00 1979,Fri Dec  7 00:00:00 UTC 1979,Fri Dec 07 00:00:00 +0000 1979,07 Dec 79 00:00 UTC,07 Dec 79 00:00 +0000,Friday\, 07-Dec-79 00:00:00 UTC,Fri\, 07 Dec 1979 00:00:00 UTC,Fri\, 07 Dec 1979 00:00:00 +0000,1979-12-07 00:00:00,1979-12-07
		`),
	}

	root := NewRoot()
	testError(t, internalMergeFile(typedFS, root, "main.conf"), "")
	ck := func(key, expectedType string, expected Value) {
		t.Helper()
		v := root.Get(key)
		testEqualString(t, fmt.Sprintf("%T", v), expectedType)
		testDeepEqual(t, v, expected)
	}

	expectedTime, _ := time.Parse("2006-01-02 15:04:05", "1979-12-07 00:00:00")
	ck("v.s", "string", "a")
	ck("v.i", "int", 1)
	ck("v.f", "float64", 3.14)
	ck("v.b", "bool", true)
	ck("v.d", "time.Duration", time.Hour)
	ck("v.t", "time.Time", expectedTime)
	ck("a.s", "[]string", []string{"a", "b", "c"})
	ck("a.i", "[]int", []int{1, 2, 3})
	ck("a.f", "[]float64", []float64{3.14, 3.15, 3.16})
	ck("a.b", "[]bool", []bool{
		true, true, true, true,
		false, false, false, false,
	})
	ck("a.d", "[]time.Duration", []time.Duration{
		time.Hour, time.Minute,
		3 * 24 * time.Hour,
		time.Hour + 2*time.Minute + 3*time.Second,
	})

	ck("a.t", "[]time.Time", []time.Time{
		expectedTime, expectedTime, expectedTime, expectedTime, expectedTime,
		expectedTime, expectedTime, expectedTime, expectedTime, expectedTime,
		expectedTime, expectedTime,
	})

	testError(t,
		node.MergeReader(bytes.NewBufferString(`bad syntax`), true),
		`line 1: bad format: "bad syntax"`,
	)

	node.MergeReader(bytes.NewBufferString(`
		a=8
		b.d=4
	`), false)
	testEqualString(t, node, `{a=8,b={c=3,d=4}}`)
}

func TestParseJSON(t *testing.T) {
	data := []byte(`
		{"a":1,"b":"lolcats","c":{"d":3.1415},"d":[1,2,3],"e":[1,"two",3.0,true]}
	`)

	node := NewRoot()
	err := json.Unmarshal(data, node)
	testError(t, err, "")
	node.SortRecursively()
	testEqualString(t, node, `{a=1,b=lolcats,c={d=3.1415},d={1=1,2=2,3=3},e={1=1,2=two,3=3,4=true}}`)
	testDeepEqual(t, node.Get("c.d"), 3.1415)
	testDeepEqual(t, node.Get("e.4"), true)
}
