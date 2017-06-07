package dao

import (
	"encoding/json"
	"io/ioutil"
)

type FieldType string

const (
	FieldTypeText      FieldType = "TEXT"
	FieldTypeBool      FieldType = "BOOL"
	FieldTypeString    FieldType = "STRING"
	FieldTypeDate      FieldType = "DATE"
	FieldTypeInt64     FieldType = "INT64"
	FieldTypeInt       FieldType = "INT"
	FieldTypeFloat64   FieldType = "FLOAT64"
	FieldTypeReference FieldType = "REFERENCE"
)

type TypeDesc struct {
	Type     *Type
	EditView *EditView
	ListView *ListView
}

type Type struct {
	Name      string
	IndexName string
	Kind      string
	Fields    []*TypeField
}

type TypeField struct {
	Id        string
	Name      string
	Multiple  bool
	FieldType FieldType
}

type EditView struct {
	Widgets []*FormWidget `json:"widgets"`
}

type ListView struct {
	Title   string            `json:"title"`
	Sort    string            `json:"sort"`
	Columns []*ListViewColumn `json:"columns"`
}

type ListViewColumn struct {
	Id       string `json:"id"`
	Path     string `json:"path,omitempty"`
	Width    string `json:"width,omitempty"`
	Sortable bool   `json:"sortable,omitempty"`
	Function string `json:"function,omitempty"`
	Align    string `json:"align,omitempty"`
}

type FormWidget struct {
	Id           string      `json:"id"`
	Title        string      `json:"title"`
	Type         string      `json:"type"`
	DefaultValue string      `json:"defaultValue"`
	DataSource   *DataSource `json:"dataSource"`
}

type DataSource struct {
	EmptyRow   bool   `json:"emptyRow"`
	ColumnName string `json:"columnName"`
	Kind       string `json:"kind"`
}

type FieldValuePair struct {
	Widget *FormWidget
	Value  string
}

func NewTypeDesc(filename string) *TypeDesc {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var typeDesc TypeDesc
	err = json.Unmarshal(b, &typeDesc)
	if err != nil {
		panic(err)
	}
	return &typeDesc
}
