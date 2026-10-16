package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotStruct       = errors.New("значение не является структурой")
	ErrInvalidRule     = errors.New("некорректное правило валидации")
	ErrUnsupportedType = errors.New("неподдерживаемый тип поля")

	ErrInvalidLength  = errors.New("неверная длина строки")
	ErrRegexpNotMatch = errors.New("строка не соответствует regexp")
	ErrNotInList      = errors.New("значение не входит в список допустимых")
	ErrLessThanMin    = errors.New("значение меньше минимума")
	ErrGreaterThanMax = errors.New("значение больше максимума")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}

	parts := make([]string, 0, len(v))
	for _, item := range v {
		parts = append(parts, item.Field+": "+item.Err.Error())
	}
	return strings.Join(parts, "; ")
}

func Validate(v interface{}) error {
	structValue := reflect.ValueOf(v)
	if structValue.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	var validationErrors ValidationErrors
	structType := structValue.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		fieldErrors, err := validateField(field.Name, structValue.Field(i), validateTag)
		if err != nil {
			return err
		}
		validationErrors = append(validationErrors, fieldErrors...)
	}

	if len(validationErrors) == 0 {
		return nil
	}
	return validationErrors
}

func validateField(fieldName string, fieldValue reflect.Value, validateTag string) (ValidationErrors, error) {
	rules := strings.Split(validateTag, "|")

	if fieldValue.Kind() == reflect.Slice {
		return validateSlice(fieldName, fieldValue, rules)
	}

	return validateStringOrInt(fieldName, fieldValue, rules)
}

func validateSlice(fieldName string, sliceValue reflect.Value, rules []string) (ValidationErrors, error) {
	var validationErrors ValidationErrors

	for i := 0; i < sliceValue.Len(); i++ {
		elemErrors, err := validateStringOrInt(fieldName, sliceValue.Index(i), rules)
		if err != nil {
			return nil, err
		}
		validationErrors = append(validationErrors, elemErrors...)
	}

	return validationErrors, nil
}

func validateStringOrInt(fieldName string, value reflect.Value, rules []string) (ValidationErrors, error) {
	if value.Kind() == reflect.String {
		return validateString(fieldName, value.String(), rules)
	}
	if value.Kind() == reflect.Int {
		return validateInt(fieldName, value.Int(), rules)
	}
	return nil, ErrUnsupportedType
}

func validateString(fieldName, value string, rules []string) (ValidationErrors, error) {
	var validationErrors ValidationErrors

	for _, rule := range rules {
		ruleName, ruleParam, err := parseRule(rule)
		if err != nil {
			return nil, err
		}

		switch ruleName {
		case "len":
			needLen, convErr := strconv.Atoi(ruleParam)
			if convErr != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidRule, rule)
			}
			if len(value) != needLen {
				validationErrors = append(validationErrors, ValidationError{
					Field: fieldName,
					Err:   ErrInvalidLength,
				})
			}

		case "regexp":
			re, compileErr := regexp.Compile(ruleParam)
			if compileErr != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidRule, rule)
			}
			if !re.MatchString(value) {
				validationErrors = append(validationErrors, ValidationError{
					Field: fieldName,
					Err:   ErrRegexpNotMatch,
				})
			}

		case "in":
			if !isStringInAllowedList(value, ruleParam) {
				validationErrors = append(validationErrors, ValidationError{
					Field: fieldName,
					Err:   ErrNotInList,
				})
			}

		default:
			return nil, fmt.Errorf("%w: %s", ErrInvalidRule, rule)
		}
	}

	return validationErrors, nil
}

func validateInt(fieldName string, value int64, rules []string) (ValidationErrors, error) {
	var validationErrors ValidationErrors

	for _, rule := range rules {
		ruleName, ruleParam, err := parseRule(rule)
		if err != nil {
			return nil, err
		}

		switch ruleName {
		case "min":
			minValue, convErr := strconv.Atoi(ruleParam)
			if convErr != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidRule, rule)
			}
			if value < int64(minValue) {
				validationErrors = append(validationErrors, ValidationError{
					Field: fieldName,
					Err:   ErrLessThanMin,
				})
			}

		case "max":
			maxValue, convErr := strconv.Atoi(ruleParam)
			if convErr != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidRule, rule)
			}
			if value > int64(maxValue) {
				validationErrors = append(validationErrors, ValidationError{
					Field: fieldName,
					Err:   ErrGreaterThanMax,
				})
			}

		case "in":
			found, inErr := isIntInAllowedList(value, ruleParam)
			if inErr != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidRule, rule)
			}
			if !found {
				validationErrors = append(validationErrors, ValidationError{
					Field: fieldName,
					Err:   ErrNotInList,
				})
			}

		default:
			return nil, fmt.Errorf("%w: %s", ErrInvalidRule, rule)
		}
	}

	return validationErrors, nil
}

func parseRule(rule string) (ruleName, ruleParam string, err error) {
	parts := strings.SplitN(rule, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("%w: %s", ErrInvalidRule, rule)
	}
	return parts[0], parts[1], nil
}

func isStringInAllowedList(value, allowedList string) bool {
	allowedValues := strings.Split(allowedList, ",")
	for _, allowed := range allowedValues {
		if allowed == value {
			return true
		}
	}
	return false
}

func isIntInAllowedList(value int64, allowedList string) (bool, error) {
	allowedValues := strings.Split(allowedList, ",")
	for _, allowed := range allowedValues {
		num, err := strconv.Atoi(allowed)
		if err != nil {
			return false, err
		}
		if value == int64(num) {
			return true, nil
		}
	}
	return false, nil
}
