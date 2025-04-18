package util

import (
	"testing"
)

func TestUnitQuickHashFields(t *testing.T) {
	tests := []struct {
		name    string
		values  []interface{}
		wantErr bool
	}{
		{
			name:    "No inputs",
			values:  []interface{}{},
			wantErr: true,
		},
		{
			name:    "Single string input",
			values:  []interface{}{"test"},
			wantErr: false,
		},
		{
			name:    "Single nil input",
			values:  []interface{}{nil},
			wantErr: false,
		},
		{
			name:    "Multiple string inputs",
			values:  []interface{}{"test1", "test2", "test3"},
			wantErr: false,
		},
		{
			name:    "Mixed type inputs",
			values:  []interface{}{"test", 123, true},
			wantErr: false,
		},
		{
			name:    "Multiple nil inputs",
			values:  []interface{}{nil, nil, nil},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QuickHashFields(tt.values...)

			if (err != nil) != tt.wantErr {
				t.Errorf("QuickHashFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
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
	}

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
			expectedHash: "72bb3bf6e605a43a",
		},
		{
			name: "Multiple Struct input",
			value: []interface{}{
				testStruct{Field1: "test", Field2: 123},
				testStruct{Field1: "test2", Field2: 456},
			},
			wantErr:      false,
			expectedHash: "a3eb3d4830f9153a",
		},
		{
			name:         "Map input",
			value:        []interface{}{map[string]string{"key": "value"}},
			wantErr:      false,
			expectedHash: "aaaaa8b65ed3ff7b",
		},
		{
			name: "Multiple Map input",
			value: []interface{}{
				map[string]string{"key": "value"},
				map[string]string{"foo": "bar"},
			},
			wantErr:      false,
			expectedHash: "fe443377ca878d7d",
		},
		{
			name:         "Slice input",
			value:        []interface{}{[]string{"test1", "test2"}},
			wantErr:      false,
			expectedHash: "0c1efb7496fa5b19",
		},
		{
			name: "Multiple Slice input",
			value: []interface{}{
				[]string{"test1", "test2"},
				[]string{"test3", "test4"},
				[]string{"test2", "test1"},
			},
			wantErr:      false,
			expectedHash: "0590d72ff80ff5e3",
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
			expectedHash: "e63ec7734b9ca5fb",
		},
		{
			name: "Multiple Mix of Inputs with some as nil",
			value: []interface{}{
				[]string{"test1", "test2"},
				map[string]string{"foo": "bar"},
				[]string{"test3", "test4"},
				nil,
				testStruct{Field1: "test5", Field2: 987},
				nil,
			},
			wantErr:      false,
			expectedHash: "e63ec7734b9ca5fb",
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
