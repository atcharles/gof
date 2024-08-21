package g2util

import (
	"regexp"
	"strings"

	lzh "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/translations/zh"
	"github.com/pkg/errors"
)

// Valid ...
var Valid = new(Validator).New()

// Validator ...
type Validator struct {
	valid *validator.Validate
	trans ut.Translator
}

type listError []string

func (l listError) Error() string { return strings.Join(l, ",") }

// TranslateZh ...
func (v *Validator) TranslateZh(es error) (err error) {
	if es == nil {
		return
	}
	switch _err := es.(type) {
	case validator.ValidationErrors:
		list := make(listError, 0)
		for _, fieldError := range _err {
			list = append(list, fieldError.Translate(v.trans))
		}
		err = list
	case *validator.InvalidValidationError:
		err = errors.Errorf("Invalid:%s", es.Error())
	default:
		err = es
		return
	}
	return
}

// ValidAndTranslateZh ...
func (v *Validator) ValidAndTranslateZh(s interface{}) (err error) {
	return v.TranslateZh(v.valid.Struct(s))
}

// ValidVarWithTransZh ...
func (v *Validator) ValidVarWithTransZh(val interface{}, tag string) (err error) {
	return v.TranslateZh(v.valid.Var(val, tag))
}

// Valid ...
func (v *Validator) Valid() *validator.Validate { return v.valid }

// Constructor ...
func (v *Validator) Constructor() {
	err := v.initialize()
	if err != nil {
		panic(err)
	}
}

// New ...
func (v *Validator) New() *Validator { v.Constructor(); return v }

// initialize ...
func (v *Validator) initialize() (err error) {
	//中文翻译器
	zh1 := lzh.New()
	trans, _ := ut.New(zh1, zh1).GetTranslator("zh")
	//验证器
	validate := validator.New()
	//验证器注册翻译器
	err = zh.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		return
	}
	v.trans = trans
	v.valid = validate
	customValidations := []func() error{
		v.regTranslationMobile,
		v.regTranslationUsername,
		v.regPassword,
	}
	for _, fn := range customValidations {
		if err = fn(); err != nil {
			return
		}
	}
	return
}

// regTranslationMobile ...
func (v *Validator) regTranslationMobile() (err error) {
	return v.validatorRegValidation("mobile", "{0}手机号码错误", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^1[3456789]\d{9}$`).MatchString(fl.Field().String())
	})
}

// regTranslationUsername ...
func (v *Validator) regTranslationUsername() (err error) {
	return v.validatorRegValidation("username", "{0}(用户账号为数字或字母组合;并且长度为[5-15])",
		func(fl validator.FieldLevel) bool {
			return regexp.MustCompile(`^\w{5,15}$`).MatchString(fl.Field().String())
		},
	)
}

// regPassword ...
func (v *Validator) regPassword() (err error) {
	_fn1 := func(str string) bool {
		if !regexp.MustCompile(`^\S{8,24}$`).MatchString(str) {
			return false
		}
		if !regexp.MustCompile(`^.*[A-Z]+.*$`).MatchString(str) {
			return false
		}
		if !regexp.MustCompile(`^.*[a-z]+.*$`).MatchString(str) {
			return false
		}
		if !regexp.MustCompile(`^.*[0-9]+.*$`).MatchString(str) {
			return false
		}
		return true
	}
	return v.validatorRegValidation("password", "{0}(密码必须同时包含大小写字母和数字;并且长度为[8-24];不能包含空格)",
		func(fl validator.FieldLevel) bool { return _fn1(fl.Field().String()) })
}

func (v *Validator) validatorRegValidation(tagName, text string, fn validator.Func) (err error) {
	if err = v.valid.RegisterValidation(tagName, fn); err != nil {
		return
	}
	return v.valid.RegisterTranslation(tagName, v.trans, func(ut ut.Translator) error {
		return ut.Add(tagName, text, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tagName, fe.Field())
		return t
	})
}
