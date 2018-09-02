package goblin

import (
	"fmt"
	"io"
)

const (
	// the predefined types that describe types
	tWireType  typeID = 16
	tArrayType typeID = 17
	// tCommonType typeID = 18 // in out implementation commontype is inlined already
	tSliceType  typeID = 19
	tStructType typeID = 20
	tFieldType  typeID = 21
	tMapType    typeID = 23

	minUserType typeID = 30
)

// v has to be a val representing a wiretype
func (d *decoder) toType(w io.Writer, v *val) {
	// must be a wireType
	if v.t != tStruct {
		return
	}
	// get the name
	name := wireTypeName(v)
	// what type of wireType
	for i, f := range v.st {
		if f.nonZero {
			switch i {
			case 0: // Array field
				tName := fmt.Sprintf("Type%d", f.v.st[0].v.st[1].v.nu)
				kt := d.idToType(int(f.v.st[1].v.ToInt()))
				l := f.v.st[2].v.ToInt()
				fmt.Fprintf(w, "type %s [%d]%s\t//%s\n", tName, l, kt, name)
			case 1: // slice field - name is the original type
				tName := fmt.Sprintf("Type%d", f.v.st[0].v.st[1].v.nu)
				kt := d.idToType(int(f.v.st[1].v.ToInt()))
				fmt.Fprintf(w, "type %s []%s\t//%s\n", tName, kt, name)
			case 2: // struct field
				fmt.Fprintf(w, "type %s struct {\n", name)
				for _, fn := range f.v.st[1].v.sl.els {
					ty := d.idToType(int(fn.st[1].v.ToInt()))
					fmt.Fprintf(w, "  %s %s\n", string(fn.st[0].v.da), ty)
				}
				fmt.Fprintln(w, "}")
			case 3: // map field
				kt := d.idToType(int(f.v.st[1].v.ToInt()))
				kv := d.idToType(int(f.v.st[2].v.ToInt()))
				fmt.Fprintf(w, "type %s map[%s]%s\n", name, kt, kv)
			}
		}
	}
}

func wireTypeName(v *val) string {
	// must be a wiretype struct
	if v.t != tStruct {
		return ""
	}
	// what type of wireType
	for _, f := range v.st {
		if f.nonZero {
			name := string(f.v.st[0].v.st[0].v.da)
			if name == "" {
				name = fmt.Sprintf("Type%d", f.v.st[0].v.st[1].v.nu)
			}
			return name
		}
	}
	return ""
}

func (d *decoder) idToType(id int) string {
	t := typeLookup[id]
	if t != "" {
		return t
	}
	ty, ok := d.types[typeID(id)]
	if !ok {
		return ""
	}
	return wireTypeName(&ty)
}

// setupTypes sets up the types index of types.
// It is prepopulated with the wireType which is
// the type that is used to describe the other types
func (d *decoder) initTypes() {
	// types is the lookup table of the type definitions
	d.types = map[typeID]val{
		tFieldType: val{
			t: tStruct,
			st: []field{
				{
					name: "name",
					v: val{
						t: tString,
					},
				},
				{
					name: "id",
					v: val{
						t: tInt,
					},
				},
			},
		},
		tMapType: val{
			t: tStruct,
			st: []field{
				{
					name: "commonType",
					v: val{
						t: tStruct,
						st: []field{
							{
								name: "name",
								v: val{
									t: tString,
								},
							},
							{
								name: "id",
								v: val{
									t: tInt,
								},
							},
						},
					},
				},
				{
					name: "key",
					v: val{
						t: tInt,
					},
				},
				{
					name: "elem",
					v: val{
						t: tInt,
					},
				},
			},
		},
		tStructType: val{
			t: tStruct,
			st: []field{
				{
					name: "commonType",
					v: val{
						t: tStruct,
						st: []field{
							{
								name: "name",
								v: val{
									t: tString,
								},
							},
							{
								name: "id",
								v: val{
									t: tInt,
								},
							},
						},
					},
				},
				{
					name: "fields",
					v: val{
						t: tSlice,
						sl: slice{
							t: tFieldType,
						},
					},
				},
			},
		},
		tSliceType: val{
			t: tStruct,
			st: []field{
				{
					name: "commonType",
					v: val{
						t: tStruct,
						st: []field{
							{
								name: "name",
								v: val{
									t: tString,
								},
							},
							{
								name: "id",
								v: val{
									t: tInt,
								},
							},
						},
					},
				},
				{
					name: "elem",
					v: val{
						t: tInt,
					},
				},
			},
		},
		tArrayType: val{
			t: tStruct,
			st: []field{
				{
					name: "commonType",
					v: val{
						t: tStruct,
						st: []field{
							{
								name: "name",
								v: val{
									t: tString,
								},
							},
							{
								name: "id",
								v: val{
									t: tInt,
								},
							},
						},
					},
				},
				{
					name: "elem",
					v: val{
						t: tInt,
					},
				},
				{
					name: "len",
					v: val{
						t: tInt,
					},
				},
			},
		},
	}

	d.types[tWireType] = val{
		t: tStruct,
		st: []field{
			{
				name: "arrayT",
				v:    d.types[tArrayType],
			},
			{
				name: "sliceT",
				v:    d.types[tSliceType],
			},
			{
				name: "structT",
				v:    d.types[tStructType],
			},
			{
				name: "mapT",
				v:    d.types[tMapType],
			},
		},
	}
}
