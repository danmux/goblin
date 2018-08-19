package goblin

import "testing"

func TestIntToByte(t *testing.T) {
	v := val{}
	val := int64(-1001)
	v.intToByte(val)

	got := v.ToInt()
	if got != val {
		t.Errorf("got %d wanted %d", got, val)
	}

	// make sure unsigned int bytes are stored fine
	uVal := ^uint64(0)
	// as an int64 this will be == -1, but the bytes are unmolested so it
	// all works out fine
	v.intToByte(int64(uVal))

	gotU := v.ToUint()
	if gotU != uVal {
		t.Errorf("got %d wanted %d", gotU, uVal)
	}
}
