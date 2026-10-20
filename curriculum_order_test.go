package main

import (
	"strings"
	"testing"

	"gotut/curriculum"
)

func TestModuleOneTeachesTypesBeforeZeroValues(t *testing.T) {
	modules := curriculum.All()
	if len(modules) == 0 || len(modules[0].Topics) < 3 {
		t.Fatal("module 1 is missing its first topics")
	}

	variables := modules[0].Topics[1]
	zeroValues := modules[0].Topics[2]

	if variables.Title != "Variables and types" {
		t.Errorf("topic 2 title = %q, want %q", variables.Title, "Variables and types")
	}
	if zeroValues.Title != "Zero values, conversions and constants" {
		t.Errorf("topic 3 title = %q, want %q", zeroValues.Title, "Zero values, conversions and constants")
	}

	if strings.Contains(variables.Theory, "## Zero values") {
		t.Error("the zero-values section still lives in the variables topic")
	}
	if !strings.Contains(zeroValues.Theory, "## Zero values") {
		t.Error("the zero-values topic is missing its zero-values section")
	}
	if strings.Contains(zeroValues.Theory, "## iota") {
		t.Error("iota still lives in the zero-values/constants topic")
	}

	zeroAt := strings.Index(zeroValues.Theory, "## Zero values")
	conversionAt := strings.Index(zeroValues.Theory, "## Conversion")
	if zeroAt < 0 || conversionAt < 0 || zeroAt > conversionAt {
		t.Errorf("expected zero values before conversion (zero at %d, conversion at %d)", zeroAt, conversionAt)
	}
}

func TestIotaLivesWithSwitch(t *testing.T) {
	modules := curriculum.All()
	if len(modules) == 0 || len(modules[0].Topics) < 5 {
		t.Fatal("module 1 is missing its switch topic")
	}

	conditionals := modules[0].Topics[4]
	if !strings.Contains(conditionals.Theory, "## Named values with iota") {
		t.Fatal("the switch topic does not explain iota")
	}

	var found bool
	for _, ex := range conditionals.Exercises {
		if ex.ID == "m1t3e3" {
			found = true
			break
		}
	}
	if !found {
		t.Error("the iota exercise was not moved into the switch topic")
	}

	for _, ex := range modules[0].Topics[2].Exercises {
		if ex.ID == "m1t3e3" {
			t.Error("the iota exercise still lives in the constants topic")
		}
	}
}
