package goblin

const (
	// the predefined types that describe types
	tWireType   typeID = 16
	tArrayType  typeID = 17
	tCommonType typeID = 18
	tSliceType  typeID = 19
	tStructType typeID = 20
	tFieldType  typeID = 21
	tMapType    typeID = 23

	minUserType typeID = 30
)

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
