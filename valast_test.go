package valast

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/hexops/autogold"
	"github.com/hexops/valast/internal/test"
)

type foo struct {
	bar string
}

type baz struct {
	Bam  complex64
	zeta foo
	Beta interface{}
}

type ExportedBaz struct {
	Bam  complex64
	zeta foo
	Beta interface{}
}

func TestString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
		err   string
	}{
		{
			name:  "bool",
			input: true,
		},
		{
			name:  "bool_unqualify",
			input: false,
			opt:   &Options{Unqualify: true},
		},
		{
			name:  "int32",
			input: int32(1234),
		},
		{
			name:  "int32_unqualify",
			input: int32(1234),
			opt:   &Options{Unqualify: true},
		},
		{
			name:  "uintptr",
			input: uintptr(1234),
		},
		{
			name:  "uintptr_unqualify",
			input: uintptr(1234),
			opt:   &Options{Unqualify: true},
		},
		{
			name:  "float64",
			input: float64(1.234),
		},
		{
			name:  "float64_unqualify",
			input: float64(1.234),
			opt:   &Options{Unqualify: true},
		},
		{
			name:  "complex64",
			input: complex64(1.234),
		},
		{
			name:  "complex64_unqualify",
			input: complex64(1.234),
			opt:   &Options{Unqualify: true},
		},
		{
			name:  "string",
			input: string("hello \t world"),
		},
		{
			name: "string_unqualify",
			input: string(`one
two
three`),
			opt: &Options{Unqualify: true},
		},
		{
			name: "struct_anonymous",
			input: struct {
				a, b int
				V    string
			}{a: 1, b: 2, V: "efg"},
		},
		{
			name: "struct_same_package",
			input: baz{
				Bam: 1.34,
				zeta: foo{
					bar: "hello",
				},
			},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "struct_external_package",
			input: test.NewBaz(),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name: "array",
			input: [2]*baz{
				&baz{Beta: "foo"},
				&baz{Beta: 123},
			},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name: "slice",
			input: []*baz{
				&baz{Beta: "foo"},
				&baz{Beta: 123},
				&baz{Beta: 3},
			},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			name:  "nil",
			input: nil,
		},
		{
			name: "interface_anonymous_nil",
			input: interface {
				a() string
			}(nil),
		},
		{
			name: "interface_anonymous",
			input: &struct {
				v interface {
					String() string
					Baz() (err error)
				}
			}{v: test.NewBaz()},
		},
		{
			name: "interface_builtin",
			input: &struct {
				v error
			}{v: nil},
		},
		{
			// TODO: `&test.Baz{Bam: (1.34+0i), zeta: &test.foo{bar: "hello"}}` is not valid code because `zeta` is unexported.
			name: "interface",
			input: &struct {
				v test.Bazer
			}{v: test.NewBaz()},
		},
		{
			name:  "unsafe_pointer",
			input: unsafe.Pointer(uintptr(0xdeadbeef)),
		},
		{
			name: "map",
			input: map[string]int32{
				"foo": 32,
				"bar": 64,
			},
		},
		// TODO: test and handle recursive struct, list, array, pointer
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got, err := String(reflect.ValueOf(tst.input), tst.opt)
			if tst.err != "" && tst.err != err.Error() || tst.err == "" && err != nil {
				t.Fatal("\ngot:\n", err, "\nwant:\n", tst.err)
				return
			}
			autogold.Equal(t, got)
		})
	}
}

// TestEdgeCases tests known edge-cases and past bugs that do not fit any of the broader test
// categories.
func TestEdgeCases(t *testing.T) {
	var (
		nilInterfacePointerBug test.Bazer
		bazer                  test.Bazer = test.NewBaz()
	)
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
		err   string
	}{
		{
			// TODO: make this produce an error
			name: "interface_pointer",
			input: &struct {
				v *test.Bazer
			}{v: &bazer},
			err: "valast: pointers to interfaces are not allowed, found *test.Bazer",
		},
		{
			// Ensures it does not produce &nil:
			//
			// 	./valast_test.go:179:9: cannot take the address of nil
			// 	./valast_test.go:179:9: use of untyped nil
			//
			name: "nil_interface_pointer_bug",
			input: &struct {
				v *test.Bazer
			}{v: &nilInterfacePointerBug},
			err: "valast: pointers to interfaces are not allowed, found *test.Bazer",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got, err := String(reflect.ValueOf(tst.input), tst.opt)
			if tst.err != "" && tst.err != err.Error() || tst.err == "" && err != nil {
				t.Fatal("\ngot:\n", err, "\nwant:\n", tst.err)
				return
			}
			autogold.Equal(t, got)
		})
	}
}

// TestExportedOnly_input tests the behavior of Options.ExportedOnly when enabled with a direct unexported input.
func TestExportedOnly_input(t *testing.T) {
	type (
		unexportedBool          bool
		unexportedInt           int
		unexportedInt8          int8
		unexportedInt16         int16
		unexportedInt32         int32
		unexportedInt64         int64
		unexportedUint          uint
		unexportedUint8         uint8
		unexportedUint16        uint16
		unexportedUint32        uint32
		unexportedUint64        uint64
		unexportedUintptr       uintptr
		unexportedFloat32       float32
		unexportedFloat64       float64
		unexportedComplex64     complex64
		unexportedComplex128    complex128
		unexportedArray         [1]float32
		unexportedInterface     error
		unexportedMap           map[string]string
		unexportedPointer       *int
		unexportedSlice         []int
		unexportedString        string
		unexportedStruct        struct{ A string }
		unexportedUnsafePointer unsafe.Pointer
	)
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
		err   string
	}{
		{
			name: "struct_same_package",
			input: baz{
				Bam: 1.34,
				zeta: foo{
					bar: "hello",
				},
			},
			err: "valast: cannot convert value of kind:struct type:valast.baz",
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "bool",
			input: unexportedBool(true),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "int",
			input: unexportedInt(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "int8",
			input: unexportedInt8(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "int16",
			input: unexportedInt16(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "int32",
			input: unexportedInt32(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "int64",
			input: unexportedInt64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "uint",
			input: unexportedUint(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "uint8",
			input: unexportedUint8(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "uint16",
			input: unexportedUint16(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "uint32",
			input: unexportedUint32(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "uint64",
			input: unexportedUint64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "uintptr",
			input: unexportedUintptr(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "float32",
			input: unexportedFloat32(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "float64",
			input: unexportedFloat64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "complex64",
			input: unexportedComplex64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "complex128",
			input: unexportedComplex128(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "array",
			input: unexportedArray{1.0},
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name: "interface",
			input: struct {
				V unexportedInterface
			}{V: nil},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
			err: "valast: cannot convert value of kind:struct type:struct { V valast.unexportedInterface }",
		},
		{
			// TODO: BUG: expect nil output
			name: "map",
			input: unexportedMap{
				"a": "b",
			},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "pointer",
			input: unexportedPointer(nil),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "slice",
			input: unexportedSlice{1, 2, 3},
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "string",
			input: unexportedString("hello"),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		{
			// TODO: BUG: expect nil output
			name:  "struct",
			input: unexportedStruct{A: "b"},
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
			err:   "valast: cannot convert value of kind:struct type:valast.unexportedStruct",
		},
		{
			// TODO: BUG: expect nil output
			name:  "unsafe_pointer",
			input: unexportedUnsafePointer(unsafe.Pointer(uintptr(0xdeadbeef))),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got, err := String(reflect.ValueOf(tst.input), tst.opt)
			if tst.err != "" && tst.err != err.Error() || tst.err == "" && err != nil {
				t.Fatal("\ngot:\n", err, "\nwant:\n", tst.err)
				return
			}
			autogold.Equal(t, got)
		})
	}
}

// TestExportedOnly_nested tests the behavior of Options.ExportedOnly when enabled with an unexported
// value/type nested below an exported one.
func TestExportedOnly_nested(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
		err   string
	}{
		{
			// TODO: bug: expect nil output
			name:  "external_struct_unexported_field_omitted",
			input: test.NewBaz(),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
		},
		/*
			{
				// TODO: BUG: nil pointer panic
				name: "struct_same_package_unexported_field_omitted",
				input: ExportedBaz{
					Bam: 1.34,
					zeta: foo{
						bar: "hello",
					},
				},
				err: "valast: cannot convert value of kind:struct type:valast.baz",
				opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
			},
		*/
		/*
			{
				// TODO: BUG: nil pointer panic
				name: "interface",
				input: struct {
					zeta foo
				}{zeta: foo{bar: "baz"}},
				opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast", ExportedOnly: true},
			},
		*/
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got, err := String(reflect.ValueOf(tst.input), tst.opt)
			if tst.err != "" && tst.err != err.Error() || tst.err == "" && err != nil {
				t.Fatal("\ngot:\n", err, "\nwant:\n", tst.err)
				return
			}
			autogold.Equal(t, got)
		})
	}
}

// TestUnexportedInputs tests the behavior of Options.ExportedOnly when disabled.
func TestUnexportedInputs(t *testing.T) {
	type (
		unexportedBool          bool
		unexportedInt           int
		unexportedInt8          int8
		unexportedInt16         int16
		unexportedInt32         int32
		unexportedInt64         int64
		unexportedUint          uint
		unexportedUint8         uint8
		unexportedUint16        uint16
		unexportedUint32        uint32
		unexportedUint64        uint64
		unexportedUintptr       uintptr
		unexportedFloat32       float32
		unexportedFloat64       float64
		unexportedComplex64     complex64
		unexportedComplex128    complex128
		unexportedArray         [1]float32
		unexportedInterface     error
		unexportedMap           map[string]string
		unexportedPointer       *int
		unexportedSlice         []int
		unexportedString        string
		unexportedStruct        struct{ A string }
		unexportedUnsafePointer unsafe.Pointer
	)
	tests := []struct {
		name  string
		input interface{}
		opt   *Options
		err   string
	}{
		{
			// TODO: BUG: not properly typed bool(true) vs. unexportedBool(true)
			name:  "bool",
			input: unexportedBool(true),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed int(1) vs. unexportedInt(1)
			name:  "int",
			input: unexportedInt(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed int8(1) vs. unexportedInt8(1)
			name:  "int8",
			input: unexportedInt8(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed int16(1) vs. unexportedInt16(1)
			name:  "int16",
			input: unexportedInt16(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed int32(1) vs. unexportedInt32(1)
			name:  "int32",
			input: unexportedInt32(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed int64(1) vs. unexportedInt64(1)
			name:  "int64",
			input: unexportedInt64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed uint(1) vs. unexportedUint(1)
			name:  "uint",
			input: unexportedUint(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed uint8(1) vs. unexportedUint8(1)
			name:  "uint8",
			input: unexportedUint8(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed uint16(1) vs. unexportedUint16(1)
			name:  "uint16",
			input: unexportedUint16(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed uint32(1) vs. unexportedUint32(1)
			name:  "uint32",
			input: unexportedUint32(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed uint64(1) vs. unexportedUint64(1)
			name:  "uint64",
			input: unexportedUint64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed uintptr(1) vs. unexportedUintptr(1)
			name:  "uintptr",
			input: unexportedUintptr(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed float32(1) vs. unexportedFloat32(1)
			name:  "float32",
			input: unexportedFloat32(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed float64(1) vs. unexportedFloat64(1)
			name:  "float64",
			input: unexportedFloat64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed complex64(1) vs. unexportedComplex64(1)
			name:  "complex64",
			input: unexportedComplex64(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: not properly typed complex128(1) vs. unexportedComplex128(1)
			name:  "complex128",
			input: unexportedComplex128(1),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: [1]float32{float32(1)} should be [1]float32{1}
			name:  "array",
			input: unexportedArray{1.0},
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		/*
			{
				// TODO: BUG: nil pointer panic
				name: "interface",
				input: struct {
					V unexportedInterface
				}{V: nil},
				opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
			},
		*/
		{
			// TODO: BUG: map[string]string{"a": "b"} should be unexportedMap{"a": "b"}
			name: "map",
			input: unexportedMap{
				"a": "b",
			},
			opt: &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: produces illegal &nil
			name:  "pointer",
			input: unexportedPointer(nil),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: []int{int(1), int(2), int(3)} should be []unexportedSlice{1, 2, 3}
			name:  "slice",
			input: unexportedSlice{1, 2, 3},
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		{
			// TODO: BUG: string("hello") should be unexportedString("hello")
			name:  "string",
			input: unexportedString("hello"),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
		/*
			{
				// TODO: BUG: nil pointer panic
				name:  "struct",
				input: unexportedStruct{A: "b"},
				opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
				err:   "valast: cannot convert value of kind:struct type:valast.unexportedStruct",
			},
		*/
		{
			// TODO: BUG: unsafe.Pointer(uintptr(0xdeadbeef)) should be unexportedUnsafePointer(unsafe.Pointer(uintptr(0xdeadbeef)))
			name:  "unsafe_pointer",
			input: unexportedUnsafePointer(unsafe.Pointer(uintptr(0xdeadbeef))),
			opt:   &Options{PackageName: "valast", PackagePath: "github.com/hexops/valast"},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got, err := String(reflect.ValueOf(tst.input), tst.opt)
			if tst.err != "" && tst.err != err.Error() || tst.err == "" && err != nil {
				t.Fatal("\ngot:\n", err, "\nwant:\n", tst.err)
				return
			}
			autogold.Equal(t, got)
		})
	}
}
