package main

import (
	"path/filepath"
)

type setLocalRule[T any] struct {
	pattern string
	val     T
}

// setLocalRules represents the rules for resolving the value of a particular
// option for a given directory. Since users will typically call `setlocal` to
// override only some options for a directory, the rules are assigned to each
// option separately instead of being shared among all options. Because it is
// possible for multiple rules (patterns) to match a directory, the first
// matching rule will be used.
type setLocalRules[T any] []setLocalRule[T]

func (rules *setLocalRules[T]) set(pattern string, val T) {
	_ = rules.update(pattern, val, func(v *T) error {
		*v = val
		return nil
	})
}

// update is a low-level function for updating rules created by `setlocal`. The
// corresponding rule is obtained, or created and appended if it does not exist.
// The rule is the updated by a callback function, which is required because the
// syntax `setlocal /path bool_option!` needs to know the current value in order
// to determine the new value.
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
