package jsonmask_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/axkit/jsonmask"
	"github.com/stretchr/testify/assert"
)

type TestHiddenAttr struct {
	ID     int   `json:"id"`
	Amount int64 `json:"amount" mask:"-"`
}

type TestStructMaskAttr struct {
	ID         int    `json:",omitempty"`
	Currency   string `json:"currency" mask:"upper"`
	MinorUnits int64  `json:"minorUnits" mask:"zero"`
	flag       bool
}

type TestEmbeddedStruct struct {
	Value int `json:"value"`
}

type TestStruct struct {
	Basic struct {
		ID                 int    `json:"id"`
		FirstName          string `json:"firstName" mask:"initialChar"`
		LastName           string `json:"lastName" mask:"upper"`
		BirthDate          string `json:"birthDate" mask:"-"`
		TestEmbeddedStruct `json:"embedded"`
		HiddenEmbedded     TestEmbeddedStruct `json:"hiddenEmbedded" mask:"-"`
		TestHiddenAttr     `json:"hiddenAttr" mask:"-"`
	}

	BasicPtr struct {
		ID                  *int    `json:"id"`
		FirstName           *string `json:"firstName" mask:"initialChar"`
		LastName            *string `json:"lastName" mask:"upper"`
		BirthDate           *string `json:"birthDate" mask:"-"`
		*TestEmbeddedStruct `json:"embedded"`
		HiddenEmbedded      *TestEmbeddedStruct `json:"hiddenEmbedded" mask:"-"`
		TestHiddenAttr      `json:"hiddenAttr" mask:"-"`
	}

	DoublePtr struct {
		ID             **int                `json:"id"`
		FirstName      **string             `json:"firstName" mask:"initialChar"`
		LastName       **string             `json:"lastName" mask:"upper"`
		BirthDate      **string             `json:"birthDate" mask:"-"`
		Embedded       **TestEmbeddedStruct `json:"embedded"`
		HiddenEmbedded **TestEmbeddedStruct `json:"hiddenEmbedded" mask:"-"`
	}

	Array struct {
		ID          int                   `json:"id"`
		Items       [3]TestStructMaskAttr `json:"items"`
		HiddenItems [3]TestHiddenAttr     `json:"hiddenItems" mask:"-"`
	}

	ArrayPtr struct {
		ID          int                    `json:"id"`
		Items       *[3]TestStructMaskAttr `json:"items"`
		HiddenItems *[3]TestHiddenAttr     `json:"hiddenItems" mask:"-"`
	}

	Slice struct {
		ID          int                  `json:"id"`
		Items       []TestStructMaskAttr `json:"items"`
		HiddenItems []TestHiddenAttr     `json:"hiddenItems" mask:"-"`
	}

	SlicePtr struct {
		ID          int                   `json:"id"`
		Items       *[]TestStructMaskAttr `json:"items"`
		HiddenItems *[]TestHiddenAttr     `json:"hiddenItems" mask:"-"`
	}

	Matrix struct {
		ID          int                    `json:"id"`
		Items       [][]TestStructMaskAttr `json:"items"`
		HiddenItems [][]TestHiddenAttr     `json:"hiddenItems" mask:"-"`
	}

	MatrixPtr struct {
		ID          int                     `json:"id"`
		Items       *[][]TestStructMaskAttr `json:"items"`
		HiddenItems *[][]TestHiddenAttr     `json:"hiddenItems" mask:"-"`
	}
}

func TestJsonMaskImpl_AddFunc(t *testing.T) {
	jm := jsonmask.New()

	result, err := jm.Mask(
		[]byte(`{"name":"john","balance":{"currency":"usd"}}`),
		jsonmask.StructMaskRules{
			Rules: []jsonmask.Rule{
				{Path: "name", Action: "initialChar"},
				{Path: "balance.currency", Action: "upper"},
			}})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"J","balance":{"currency":"USD"}}`, string(result))
}

func TestJsonMaskerImpl_ParseStruct(t *testing.T) {

	var s TestStruct
	jm := jsonmask.NewWithMaskTag("mask")

	t.Run("Basic", func(t *testing.T) {
		fields := jm.ParseStruct(s.Basic)
		assert.Len(t, fields.Rules, 5)
		checkRule(t, fields.Rules, 0, "firstName", "initialChar")
		checkRule(t, fields.Rules, 1, "lastName", "upper")
		checkRule(t, fields.Rules, 2, "birthDate", "-")
		checkRule(t, fields.Rules, 3, "hiddenEmbedded", "-")
		checkRule(t, fields.Rules, 4, "hiddenAttr", "-")
	})

	t.Run("BasicPtr", func(t *testing.T) {
		fields := jm.ParseStruct(s.BasicPtr)
		assert.Len(t, fields.Rules, 5)
		checkRule(t, fields.Rules, 0, "firstName", "initialChar")
		checkRule(t, fields.Rules, 1, "lastName", "upper")
		checkRule(t, fields.Rules, 2, "birthDate", "-")
		checkRule(t, fields.Rules, 3, "hiddenEmbedded", "-")
		checkRule(t, fields.Rules, 4, "hiddenAttr", "-")
	})

	t.Run("DoublePtr", func(t *testing.T) {
		fields := jm.ParseStruct(s.DoublePtr)
		assert.Len(t, fields.Rules, 4)
		checkRule(t, fields.Rules, 0, "firstName", "initialChar")
		checkRule(t, fields.Rules, 1, "lastName", "upper")
		checkRule(t, fields.Rules, 2, "birthDate", "-")
		checkRule(t, fields.Rules, 3, "hiddenEmbedded", "-")
	})

	t.Run("Array", func(t *testing.T) {
		fields := jm.ParseStruct(s.Array)
		assert.Len(t, fields.Rules, 3)
		checkRule(t, fields.Rules, 0, "items.#.currency", "upper")
		checkRule(t, fields.Rules, 1, "items.#.minorUnits", "zero")
		checkRule(t, fields.Rules, 2, "hiddenItems", "-")
	})

	t.Run("ArrayPtr", func(t *testing.T) {
		fields := jm.ParseStruct(s.ArrayPtr)
		assert.Len(t, fields.Rules, 3)
		checkRule(t, fields.Rules, 0, "items.#.currency", "upper")
		checkRule(t, fields.Rules, 1, "items.#.minorUnits", "zero")
		checkRule(t, fields.Rules, 2, "hiddenItems", "-")
	})

	t.Run("Slice", func(t *testing.T) {
		fields := jm.ParseStruct(s.Slice)
		assert.Len(t, fields.Rules, 3)
		checkRule(t, fields.Rules, 0, "items.#.currency", "upper")
		checkRule(t, fields.Rules, 1, "items.#.minorUnits", "zero")
		checkRule(t, fields.Rules, 2, "hiddenItems", "-")
	})

	t.Run("SlicePtr", func(t *testing.T) {
		fields := jm.ParseStruct(s.SlicePtr)
		assert.Len(t, fields.Rules, 3)
		checkRule(t, fields.Rules, 0, "items.#.currency", "upper")
		checkRule(t, fields.Rules, 1, "items.#.minorUnits", "zero")
		checkRule(t, fields.Rules, 2, "hiddenItems", "-")
	})

	t.Run("Matrix", func(t *testing.T) {
		fields := jm.ParseStruct(s.Matrix)
		assert.Len(t, fields.Rules, 3)
		checkRule(t, fields.Rules, 0, "items.#.#.currency", "upper")
		checkRule(t, fields.Rules, 1, "items.#.#.minorUnits", "zero")
		checkRule(t, fields.Rules, 2, "hiddenItems", "-")
	})
}

func checkRule(t *testing.T, rules []jsonmask.Rule, index int, path, action string) {
	t.Helper()
	assert.Equal(t, path, rules[index].Path)
	assert.Equal(t, action, rules[index].Action)
}

func compareBytes(t *testing.T, expected, actual []byte) {
	t.Helper()
	if !assert.True(t, bytes.Equal(expected, actual)) {
		for i, b := range actual {
			if b != expected[i] {
				t.Errorf("byte pos: %d\ngot: %s\nexp: %s", i, string(actual), string(expected))
				break
			}
		}
	}
}

func TestMask(t *testing.T) {
	var src TestStruct

	jm := jsonmask.New()

	t.Run("NonStructType", func(t *testing.T) {
		parsed := jm.ParseStruct(new(int))
		if parsed.Rules != nil {
			t.Error("Expected nil")
		}
	})

	t.Run("Basic", func(t *testing.T) {
		src.Basic.ID = 1
		src.Basic.FirstName = "john"
		src.Basic.LastName = "doe"
		src.Basic.BirthDate = "2000-01-01"
		src.Basic.TestEmbeddedStruct.Value = 2
		src.Basic.HiddenEmbedded.Value = 3
		src.Basic.TestHiddenAttr.ID = 4
		src.Basic.TestHiddenAttr.Amount = 100

		jsonData, err := json.Marshal(src.Basic)
		assert.NoError(t, err)
		parsed := jm.ParseStruct(&src.Basic)

		result, err := jm.Mask(jsonData, parsed)
		assert.NoError(t, err)

		expected := []byte(`{"id":1,"firstName":"J","lastName":"DOE","embedded":{"value":2}}`)
		compareBytes(t, expected, result)
	})

	t.Run("SlicePtr", func(t *testing.T) {
		src.SlicePtr.ID = 1
		src.SlicePtr.Items = &[]TestStructMaskAttr{
			{ID: 1, Currency: "usd"},
			{ID: 2, Currency: "eur"},
		}
		src.SlicePtr.HiddenItems = &[]TestHiddenAttr{
			{ID: 1, Amount: 100},
			{ID: 2, Amount: 200},
			{ID: 3, Amount: 300},
		}

		jsonData, err := json.Marshal(src.SlicePtr)
		assert.NoError(t, err)
		parsed := jm.ParseStruct(src.SlicePtr)

		result, err := jm.Mask(jsonData, parsed)
		assert.NoError(t, err)

		expected := []byte(`{"id":1,"items":[{"ID":1,"currency":"USD","minorUnits":0},{"ID":2,"currency":"EUR","minorUnits":0}]}`)
		compareBytes(t, expected, result)
	})

	t.Run("Array", func(t *testing.T) {
		src.Array.ID = 1
		src.Array.Items = [3]TestStructMaskAttr{
			{ID: 1, Currency: "usd"},
			{ID: 2, Currency: "eur"},
			{ID: 3, Currency: "czk"},
		}
		src.Array.HiddenItems = [3]TestHiddenAttr{
			{ID: 1, Amount: 100},
			{ID: 2, Amount: 200},
			{ID: 3, Amount: 300},
		}

		jsonData, err := json.Marshal(src.Array)
		assert.NoError(t, err)
		parsed := jm.ParseStruct(src.Array)

		result, err := jm.Mask(jsonData, parsed)
		assert.NoError(t, err)

		expected := []byte(`{"id":1,"items":[{"ID":1,"currency":"USD","minorUnits":0},{"ID":2,"currency":"EUR","minorUnits":0},{"ID":3,"currency":"CZK","minorUnits":0}]}`)
		compareBytes(t, expected, result)
	})

	t.Run("Matrix", func(t *testing.T) {
		src.Matrix.ID = 1
		src.Matrix.Items = [][]TestStructMaskAttr{
			{{ID: 1, Currency: "usd"}, {ID: 2, Currency: "eur"}},
			{{ID: 3, Currency: "czk"}, {ID: 4, Currency: "gbp"}},
		}
		src.Matrix.HiddenItems = [][]TestHiddenAttr{
			{{ID: 1, Amount: 100}, {ID: 2, Amount: 200}},
			{{ID: 3, Amount: 300}, {ID: 4, Amount: 400}},
		}

		jsonData, err := json.Marshal(src.Matrix)
		assert.NoError(t, err)
		parsed := jm.ParseStruct(src.Matrix)

		result, err := jm.Mask(jsonData, parsed)
		assert.NoError(t, err)

		expected := []byte(`{"id":1,"items":[[{"ID":1,"currency":"USD","minorUnits":0},{"ID":2,"currency":"EUR","minorUnits":0}],[{"ID":3,"currency":"CZK","minorUnits":0},{"ID":4,"currency":"GBP","minorUnits":0}]]}`)
		compareBytes(t, expected, result)
	})

}
