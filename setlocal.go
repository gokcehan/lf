package main

import (
	"path/filepath"
)

type setLocalRule[T any] struct {
	pattern string
	val     T
}

type setLocalRules[T any] []setLocalRule[T]

func (rules *setLocalRules[T]) set(pattern string, val T) {
	rules.update(pattern, val, func(v *T) error {
		*v = val
		return nil
	})
}

func (rules *setLocalRules[T]) update(pattern string, initVal T, updater func(val *T) error) error {
	for i := range *rules {
		rule := &(*rules)[i]
		if rule.pattern == pattern {
			return updater(&rule.val)
		}
	}

	rule := setLocalRule[T]{pattern, initVal}
	if err := updater(&rule.val); err != nil {
		return err
	}

	*rules = append(*rules, rule)
	return nil
}

func (rules *setLocalRules[T]) get(path string, fallback T) T {
	for i := range *rules {
		rule := &(*rules)[i]
		if ok, err := filepath.Match(rule.pattern, path); err == nil && ok {
			return rule.val
		}
	}

	return fallback
}
