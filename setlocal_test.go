package main

import (
	"reflect"
	"testing"
)

func TestSetLocalRuleSet(t *testing.T) {
	tests := []struct {
		rules   setLocalRules[int]
		pattern string
		val     int
		exp     setLocalRules[int]
	}{
		{setLocalRules[int]{}, "foo", 1, setLocalRules[int]{{"foo", 1}}},
		{setLocalRules[int]{{"foo", 1}}, "bar", 2, setLocalRules[int]{{"foo", 1}, {"bar", 2}}},
		{setLocalRules[int]{{"foo", 1}}, "foo", 2, setLocalRules[int]{{"foo", 2}}},
		{setLocalRules[int]{{"foo", 1}, {"bar", 2}}, "foo", 3, setLocalRules[int]{{"foo", 3}, {"bar", 2}}},
	}

	for _, test := range tests {
		if test.rules.set(test.pattern, test.val); !reflect.DeepEqual(test.rules, test.exp) {
			t.Errorf("at test '%v' expected '%v' but got '%v'", test, test.exp, test.rules)
		}
	}
}

func TestSetLocalRuleUpdate(t *testing.T) {
	tests := []struct {
		rules   setLocalRules[int]
		pattern string
		exp     setLocalRules[int]
	}{
		{setLocalRules[int]{}, "foo", setLocalRules[int]{{"foo", -10}}},
		{setLocalRules[int]{{"foo", 1}}, "bar", setLocalRules[int]{{"foo", 1}, {"bar", -10}}},
		{setLocalRules[int]{{"foo", 1}}, "foo", setLocalRules[int]{{"foo", -1}}},
		{setLocalRules[int]{{"foo", 1}, {"bar", 2}}, "foo", setLocalRules[int]{{"foo", -1}, {"bar", 2}}},
	}

	updater := func(val *int) error {
		*val *= -1
		return nil
	}

	for _, test := range tests {
		if _ = test.rules.update(test.pattern, 10, updater); !reflect.DeepEqual(test.rules, test.exp) {
			t.Errorf("at test '%v' expected '%v' but got '%v'", test, test.exp, test.rules)
		}
	}
}

func TestSetLocalRuleGet(t *testing.T) {
	tests := []struct {
		rules setLocalRules[int]
		path  string
		exp   int
	}{
		{setLocalRules[int]{}, "foo", -1},
		{setLocalRules[int]{{"foo", 1}, {"bar", 2}}, "foo", 1},
		{setLocalRules[int]{{"foo", 1}, {"bar", 2}}, "bar", 2},
		{setLocalRules[int]{{"bar", 1}, {"b*", 2}}, "bar", 1},
		{setLocalRules[int]{{"bar", 1}, {"b*", 2}}, "baz", 2},
	}

	for _, test := range tests {
		if got := test.rules.get(test.path, -1); got != test.exp {
			t.Errorf("at test '%v' expected '%v' but got '%v'", test, test.exp, got)
		}
	}
}
