package goblin

import (
	"math"
	"math/bits"
)

const (
	// the primitives
	tBool   typeID = 1
	tInt    typeID = 2
	tUint   typeID = 3
	tFloat  typeID = 4
	tBytes  typeID = 5
	tString typeID = 6
	// tComplex   typeID = 7 // TODO
	// tInterface typeID = 8 // TODO
	tSlice  typeID = 9
	tMap    typeID = 10
	tStruct typeID = 11
)

var typeLookup = map[int]string{
	1:  "bool",
	2:  "int64",
	3:  "uint64",
	4:  "float64",
	5:  "[]byte",
	6:  "string",
	7:  "",
	8:  "",
	9:  "slice",
	10: "map",
	11: "struct",
}

type typeID int

type mapv struct {
	kt  typeID // key type id
	vt  typeID // value type id
	els map[string]val
}

func (m *mapv) copy(om mapv) {
	m.kt = om.kt
	m.vt = om.vt
	for k, v := range om.els {
		nv := val{}
		nv.copy(v)
		m.els[k] = nv
	}
}

func (m mapv) obj() interface{} {
	ma := map[string]interface{}{}
	for k, v := range m.els {
		ma[k] = v.obj()
	}
	return ma
}

// the slice value
type slice struct {
	t   typeID
	els []val
}

func (s slice) obj() interface{} {
	var ar []interface{}
	for _, v := range s.els {
		ar = append(ar, v.obj())
	}
	return ar
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

// tructs are slices of fields
type structv []field

func (s structv) obj() interface{} {
	ma := map[string]interface{}{}
	for _, v := range s {
		ma[v.name] = v.v.obj()
	}
	return ma
}

func (s *structv) copy(os structv) {
	fls := make([]field, len(os))
	for i, f := range os {
		nf := field{}
		nf.copy(f)
		fls[i] = nf
	}
	*s = structv(fls)
}

func (f *field) copy(of field) {
	f.nonZero = of.nonZero
	f.name = of.name
	f.v = val{}
	f.v.copy(of.v)
}

// val represents values of any of the builtin type
type val struct {
	t  typeID  // what primitive type id
	da []byte  // for strings and []byte
	nu uint64  // for all int, uint, float
	sl slice   // for slice type
	ma mapv    // for map type. N.B. only primitive types supported in the index for now, they will be converted to string
	st structv // for struct type
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

func (v val) obj() interface{} {
	switch v.t {
	case tBool:
		return v.ToBool()
	case tInt:
		return v.ToInt()
	case tUint:
		return v.ToUint()
	case tBytes:
		return v.da
	case tFloat:
		return v.ToFloat()
	case tString:
		return string(v.da)
	case tSlice:
		return v.sl.obj()
	case tStruct:
		return v.st.obj()
	case tMap:
		return v.ma.obj()
	}
	return nil
}

func (v *val) copy(t val) {
	v.t = t.t
	v.da = nil
	v.ma.copy(t.ma)
	v.sl.copy(t.sl)
	v.st.copy(t.st)
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
	case "mapT":
		dv.t = tMap
		dv.ma.kt = typeID(v.st[1].v.ToInt())
		dv.ma.vt = typeID(v.st[2].v.ToInt())
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
	case "sliceT", "arrayT":
		dv.t = tSlice
		dv.sl.t = typeID(v.st[1].v.ToInt()) // second field is the 'elem' field
	}

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
