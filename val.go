package goblin

import (
	"fmt"
	"io"
	"math"
	"math/bits"
	"strings"
)

const (
	// the primitives
	tBool      typeID = 1  //
	tInt       typeID = 2  //
	tUint      typeID = 3  //
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
	// all struct fields can be skipped and contain no data
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
	da []byte         // for strings and []byte
	nu uint64         // for all int, uint, float
	sl slice          // for slice type
	ma map[string]val // for map type. N.B. only primitive types supported in the index for now, they will be converted to string
	st []field        // for struct type
}

// ToUint returns the unsigned integer value of the val data
func (v val) ToUint() uint64 {
	return v.nu
}

// ToInt returns the integer value of the val data
func (v val) ToInt() int64 {
	if v.nu&1 != 0 {
		return ^int64(v.nu >> 1)
	}
	return int64(v.nu >> 1)
}

func (v val) ToFloat() float64 {
	fv := bits.ReverseBytes64(v.nu)
	return math.Float64frombits(fv)
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

func (v val) dump(w io.Writer) {
	v.dumpi(w, 0)
}

func pad(i int) string {
	return strings.Repeat(" ", i)
}

func (v val) dumpi(w io.Writer, indent int) {
	switch v.t {
	case tBool:
		fmt.Fprintf(w, "%s%t \t(%s)\n", pad(indent), v.ToBool(), v.t)
	case tInt:
		fmt.Fprintf(w, "%s%d \t(%s)\n", pad(indent), v.ToInt(), v.t.String())
	case tUint:
		fmt.Fprintf(w, "%s%d \t(%s)\n", pad(indent), v.ToUint(), v.t.String())
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
	case tMap:
		// TODO
	default:
		var iv interface{}
		switch v.t {
		case tBool:
			iv = v.ToBool()
		case tInt:
			iv = v.ToInt()
		case tUint:
			iv = v.ToUint()
		case tString:
			iv = string(v.da)
		case tFloat:
			iv = v.ToFloat()
		}
		fmt.Fprintf(w, "%s%v \t(%s)\n", pad(indent), iv, v.t)
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
