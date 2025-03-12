package calc

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// Tokenize - returns tokens from string
func Tokenize(expression string) ([]string, error) {
	var tokens []string
	var currentNumber strings.Builder

	expression = strings.Replace(expression, " ", "", -1)

	for _, ch := range expression {
		if unicode.IsDigit(ch) || ch == '.' {
			currentNumber.WriteRune(ch)
		} else {
			if currentNumber.Len() > 0 {
				tokens = append(tokens, currentNumber.String())
				currentNumber.Reset()
			}
			tokens = append(tokens, string(ch))
		}
	}

	if currentNumber.Len() > 0 {
		tokens = append(tokens, currentNumber.String())
	}

	return tokens, nil
}

// InfixToPostfix - transforms the list of tokens in infix to postfix
func InfixToPostfix(tokens []string) ([]string, error) {
	var output []string
	var stack []string

	precedence := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	for _, token := range tokens {
		switch {
		case IsNumber(token):
			output = append(output, token)
		case token == "(":
			stack = append(stack, token)
		case token == ")":
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 || stack[len(stack)-1] != "(" {
				return nil, errors.New("mismatched parentheses")
			}
			stack = stack[:len(stack)-1]
		default:
			for len(stack) > 0 && precedence[token] <= precedence[stack[len(stack)-1]] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, errors.New("mismatched parentheses")
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

// IsNumber - checks if string is number
func IsNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsOperator - check if string is allowed operator
func IsOperator(s string) bool {
	return s == "+" || s == "-" || s == "*" || s == "/"
}
