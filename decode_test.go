package goblin

import (
	"bytes"
	"encoding/gob"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestGoblin(t *testing.T) {
	type other struct {
		Colour uint
	}

	type bart struct {
		Name    string
		Age     int
		Pimples int64
		Sane    bool
		Lengths []int
		Other   *other
		Height  float64
		Blob    []byte
		Sizes   [3]int
	}

	thing := bart{
		Name:    "goober",
		Age:     19,
		Pimples: 12345678,
		Sane:    false,
		Lengths: []int{8, 1001},
		Other: &other{
			Colour: uint(12345678),
		},
		Height: 14.679,
	}
	thing.Sizes[1] = 25

	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(thing)
	if err != nil {
		t.Fatal("shame", err)
	}

	thing2 := bart{
		Name:    "gopher",
		Age:     49,
		Sane:    true,
		Lengths: []int{0, -11155},
		Blob:    []byte{9, 8, 7, 0x55},
	}
	err = enc.Encode(thing2)
	if err != nil {
		t.Fatal("shame", err)
	}

	// make the new decoder
	d := New(buf)

	// scan back the first one
	good := d.Scan()
	if !good {
		t.Error("got a decode error:", d.Err())
		return
	}

	// dump the types out and compare to expected
	buf = &bytes.Buffer{}
	d.WriteTypes(buf)
	if !sameLines(buf.String(), expectedTypes) {
		t.Error("not expected types")
	}

	// get the json of the first object
	b, err := d.JSON()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(b, expected1) {
		t.Error("not expected")
	}

	// scan the second object
	good = d.Scan()
	if !good {
		t.Error("got a decode error:", d.Err())
	}

	// get the json of the second object
	b, err = d.JSON()
	if err != nil {
		t.Fatal(err)
	}

	// check it is as expected
	if !bytes.Equal(b, expected2) {
		t.Error("not expected")
	}

	good = d.Scan()
	if good {
		t.Error("should have finished", d.Err())
	}
	if d.Err() != nil {
		t.Error("should not have got an error, but got:", d.Err())
	}
	o := d.Obj()
	if o != nil {
		t.Error("should not have got an obj if no current obj")
	}
	_, err = d.JSON()
	if err == nil {
		t.Error("should have got an error getting json if no current obj")
	}
}

var expected1 = []byte(`{
  "Age": 19,
  "Blob": null,
  "Height": 14.679,
  "Lengths": [
    8,
    1001
  ],
  "Name": "goober",
  "Other": {
    "Colour": 12345678
  },
  "Pimples": 12345678,
  "Sane": false,
  "Sizes": [
    0,
    25,
    0
  ]
}`)

var expected2 = []byte(`{
  "Age": 49,
  "Blob": "CQgHVQ==",
  "Height": 0,
  "Lengths": [
    0,
    -11155
  ],
  "Name": "gopher",
  "Other": {
    "Colour": 0
  },
  "Pimples": 0,
  "Sane": true,
  "Sizes": [
    0,
    0,
    0
  ]
}`)

var expectedTypes = `
type Type136 [3]int64	//[3]int

type Type132 []int64	//[]int

type other struct {
  Colour uint64
}

type bart struct {
  Name string
  Age int64
  Pimples int64
  Sane bool
  Lengths []int
  Other other
  Height float64
  Blob []byte
  Sizes [3]int
}`

// sameLines compares l and r line by line for matching sorted lines
func sameLines(l, r string) bool {

	ll := strings.Split(l, "\n")
	sort.Slice(ll, func(i, j int) bool {
		return ll[i] < ll[j]
	})
	ll = trimBlank(ll)
	rl := strings.Split(r, "\n")
	sort.Slice(rl, func(i, j int) bool {
		return rl[i] < rl[j]
	})
	rl = trimBlank(rl)

	for i, lin := range ll {
		if lin != rl[i] {
			return false
		}
	}
	return true
}
func trimBlank(ls []string) []string {
	i := 0
	for _, l := range ls {
		if l == "" {
			continue
		}
		ls[i] = l
		i++
	}
	return ls[:i]
}

func TestGoblinMap(t *testing.T) {
	type mini struct {
		Name string
		Size float32
	}
	type bart struct {
		Thing       map[string]string
		IntThing    map[int]float64
		StructThing map[float32]mini
	}

	thing := bart{
		Thing: map[string]string{
			"hi": "ruth",
		},
		IntThing: map[int]float64{
			1: 2.3,
			4: .00005,
		},
		StructThing: map[float32]mini{
			12.5: mini{
				Name: "jon",
				Size: 4.5,
			},
			1.5: mini{
				Name: "ty",
				Size: 8,
			},
		},
	}

	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(thing)
	if err != nil {
		t.Fatal("shame", err)
	}

	d := New(buf)
	good := d.Scan()
	if !good {
		t.Error("got a decode error:", d.Err())
		return
	}

	b, err := d.JSON()
	if err != nil {
		t.Error("could not get json string", err)
	}

	if string(b) != expMaps {
		t.Error("did not get the expected maps")
		t.Log(string(b))
	}

}

var expMaps = `{
  "IntThing": {
    "1": 2.3,
    "4": 0.00005
  },
  "StructThing": {
    "1.5": {
      "Name": "ty",
      "Size": 8
    },
    "12.5": {
      "Name": "jon",
      "Size": 4.5
    }
  },
  "Thing": {
    "hi": "ruth"
  }
}`

func TestRootObjMap(t *testing.T) {
	m := map[int]string{
		25: "zero",
		1:  "wun",
		2:  "too",
		33: "flirty flee",
	}
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(m)
	if err != nil {
		t.Fatal("shame", err)
	}

	d := New(buf)
	good := d.Scan()
	if !good {
		t.Error("got a decode error:", d.Err())
		return
	}

	buf = &bytes.Buffer{}
	d.WriteTypes(buf)
	if buf.String() != "type Type148 map[int64]string\n\n" {
		t.Errorf("not expected map types: %q", buf.String())
	}

	rm := d.Obj().(map[string]interface{})
	for k, v := range m {
		ks := strconv.Itoa(k)
		if rm[ks].(string) != v {
			t.Errorf("%s was wrong, expected %s got %d", ks, v, rm[ks])
		}
	}
}

func TestRootObjPrimitive(t *testing.T) {
	fixs := []struct {
		val   interface{}
		check func(interface{})
	}{
		{
			val: 238,
			check: func(out interface{}) {
				if out.(int64) != int64(238) {
					t.Error("did not decode a correct int")
				}
			},
		},
		{
			val: "hello weld",
			check: func(out interface{}) {
				if out.(string) != "hello weld" {
					t.Error("did not decode a correct string")
				}
			},
		},
		{
			val: []int{9, 7, 5},
			check: func(out interface{}) {
				o := out.([]interface{})
				exp := []int64{9, 7, 5}
				for i, no := range o {
					if no.(int64) != exp[i] {
						t.Error("got bad slice element")
					}
				}
			},
		},
		{
			val: [3]int{13, 2, 12},
			check: func(out interface{}) {
				o := out.([]interface{})
				exp := []int64{13, 2, 12}
				for i, no := range o {
					if no.(int64) != exp[i] {
						t.Error("got bad array element")
					}
				}
			},
		},
	}

	for _, fx := range fixs {
		// gob encode it
		buf := &bytes.Buffer{}
		enc := gob.NewEncoder(buf)
		err := enc.Encode(fx.val)
		if err != nil {
			t.Fatal("shame", err)
		}

		// decode it
		d := New(buf)
		good := d.Scan()
		if !good {
			t.Error("got a decode error:", d.Err())
			return
		}

		// confirm
		fx.check(d.Obj())
	}
}

func TestDecodeInt(t *testing.T) {
	cases := []struct {
		in  []byte
		exp int
	}{
		{in: []byte{0x00}, exp: 0},
		{in: []byte{0xFF, 0x81}, exp: -65},
		{in: []byte{0xFE, 0x01, 0x03}, exp: -130},
		{in: []byte{0x03}, exp: -2},
		{in: []byte{0x02}, exp: 1},
		{in: []byte{0x01}, exp: -1},
		{in: []byte{0xFF, 0x80}, exp: 64},
		{in: []byte{0xFE, 0x02, 0x00}, exp: 256},
		{in: []byte{0xFD, 0x33, 0x02, 0x00}, exp: 1671424},
		{in: []byte{0xfe, 0x57, 0x25}, exp: -11155},
	}
	for i, c := range cases {
		d := &decoder{
			b: c.in,
		}

		out, err := d.decodeInt()
		if err != nil {
			t.Error(err)
		}
		if int(out) != c.exp {
			t.Errorf("%d) wanted:%d got %d", i, c.exp, out)
		}
	}
}

func TestDecodeBool(t *testing.T) {
	cases := []struct {
		in  []byte
		exp bool
	}{
		{in: []byte{0x00}, exp: false},
		{in: []byte{0x01}, exp: true},
	}
	for i, c := range cases {
		d := &decoder{
			b: c.in,
		}
		out := d.decodeBool()
		if out != c.exp {
			t.Errorf("%d) wanted:%t got %t", i, c.exp, out)
		}
	}
}
