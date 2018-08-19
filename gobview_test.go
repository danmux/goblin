package goblin

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestGoblin(t *testing.T) {
	type other struct {
		Colour uint
	}

	type bart struct {
		Name    string
		Age     int
		Sane    bool
		Lengths []int
		Other   *other
	}

	thing := bart{
		Name:    "goober",
		Age:     19,
		Sane:    false,
		Lengths: []int{8, 1001},
		Other: &other{
			Colour: uint(12345678),
		},
	}

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
	}
	err = enc.Encode(thing2)
	if err != nil {
		t.Fatal("shame", err)
	}

	spew.Dump(buf.Bytes())

	d := New(buf)
	good := d.Scan()
	if !good {
		t.Error("got a decode error:", d.Err())
		return
	}
	println(d.String())

	good = d.Scan()
	if !good {
		t.Error("got a decode error:", d.Err())
	}
	println(d.String())

	good = d.Scan()
	if good {
		t.Error("should have finished", d.Err())
	}
	if d.String() != "" {
		t.Error("should not have got a string", d.String())
	}
	if d.Err() != nil {
		t.Error("should not have got an error, but got:", d.Err())
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
