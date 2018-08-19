// package goblin reads gob encoded data and constructs a dynamic in memory representation
// of the decoded data for conversion into json, or dynamic access
//
// The folling ins an example breakdown of a full gob file including the wireType data
//
// 39            - 57
// ff            - id
// 81            - id -65
// 03            - fld delta of implicit wireType = structT
// 01            - delta field 0
// 01            -   commonType field 0 - name
// 04            -     len 4
// 62 61 72 74   -     "bart"
// 01            -     delta field 1
// ff            -     id
// 82            -     id - = 65
// 00            -   end of commonType
// 01            - delta field 1 slice of fieldType
// 04            - 4 elements in slice
// 01            -    delta field 0 - name
// 04            -     name len 4
// 4e 61 6d 65   -     "Name"
// 01            -    delta field 1 - id
// 0c            -      6
// 00            -  end of fld
// 01            -    delta field 0 - name
// 03            -     len 3
// 41 67 65      -      "Age"
// 01            -    delta field 1 - id
// 04            -      4
// 00            -  end of fld
// 01            -    delta field 0 - name
// 04            -      len 4
// 53 61 6e 65   -     "Sane"
// 01            -    delta field - id
// 02            -       1
// 00            -  end of struct fld
// 01            -    delta field 0 name
// 07            -       len 7
// 4c 65 6e 67 74 68 73 "Lengths"
// 01            -    delta field 1 id
// ff            -     id
// 84            -     id  66
// 00            -  end of fld fieldType
// 00            - end of slice
// 00            -end of wireType
// 13            ------------------------- new length 19
// ff            - id
// 83            - id  - 66
// 02            - fld delta of implicit wireType = sliceT
// 01            - fld delta = field 0 = embedded common struct
// 01            -   delta fld 0 of commonStruct - (string)
// 05            -   len 5
// 5b 5d 69 6e 74 -    "[]int"
// 01            -   delta fld 1  (int)
// ff            -
// 84            -  66
// 00            - end of common struct
// 01            - fld delta = field 0 'elem' type int
// 04            - 02 = (int type)
// 00            - end of sliceT
// 00            - end of wireType
// 13          --------------------------- beginning of actual data element length 19
// ff            -
// 82            - type id 65  (the struct with 4 fields described above)
// 01            - fld delta = field 0 (a string - from description above)
// 06            -   len 6
// 67 6f 6f 62 65 72 - "goober"
// 01            - fld delta = field 1 (an int- from description above)
// 26            -   19 (38/2)
// 02            - fld delta = field 3 (skipped bool 'Sane' so that will be default false, slice of int)
// 02            -   2 elements
// 10            -      8  (16/2)
// fe            -
// 07            -
// d2            -      1001
// 00            - end of val

package goblin
