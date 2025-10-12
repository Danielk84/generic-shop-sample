package app

import (
	"regexp"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func setCustomValidators() {
	cv := map[string]validator.Func{
		"username":          usernameValidator,
		"password":          passwordValidator,
		"date":              dateValidator,
		"iran_phone_number": iranPhoneNumberValidator,
	}

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		panic("failed to get validator engine")
	}
	for key, value := range cv {
		v.RegisterValidation(key, value)
	}
}

var usernameValidatorRegexp = regexp.MustCompile(`(^([a-z]|[0-9]|[_-]){4,128}$)`)

func usernameValidator(fl validator.FieldLevel) bool {
	return usernameValidatorRegexp.MatchString(fl.Field().String())
}

var basicPasswordRegexp = regexp.MustCompile(`(^[a-zA-Z][!-~]{7,63}$)`)
var isUpperExistRegexp = regexp.MustCompile(`[A-Z]`)

func passwordValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	regexPatterns := []*regexp.Regexp{basicPasswordRegexp, isUpperExistRegexp}
	for _, rp := range regexPatterns {
		if !rp.MatchString(value) {
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
