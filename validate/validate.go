package validate

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/google/uuid"
)

var validate *validator.Validate

var translator ut.Translator

func init() {

	validate = validator.New()

	translator, _ = ut.New(en.New(), en.New()).GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, translator)
}

func Check(val any) error {
	if err := validate.Struct(val); err != nil {

		verrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		if len(verrors) < 1 {
			return nil
		}

		return errors.New(verrors[0].Translate(translator))
	}

	return nil
}

func GenerateID() string {
	return uuid.NewString()
}

func CheckID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("ID is not in its proper form")
	}
	return nil
}
