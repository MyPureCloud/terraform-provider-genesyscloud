package util

import (
	"testing"
)

func TestUnitQuickHashFields(t *testing.T) {
	str := "test"
	strPtr := &str
	var nilStrPtr *string
	num := 123
	numPtr := &num

	tests := []struct {
		name     string
		values   []interface{}
		wantErr  bool
		wantHash bool
	}{
		{
			name:     "No inputs",
			values:   []interface{}{},
			wantErr:  true,
			wantHash: false,
		},
		{
			name:     "Single string input",
			values:   []interface{}{"test"},
			wantErr:  false,
			wantHash: true,
		},
		{
			name:     "Single nil input",
			values:   []interface{}{nil},
			wantErr:  false,
			wantHash: false,
		},
		{
			name:     "String pointer input",
			values:   []interface{}{strPtr},
			wantErr:  false,
			wantHash: true,
		},
		{
			name:     "Nil string pointer input",
			values:   []interface{}{nilStrPtr},
			wantErr:  false,
			wantHash: false,
		},
		{
			name:     "Single int input",
			values:   []interface{}{123},
			wantErr:  false,
			wantHash: true,
		},
		{
			name:     "Single int pointer input",
			values:   []interface{}{numPtr},
			wantErr:  false,
			wantHash: true,
		},
		{
			name:     "Multiple string inputs",
			values:   []interface{}{"test1", "test2", "test3"},
			wantErr:  false,
			wantHash: true,
		},
		{
			name:     "Mixed type inputs",
			values:   []interface{}{"test", 123, true},
			wantErr:  false,
			wantHash: true,
		},
		{
			name:     "Multiple nil inputs",
			values:   []interface{}{nil, nil, nil},
			wantErr:  false,
			wantHash: false,
		},
		{
			name: "Mixed pointer and value types",
			values: []interface{}{
				"test",
				str,
				num,
				nilStrPtr,
				nil,
			},
			wantErr:  false,
			wantHash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QuickHashFields(tt.values...)

			if (err != nil) != tt.wantErr {
				t.Errorf("QuickHashFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantHash {
				// Verify hash is exactly 16 characters
				if len(got) != 16 {
					t.Errorf("QuickHashFields() returned hash of length %d, want 16", len(got))
				}

				// Verify consistent hashing
				got2, _ := QuickHashFields(tt.values...)
				if got != got2 {
					t.Errorf("QuickHashFields() not consistent: %v != %v", got, got2)
				}
			}
		})
	}
}

func TestUnitQuickHashFields_DifferentInputs(t *testing.T) {
	hash1, _ := QuickHashFields("test1")
	hash2, _ := QuickHashFields("test2")

	if hash1 == hash2 {
		t.Errorf("Different inputs produced same hash: %v", hash1)
	}
}

func TestUnitQuickHashFields_ComplexTypes(t *testing.T) {
	type testStruct struct {
		Field1 string
		Field2 int
		Field3 *string
	}

	str := "test"
	var nilStr *string

	tests := []struct {
		name         string
		value        []interface{}
		wantErr      bool
		expectedHash string
	}{
		{
			name: "Struct input",
			value: []interface{}{
				testStruct{Field1: "test", Field2: 123},
			},
			wantErr:      false,
			expectedHash: "ef1a14bd9bd591c3",
		},
		{
			name: "Struct input with nil pointer",
			value: []interface{}{
				testStruct{Field1: "test", Field2: 123, Field3: nilStr},
			},
			wantErr:      false,
			expectedHash: "ef1a14bd9bd591c3",
		},
		{
			name: "Struct input with valid pointer",
			value: []interface{}{
				testStruct{Field1: "test", Field2: 123, Field3: &str},
			},
			wantErr:      false,
			expectedHash: "caf965b7315c3dd6",
		},
		{
			name: "Multiple Struct input",
			value: []interface{}{
				testStruct{Field1: "test", Field2: 123},
				testStruct{Field1: "test2", Field2: 456},
			},
			wantErr:      false,
			expectedHash: "f901b0dc6747c4c3",
		},
		{
			name:         "Map input",
			value:        []interface{}{map[string]string{"key": "value"}},
			wantErr:      false,
			expectedHash: "cbdea9ab8317fcd1",
		},
		{
			name: "Multiple Map input",
			value: []interface{}{
				map[string]string{"key": "value"},
				map[string]string{"foo": "bar"},
			},
			wantErr:      false,
			expectedHash: "92422f54645bb2d7",
		},
		{
			name: "Map with pointer values",
			value: []interface{}{
				map[string]*string{"key": &str, "nil_key": nilStr},
			},
			wantErr:      false,
			expectedHash: "00b5d32006713baa",
		},
		{
			name:         "Slice input",
			value:        []interface{}{[]string{"test1", "test2"}},
			wantErr:      false,
			expectedHash: "45a7a10579268d62",
		},
		{
			name: "Slice with pointer values",
			value: []interface{}{
				[]*string{&str, nilStr, &str},
			},
			wantErr:      false,
			expectedHash: "1492c421fc01e363",
		},
		{
			name: "Multiple Slice input",
			value: []interface{}{
				[]string{"test1", "test2"},
				[]string{"test3", "test4"},
				[]string{"test2", "test1"},
			},
			wantErr:      false,
			expectedHash: "758cf2b128d714ab",
		},
		{
			name: "Multiple Mix of Inputs",
			value: []interface{}{
				[]string{"test1", "test2"},
				map[string]string{"foo": "bar"},
				[]string{"test3", "test4"},
				testStruct{Field1: "test5", Field2: 987},
			},
			wantErr:      false,
			expectedHash: "17f7261250799f2d",
		},
		{
			name: "Mixed complex types with nil pointers",
			value: []interface{}{
				testStruct{Field1: "test", Field2: 123, Field3: nilStr},
				map[string]*string{"key": nilStr, "nil_key": nilStr},
				[]*string{nilStr, nilStr, nilStr},
				[]*string{&str, nilStr, &str},
			},
			wantErr:      false,
			expectedHash: "84e42603d87047fe",
		},
		{
			name: "Multiple mixed types with some nils",
			value: []interface{}{
				[]string{"test1", "test2"},
				map[string]string{"foo": "bar"},
				[]string{"test3", "test4"},
				nil,
				testStruct{Field1: "test5", Field2: 987},
				nil,
			},
			wantErr:      false,
			expectedHash: "17f7261250799f2d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QuickHashFields(tt.value...)

			if (err != nil) != tt.wantErr {
				t.Errorf("QuickHashFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != 16 {
				t.Errorf("QuickHashFields() returned hash of length %d, want 16", len(got))
			}

			if tt.expectedHash != got {
				t.Errorf("QuickHashFields() returned hash %v, want %v", got, tt.expectedHash)
			}
		})
	}
}
