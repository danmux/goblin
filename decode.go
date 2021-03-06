package goblin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const uint64Size = 8

// decoder does all the hard work decoding a gob file
type decoder struct {
	b []byte    // the current buffer of data just read in
	r io.Reader // the reader to read in chunks of data

	types   map[typeID]val // the type definitions for this decoder
	lastVal *val           // the last scanned value
	lastErr error          // errors on the last scan

	level int      // for debugging
	path  []string // for debugging and pretty errors
}

// New returns a new decoder
func New(r io.Reader) *decoder {
	d := &decoder{
		r: r,
	}
	return d
}

// Scan decodes the next available value.
// If this is the first call to Scan on a new decoder, then the gob type information
// will be decoded.
func (d *decoder) Scan() bool {
	d.lastErr = nil
	d.lastVal = nil
	// if we have not set up yet
	if len(d.types) == 0 {
		d.lastErr = d.setupTypes()
		if d.lastErr != nil {
			return false
		}
		// d.b will be set up for the first data
	} else {
		// the second data needs us to set up d.b
		d.lastErr = d.getBuf()
		if d.lastErr != nil {
			return false
		}
		if len(d.b) == 0 {
			// this should be the normal end
			return false
		}
	}

	d.lastErr = d.decodeData()
	return d.lastErr == nil
}

// Err returns any errors from the last call to Scan
func (d *decoder) Err() error {
	return d.lastErr
}

// Obj returns the result of the last Scan as a limited object where any map keys are
// represented as strings, making the object json compatible.
func (d *decoder) Obj() interface{} {
	if d.lastVal == nil {
		return nil
	}
	return d.lastVal.obj()
}

// Json returns the result of a call to Obj marshalled as an indented JSON []byte
func (d *decoder) JSON() ([]byte, error) {
	if d.lastVal == nil {
		return nil, errors.New("can not return json representation of none existant object")
	}
	return json.MarshalIndent(d.lastVal.obj(), "", "  ")
}

// WriteTypes dumps to the given writer the representation of the type information
// in a golang struct compatible way
func (d *decoder) WriteTypes(w io.Writer) {
	for k, v := range d.types {
		if k > minUserType {
			d.toType(w, &v)
			fmt.Fprintln(w)
		}
	}
}

func (d *decoder) setupTypes() error {
	d.initTypes()

	// load in the wiretypes into the types index
	return d.decodeTypes()
}

// decodeTypes loads in the wireTypes from the gob types section
func (d *decoder) decodeTypes() error {
	for {
		err := d.getBuf()
		if err != nil {
			return err
		}
		start := d.b

		// decode the length - though we don't actually use it in here
		_, err = d.decodeUint()
		if err != nil {
			return err
		}
		// get the negative type ID
		typ, err := d.decodeInt()
		if err != nil {
			return err
		}
		// if the type id is > 0 then it is the actual value data
		if typ > 0 {
			d.b = start // restore the last three bytes that are the start of a value
			return nil
		}
		ty := val{}

		ty.copy(d.types[tWireType])
		nt, err := d.start(ty)
		if err != nil {
			return err
		}
		// add it to the index of types
		d.types[-typeID(typ)] = nt
	}
}

// decodeTypes loads in the wireTypes from the gob types section
func (d *decoder) decodeData() error {
	// if we have used up all the bytes then there is no more data
	if len(d.b) == 0 {
		return nil
	}
	// decode the length - though we don't actually use it
	_, err := d.decodeUint()
	if err != nil {
		return err // TODO - richer nested errors
	}
	// get the type ID
	typ, err := d.decodeInt()
	if err != nil {
		return err
	}

	var data val
	tid := typeID(typ)
	// top level primitive types
	if tid < minUserType {
		data = val{
			t: tid,
		}
	} else {
		typeDesc, ok := d.types[tid]
		if !ok {
			return fmt.Errorf("got type index entry %d that does not exist", tid)
		}
		// get the initial type
		data = d.fromWireType(typeDesc)
	}

	// is it a top level type that has a field delta of 0
	if data.t != tStruct {
		//then read in the zero delta (wonder why this exists?)
		_, err = d.decodeUint()
		if err != nil {
			return err
		}
	}

	data, err = d.start(data)
	if err != nil {
		return err
	}
	d.lastVal = &data

	return nil
}

func (d decoder) paths() string {
	return strings.Join(d.path, ".")
}

func (d *decoder) start(v val) (val, error) {
	// copy behaves as deep copy as val has no references
	d.level = -1
	x := v
	err := d.decode(&x)
	return x, err
}

func (d *decoder) decode(x *val) error {
	switch x.t {
	case tBool:
		b := d.decodeBool()
		if b {
			x.da = []byte{1}
		} else {
			x.da = []byte{0}
		}
		return nil
	case tInt, tUint, tFloat:
		iv, err := d.decodeUint()
		if err != nil {
			return err
		}
		x.nu = iv
		return nil
	case tBytes, tString:
		return d.decodeBytes(x)
	case tSlice:
		err := d.decodeSlice(x)
		return err
	case tMap:
		return d.decodeMap(x)
	case tStruct:
		return d.decodeStruct(x)
	}
	// dereference the indexed type up and add a copy to the val
	t, ok := d.types[x.t]
	if !ok {
		return fmt.Errorf("%q found type id that is not in index: %d", d.paths(), x.t)
	}
	// the wiretypes are already expanded
	if x.t < minUserType {
		x.copy(t)
	} else {
		// for other dynamic types
		dt := d.fromWireType(t)
		x.copy(dt)
	}
	// and decode it
	return d.decode(x)
}

func (d *decoder) decodeMap(v *val) error {
	d.path = append(d.path, "")
	d.level++
	defer func() {
		d.level--
		d.path = d.path[:len(d.path)-1]
	}()

	// get element count
	ui, err := d.decodeUint()
	if err != nil {
		return err
	}
	v.ma.els = map[string]val{}
	for i := 0; i < int(ui); i++ {
		k := val{
			t: v.ma.kt,
		}
		err := d.decode(&k)
		if err != nil {
			return err
		}
		nv := val{
			t: v.ma.vt,
		}
		err = d.decode(&nv)
		if err != nil {
			return err
		}
		// all keys take their standard string representation
		v.ma.els[fmt.Sprintf("%v", k.obj())] = nv
	}

	return nil
}

func (d *decoder) decodeStruct(x *val) error {
	fc := -1
	d.path = append(d.path, "")
	d.level++
	defer func() {
		d.level--
		d.path = d.path[:len(d.path)-1]
	}()
	for {
		// get the field delta
		delta, err := d.decodeUint()
		if err != nil {
			return err
		}
		if delta == 0 { // end of fields with the 0 delta terminator
			return nil
		}
		fc += int(delta)
		if fc >= len(x.st) {
			return fmt.Errorf("%s bad encoding more fields than the type len: %d expected: %d ", d.paths(), fc, len(x.st))
		}

		d.path[d.level] = x.st[fc].name
		x.st[fc].nonZero = true
		err = d.decode(&x.st[fc].v)
		if err != nil {
			return err
		}
	}
}

func (d *decoder) decodeBytes(v *val) error {
	len, err := d.decodeUint()
	if err != nil {
		return err
	}
	v.da = make([]byte, len)
	copy(v.da, d.b[:len])
	d.b = d.b[len:]
	return nil
}

func (d *decoder) decodeSlice(v *val) error {
	ui, err := d.decodeUint()
	if err != nil {
		return err
	}
	len := int(ui)
	v.sl.els = make([]val, len)
	for i := 0; i < len; i++ {
		// give each elem the predefined type for this slice
		v.sl.els[i].t = v.sl.t
		// and decode each one
		if err := d.decode(&(v.sl.els[i])); err != nil {
			return err
		}
	}
	return nil
}

func (d *decoder) getBuf() error {
	buf := make([]byte, uint64Size)
	n, err := d.r.Read(buf)
	if n == 0 {
		d.b = nil
		return nil
	}
	if err != nil && err != io.EOF {
		return err
	}

	l, v, err := decodeUint(buf)
	if err != nil {
		return err
	}
	// make enough for the whole block
	blockLen := l + int(v)
	d.b = make([]byte, blockLen)
	// work out how much more we actually need to read
	// taking off what we already read
	toRead := blockLen - n

	// copy over what we have read already
	for i := 0; i < n; i++ {
		d.b[i] = buf[i]
	}
	// fill the bit we have not read
	buf = d.b[n:]
	n, err = d.r.Read(buf)
	if n != toRead {
		return fmt.Errorf("could not read the required number (%d) of bytes, only read (%d)", toRead, n)
	}
	return err
}

// decodeUint reads an encoded unsigned integer from b
// Does not check for overflow.
//
// This func is closely copied from gob decode in the stdlib so
// is copyright the Go authors. (https://golang.org/src/encoding/gob/decode.go)
func (d *decoder) decodeUint() (x uint64, err error) {
	l, v, err := decodeUint(d.b)
	if err != nil {
		return 0, err
	}
	d.b = d.b[l:]
	return v, nil
}

func decodeUint(buf []byte) (l int, x uint64, err error) {
	b := buf[0]
	if b <= 0x7f {
		return 1, uint64(b), nil
	}
	n := -int(int8(b))
	if n > uint64Size {
		return 0, 0, errors.New("bad unit size")
	}
	if len(buf) <= n {
		return 0, 0, fmt.Errorf("invalid uint data length %d: exceeds input size %d", n, len(buf))
	}
	// Don't need to check error; it's safe to loop regardless.
	// Could check that the high byte is zero but it's not worth it
	for _, v := range buf[1 : 1+n] {
		x = x<<8 | uint64(v)
	}
	return 1 + n, x, nil
}

func (d *decoder) decodeInt() (int64, error) {
	x, err := d.decodeUint()
	if err != nil {
		return 0, err
	}
	if x&1 != 0 {
		return ^int64(x >> 1), nil
	}
	return int64(x >> 1), nil
}

func (d *decoder) decodeBool() bool {
	b := d.b[0]
	d.b = d.b[1:]
	return b == 1
}
