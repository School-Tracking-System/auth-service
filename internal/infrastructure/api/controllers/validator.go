package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"

	apiErrors "github.com/fercho/school-tracking/services/auth/internal/infrastructure/api/errors"
)

// Time format regex to match hh:mm (24-hour clock).
var timeFormatRegex = regexp.MustCompile(`^(2[0-3]|[01]?[0-9]):([0-5]?[0-9])$`)

// validate is the validator instance.
var validate = validator.New()

// init registers custom validations when the package loads.
func init() {
	err := RegisterCustomValidations(validate)
	if err != nil {
		return
	}
}

// bindAndValidateBody binds the request body to the request object.
func bindAndValidateBody(ctx context.Context, r *http.Request, request interface{}) *apiErrors.Error {
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return apiErrors.NewBodyReadError(err)
	}

	requestValue := reflect.ValueOf(request).Elem()
	requestType := requestValue.Type()

	if requestType.Kind() == reflect.Slice || requestType.Kind() == reflect.Array {
		if requestValue.Len() == 0 {
			return apiErrors.NewBadRequestError("request body slice cannot be empty")
		}

		if requestType.Elem().Kind() != reflect.Struct {
			// No validation needed for primitive types, return nil on successful decoding
			return nil
		}

		for i := 0; i < requestValue.Len(); i++ {
			element := requestValue.Index(i).Addr().Interface()
			if err := validate.StructCtx(ctx, element); err != nil {
				return apiErrors.NewValidatorError(err)
			}
		}

		return nil
	}

	if err := validate.StructCtx(ctx, request); err != nil {
		return apiErrors.NewValidatorError(err)
	}

	return nil
}

// ValidationFunction is a type for custom validation functions.
type ValidationFunction struct {
	Tag  string
	Func validator.Func
}

// RegisterCustomValidations registers custom validation functions and returns any errors.
func RegisterCustomValidations(validate *validator.Validate) error {
	validations := []ValidationFunction{
		{Tag: "time_format", Func: validateTimeFormat},
	}

	for _, validation := range validations {
		if err := validate.RegisterValidation(validation.Tag, validation.Func); err != nil {
			return err // Return the error if registration fails
		}
	}

	return nil // Return nil if all registrations succeed
}

// Custom validation function for hh:mm format.
func validateTimeFormat(fl validator.FieldLevel) bool {
	timeStr := fl.Field().String()
	return timeFormatRegex.MatchString(timeStr)
}
