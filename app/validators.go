package app

import (
	"log"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func setCustomValidators() {
	cv := map[string]validator.Func{
		"password": passwordValidator,
	}

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		log.Panicln("failed to get validator engine")
	}
	for key, value := range cv {
		v.RegisterValidation(key, value)
	}
}

var basicPasswordRegex = regexp.MustCompile("(?s)(^[a-zA-Z][!-~]{8,64}$)")
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
