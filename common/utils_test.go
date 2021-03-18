package common

import (
	"log"
	"strings"
	"testing"
)

func TestLowFirstCase(t *testing.T) {
	input := []struct {
		In  string
		Out string
	}{
		{In: "Lower", Out: "lower"},
		{In: "First", Out: "first"},
		{In: "Test", Out: "test"},
		{In: "", Out: ""},
		{In: "T", Out: "t"},
		{In: "Te", Out: "te"},
	}
	for i := range input {
		result := LowFirstCase(input[i].In)
		if strings.Compare(result, input[i].Out) != 0 {
			log.Fatal("Result:", result, " not equal Out:", input[i].Out)
		} else {
			log.Println(input[i].In, result, input[i].Out)
		}
	}
}

func TestLowCasePaddingUnderline(t *testing.T) {
	input := []struct {
		In  string
		Out string
	}{
		{In: "Lower", Out: "lower"},
		{In: "First", Out: "first"},
		{In: "Test", Out: "test"},
		{In: "", Out: ""},
		{In: "T", Out: "t"},
		{In: "Te", Out: "te"},
		{In: "TeTeTe", Out: "te_te_te"},
	}
	for i := range input {
		result := LowCasePaddingUnderline(input[i].In)
		if strings.Compare(result, input[i].Out) != 0 {
			log.Fatal("Result:", result, " not equal Out:", input[i].Out)
		} else {
			log.Println(input[i].In, result, input[i].Out)
		}
	}
}

func TestGoModuleName(t *testing.T) {
	module := GoModuleName("F://go//src//stock")
	t.Log(module)
}
