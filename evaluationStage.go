package govaluate

import (
	"errors"
	"fmt"
	"math"
	"regexp"
)

const (
	TYPEERROR_LOGICAL    string = "Value '%v' cannot be used with the logical operator '%v', it is not a bool"
	TYPEERROR_MODIFIER   string = "Value '%v' cannot be used with the modifier '%v', it is not a number"
	TYPEERROR_COMPARATOR string = "Value '%v' cannot be used with the comparator '%v', it is not a number"
	TYPEERROR_TERNARY    string = "Value '%v' cannot be used with the ternary operator '%v', it is not a bool"
	TYPEERROR_PREFIX     string = "Value '%v' cannot be used with the prefix '%v'"
)

type evaluationOperator func(left interface{}, right interface{}, parameters Parameters) (interface{}, error)
type stageTypeCheck func(value interface{}) bool
type stageCombinedTypeCheck func(left interface{}, right interface{}) bool

type evaluationStage struct {
	symbol OperatorSymbol

	leftStage, rightStage *evaluationStage

	// the operation that will be used to evaluate this stage (such as adding [left] to [right] and return the result)
	operator evaluationOperator

	// ensures that both left and right values are appropriate for this stage. Returns an error if they aren't operable.
	leftTypeCheck  stageTypeCheck
	rightTypeCheck stageTypeCheck

	// if specified, will override whatever is used in "leftTypeCheck" and "rightTypeCheck".
	// primarily used for specific operators that don't care which side a given type is on, but still requires one side to be of a given type
	// (like string concat)
	typeCheck stageCombinedTypeCheck

	// regardless of which type check is used, this string format will be used as the error message for type errors
	typeErrorFormat string
}

func (this *evaluationStage) swapWith(other *evaluationStage) {

	temp := *other
	other.setToNonStage(*this)
	this.setToNonStage(temp)
}

func (this *evaluationStage) setToNonStage(other evaluationStage) {

	this.symbol = other.symbol
	this.operator = other.operator
	this.leftTypeCheck = other.leftTypeCheck
	this.rightTypeCheck = other.rightTypeCheck
	this.typeCheck = other.typeCheck
	this.typeErrorFormat = other.typeErrorFormat
}

func addStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {

	// string concat if either are strings
	if isString(left) || isString(right) {
		return fmt.Sprintf("%v%v", left, right), nil
	}

	return left.(float64) + right.(float64), nil
}
func subtractStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left.(float64) - right.(float64), nil
}
func multiplyStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left.(float64) * right.(float64), nil
}
func divideStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left.(float64) / right.(float64), nil
}
func exponentStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return math.Pow(left.(float64), right.(float64)), nil
}
func modulusStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return math.Mod(left.(float64), right.(float64)), nil
}
func gteStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left.(float64) >= right.(float64), nil
}
func gtStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left.(float64) > right.(float64), nil
}
func lteStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left.(float64) <= right.(float64), nil
}
func ltStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left.(float64) < right.(float64), nil
}
func equalStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left == right, nil
}
func notEqualStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left != right, nil
}
func andStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left.(bool) && right.(bool), nil
}
func orStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return left.(bool) || right.(bool), nil
}
func negateStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return -right.(float64), nil
}
func invertStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return !right.(bool), nil
}
func bitwiseNotStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return float64(^int64(right.(float64))), nil
}
func ternaryIfStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	if left.(bool) {
		return right, nil
	}
	return nil, nil
}
func ternaryElseStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	if left != nil {
		return left, nil
	}
	return right, nil
}

func regexStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {

	var pattern *regexp.Regexp
	var err error

	switch right.(type) {
	case string:
		pattern, err = regexp.Compile(right.(string))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Unable to compile regexp pattern '%v': %v", right, err))
		}
	case *regexp.Regexp:
		pattern = right.(*regexp.Regexp)
	}

	return pattern.Match([]byte(left.(string))), nil
}

func notRegexStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {

	ret, err := regexStage(left, right, parameters)
	if err != nil {
		return nil, err
	}

	return !(ret.(bool)), nil
}

func bitwiseOrStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return float64(int64(left.(float64)) | int64(right.(float64))), nil
}
func bitwiseAndStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return float64(int64(left.(float64)) & int64(right.(float64))), nil
}
func bitwiseXORStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return float64(int64(left.(float64)) ^ int64(right.(float64))), nil
}
func leftShiftStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return float64(uint64(left.(float64)) << uint64(right.(float64))), nil
}
func rightShiftStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return float64(uint64(left.(float64)) >> uint64(right.(float64))), nil
}

func makeParameterStage(parameterName string) evaluationOperator {

	return func(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
		value, err := parameters.Get(parameterName)
		if err != nil {
			return nil, err
		}

		return value, nil
	}
}

func makeLiteralStage(literal interface{}) evaluationOperator {
	return func(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
		return literal, nil
	}
}

func makeFunctionStage(function ExpressionFunction) evaluationOperator {

	return func(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {

		if(right == nil) {
			return function()
		}

		switch right.(type) {
		case []interface{}:
			return function(right.([]interface{})...)
		default:
			return function(right)
		}

	}
}

func separatorStage(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {

	var ret []interface{}

	if(left != nil) {
		ret = []interface{} {left}
	}

	switch right.(type) {
	case []interface{}:
		ret = append(ret, left)
	}

	return ret, nil
}

//

func isString(value interface{}) bool {

	switch value.(type) {
	case string:
		return true
	}
	return false
}

func isRegexOrString(value interface{}) bool {

	switch value.(type) {
	case string:
		return true
	case *regexp.Regexp:
		return true
	}
	return false
}

func isBool(value interface{}) bool {
	switch value.(type) {
	case bool:
		return true
	}
	return false
}

func isFloat64(value interface{}) bool {
	switch value.(type) {
	case float64:
		return true
	}
	return false
}

/*
	Addition usually means between numbers, but can also mean string concat.
	String concat needs one (or both) of the sides to be a string.
*/
func additionTypeCheck(left interface{}, right interface{}) bool {

	if isFloat64(left) && isFloat64(right) {
		return true
	}
	if !isString(left) && !isString(right) {
		return false
	}
	return true
}
