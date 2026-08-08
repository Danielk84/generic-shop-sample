package internal

import (
	"regexp"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func SetCustomValidators() {
	cv := map[string]validator.Func{
		"username":          usernameValidator,
		"password":          passwordValidator,
		"date":              dateValidator,
		"iran_phone_number": iranPhoneNumberValidator,
	}

	v := GetValidator()
	for key, value := range cv {
		if err := v.RegisterValidation(key, value); err != nil {
			panic(err)
		}
	}
}

func GetValidator() *validator.Validate {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		panic("failed to get validator engine")
	}
	return v
}

var usernameValidatorRegexp = regexp.MustCompile(`(^[-a-z0-9_]{4,128}$)`)

func usernameValidator(fl validator.FieldLevel) bool {
	return usernameValidatorRegexp.MatchString(fl.Field().String())
}

var basicPasswordRegexp = regexp.MustCompile(`(^[-a-zA-Z0-9~!@#$%^&*?_+]{8,64}$)`)
var isLowerExistRegexp = regexp.MustCompile(`[a-z]`)
var isUpperExistRegexp = regexp.MustCompile(`[A-Z]`)
var isNumberExistRegexp = regexp.MustCompile(`[0-9]`)
var isSymbolExistRegexp = regexp.MustCompile(`[-~!@#$%^&*?_+]`)

func passwordValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	regexPatterns := []*regexp.Regexp{
		basicPasswordRegexp,
		isLowerExistRegexp,
		isUpperExistRegexp,
		isNumberExistRegexp,
		isSymbolExistRegexp}
	for _, r := range regexPatterns {
		if !r.MatchString(value) {
			return false
		}
	}
	return true
}

func dateValidator(fl validator.FieldLevel) bool {
	if _, err := time.Parse(time.DateOnly, fl.Field().String()); err != nil {
		return false
	}
	return true
}

var basicIranPhoneNumberRegexp = regexp.MustCompile(`(^(0098|98|\+98|0)9[0-9]{9}$)`)

func iranPhoneNumberValidator(fl validator.FieldLevel) bool {
	return basicIranPhoneNumberRegexp.MatchString(fl.Field().String())
}
