package goblin

import (
	"fmt"
	"io"
	"strings"
)

const (
	// the primitives
	tBool      typeID = 1  //
	tInt       typeID = 2  //
	tUint      typeID = 3  // TODO
	tFloat     typeID = 4  // TODO
	tBytes     typeID = 5  // TODO
	tString    typeID = 6  //
	tComplex   typeID = 7  // TODO
	tInterface typeID = 8  // TODO
	tSlice     typeID = 9  //
	tMap       typeID = 10 // TODO
	tStruct    typeID = 11 //
)

type typeID int

// the slice value
type slice struct {
	t   typeID
	els []val
}

func (s slice) dumpi(w io.Writer, indent int) {
	fmt.Fprintf(w, "%s[\n", pad(indent))
	for _, v := range s.els {
		v.dumpi(w, indent+2)
	}
	fmt.Fprintf(w, "%s]\n", pad(indent))
}

func (s *slice) copy(os slice) {
	s.t = os.t
	s.els = make([]val, len(os.els))
	for i, v := range os.els {
		nv := val{}
		nv.copy(v)
		s.els[i] = nv
	}
}

// the field value
type field struct {
	nonZero bool
	name    string
	v       val
}

func (f field) dumpi(w io.Writer, indent int) {
	fmt.Fprintf(w, "%s%q:\n", pad(indent), f.name)
	if !f.nonZero {
		fmt.Fprintf(w, "%s(nil)\n", pad(indent+2))
		return
	}
	f.v.dumpi(w, indent+2)
}

func (f *field) copy(of field) {
	f.nonZero = of.nonZero
	f.name = of.name
	f.v = val{}
	f.v.copy(of.v)
}

// val represents values of any of the builtin type
type val struct {
	t  typeID         // what primitive type id
	da []byte         // for primitive type
	sl slice          // for slice type
	ma map[string]val // for map type. N.B. only primitive types supported in the index for now, they will be converted to string
	st []field        // for struct type
}

// ToInt returns the integer value of the val data
func (v val) ToInt() int64 {
	var ans int64
	for i := 0; i < 8; i++ {
		t := int64(v.da[i])
		ans += t << uint(8*(7-i))
	}
	return ans
}

// ToUint returns the unsigned integer value of the val data
func (v val) ToUint() uint64 {
	return uint64(v.ToInt())
}

// ToBool returns the bool value of the val data
func (v val) ToBool() bool {
	if len(v.da) == 0 { // default false is represented by missing data
		return false
	}
	return v.da[0] == 1 // should always be a 1
}

func (v *val) copy(t val) {
	v.t = t.t
	v.da = nil
	v.ma = map[string]val{}
	for k, v := range t.ma {
		nv := val{}
		nv.copy(v)
		v.ma[k] = nv
	}
	v.sl = slice{}
	v.sl.copy(t.sl)

	v.st = make([]field, len(t.st))
	for i, f := range t.st {
		nf := field{}
		nf.copy(f)
		v.st[i] = nf
	}
}

func (v *val) intToByte(i int64) {
	// TODO consider minimum packaging
	v.da = make([]byte, 8)
	for n := 0; n < 8; n++ {
		v.da[7-n] = byte(i) // big endian
		i = i >> 8
	}
}

func (v val) dump(w io.Writer) {
	v.dumpi(w, 0)
}

func pad(i int) string {
	return strings.Repeat(" ", i)
}

func (v val) IsNil() bool {
	switch v.t {
	case tSlice, tStruct, tMap, tBool: // bool false is encoded as missing field
		return false
	}
	return len(v.da) == 0
}

func (v val) dumpi(w io.Writer, indent int) {
	if v.IsNil() {
		fmt.Fprintf(w, "%s(nil)\n", pad(indent))
		return
	}

	switch v.t {
	case tBool:
		fmt.Fprintf(w, "%s%t \t(%s)\n", pad(indent), v.ToBool(), v.t)
	case tInt, tUint:
		fmt.Fprintf(w, "%s%d \t(%s)\n", pad(indent), v.ToInt(), v.t.String())
	case tString:
		fmt.Fprintf(w, "%s%q \t(%s)\n", pad(indent), string(v.da), v.t.String())
	case tSlice:
		v.sl.dumpi(w, indent)
	case tStruct:
		fmt.Fprintf(w, "%s{\n", pad(indent))
		for _, f := range v.st {
			f.dumpi(w, indent+2)
		}
		fmt.Fprintf(w, "%s}\n", pad(indent))
	}
}

// v must be a representation of a wire type
func (d *decoder) fromWireType(v val) val {
	// the data val
	dv := val{}

	var name string
	// what type are we constructing
	for _, t := range v.st {
		if t.nonZero {
			name = t.name
			v = t.v
			break
		}
	}

	switch name {
	case "structT":
		dv.t = tStruct
		for _, f := range v.st[1].v.sl.els { // sl is a slice of field types
			fld := field{
				name: string(f.st[0].v.da), // first field of the field definition is name
				v:    d.makeVal(typeID(f.st[1].v.ToInt())),
			}
			if fld.v.t == tBool {
				fld.nonZero = true // missing field data == a false value
			}
			dv.st = append(dv.st, fld)
		}
	case "sliceT":
		dv.t = tSlice
		dv.sl.t = typeID(v.st[1].v.ToInt()) // second field is the 'elem' field
	}

	// types[-typeID(typ)] = nt
	return dv
}

func (d *decoder) makeVal(t typeID) val {
	// is it a wiretype definition in the table
	if t >= minUserType {
		return d.fromWireType(d.types[t])
	}
	return val{
		t: t,
	}
}
