package app

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func setCustomValidators() {
	cv := map[string]validator.Func{
		"username": usernameValidator,
		"password": passwordValidator,
	}

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		panic("failed to get validator engine")
	}
	for key, value := range cv {
		v.RegisterValidation(key, value)
	}
}

var usernameValidatorRegex = regexp.MustCompile("(?s)(^([a-z]|[0-9]|[_-]){4,128}$)")

func usernameValidator(fl validator.FieldLevel) bool {
	return usernameValidatorRegex.MatchString(fl.Field().String())
}

var basicPasswordRegex = regexp.MustCompile("(?s)(^[a-zA-Z][!-~]{7,63}$)")
var isUpperExistRegex = regexp.MustCompile("[A-Z]")

func passwordValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	regexPatterns := []*regexp.Regexp{basicPasswordRegex, isUpperExistRegex}
	for _, rp := range regexPatterns {
		if !rp.MatchString(value) {
			return false
		}
	}
	return true
}
