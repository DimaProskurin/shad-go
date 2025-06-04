//go:build !solution

package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Function []func(stack *[]int) error

type Evaluator struct {
	stack     []int
	functions map[string]Function
}

// NewEvaluator creates evaluator.
func NewEvaluator() *Evaluator {
	functions := map[string]Function{
		"+": []func(*[]int) error{
			func(stack *[]int) error {
				snd, err := popStack(stack)
				if err != nil {
					return err
				}
				fst, err := popStack(stack)
				if err != nil {
					return err
				}
				*stack = append(*stack, fst+snd)
				return nil
			},
		},
		"-": []func(*[]int) error{
			func(stack *[]int) error {
				snd, err := popStack(stack)
				if err != nil {
					return err
				}
				fst, err := popStack(stack)
				if err != nil {
					return err
				}
				*stack = append(*stack, fst-snd)
				return nil
			},
		},
		"*": []func(*[]int) error{
			func(stack *[]int) error {
				snd, err := popStack(stack)
				if err != nil {
					return err
				}
				fst, err := popStack(stack)
				if err != nil {
					return err
				}
				*stack = append(*stack, fst*snd)
				return nil
			},
		},
		"/": []func(*[]int) error{
			func(stack *[]int) error {
				snd, err := popStack(stack)
				if err != nil {
					return err
				}
				fst, err := popStack(stack)
				if err != nil {
					return err
				}
				if snd == 0 {
					return fmt.Errorf("division by zero")
				}
				*stack = append(*stack, fst/snd)
				return nil
			},
		},
		"over": []func(*[]int) error{
			func(stack *[]int) error {
				snd, err := popStack(stack)
				if err != nil {
					return err
				}
				fst, err := popStack(stack)
				if err != nil {
					return err
				}
				*stack = append(*stack, fst, snd, fst)
				return nil
			},
		},
		"swap": []func(*[]int) error{
			func(stack *[]int) error {
				snd, err := popStack(stack)
				if err != nil {
					return err
				}
				fst, err := popStack(stack)
				if err != nil {
					return err
				}
				*stack = append(*stack, snd, fst)
				return nil
			},
		},
		"dup": []func(*[]int) error{
			func(stack *[]int) error {
				fst, err := popStack(stack)
				if err != nil {
					return err
				}
				*stack = append(*stack, fst, fst)
				return nil
			},
		},
		"drop": []func(*[]int) error{
			func(stack *[]int) error {
				_, err := popStack(stack)
				return err
			},
		},
	}
	return &Evaluator{
		stack:     make([]int, 0),
		functions: functions,
	}
}

// Process evaluates sequence of words or definition.
//
// Returns resulting stack state and an error.
func (e *Evaluator) Process(row string) ([]int, error) {
	if row[0] == ':' {
		defRowParts := strings.Split(row[2:len(row)-2], " ")
		wordName := strings.ToLower(defRowParts[0])
		definition := strings.Join(defRowParts[1:], " ")

		if !isCorrectWord(wordName) {
			return nil, fmt.Errorf("incorrect word name %s", wordName)
		}

		fn, err := e.evalDefinition(definition)
		if err != nil {
			return nil, err
		}
		e.functions[wordName] = fn

		return e.stack, nil
	}

	rowParts := strings.Split(row, " ")
	for _, part := range rowParts {
		partInt, err := strconv.Atoi(part)
		if err == nil {
			e.stack = append(e.stack, partInt)
			continue
		}

		part = strings.ToLower(part)
		fn, exists := e.functions[part]
		if !exists {
			return nil, fmt.Errorf("no such function %s", part)
		}

		for _, op := range fn {
			err = op(&e.stack)
			if err != nil {
				return nil, err
			}
		}
	}
	return e.stack, nil
}

func (e *Evaluator) evalDefinition(def string) (Function, error) {
	ops := make([]func(*[]int) error, 0)
	for _, part := range strings.Split(def, " ") {
		part = strings.ToLower(part)
		if fn, exists := e.functions[part]; exists {
			ops = append(ops, fn...)
			continue
		}
		partInt, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		ops = append(ops, func(stack *[]int) error {
			*stack = append(*stack, partInt)
			return nil
		})
	}
	return ops, nil
}

func popStack(stack *[]int) (int, error) {
	if stack == nil || len(*stack) == 0 {
		return 0, fmt.Errorf("can't pop, stack is empty")
	}
	last := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return last, nil
}

func isCorrectWord(s string) bool {
	_, err := strconv.Atoi(s)
	return err != nil
}
