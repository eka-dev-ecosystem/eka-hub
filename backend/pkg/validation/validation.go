// Package validation provides content-aware request binding plus schema
// validation for Fiber handlers. It binds both application/json bodies and
// multipart/form-data (including uploads) into the same DTO struct by reusing
// the JSON tags as the canonical field names, then runs the
// go-playground/validator rules declared via the standard `validate` tags.
//
// Supported binding sources:
//   - application/json
//   - application/x-www-form-urlencoded
//   - multipart/form-data (form fields + *multipart.FileHeader uploads)
//   - query string (via BindQuery / BindQueryAndValidate)
//
// Numeric and boolean form/query fields are converted to their destination
// type (string, int, int64, uint, float32, float64, bool, time.Time, slices,
// ...) by the underlying schema decoder, so one DTO handles "all types".
package validation

import (
	"mime/multipart"
	"reflect"
	"strings"
	"sync"

	customErrors "golang/pkg/errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/schema"
)

// Default is a ready-to-use validator that handles the common case.
var Default = New()

// Validator configures how request bodies are bound and validated.
type Validator struct {
	v *validator.Validate
	d *schema.Decoder

	mu sync.Mutex
}

// New builds a Validator backed by a fresh go-playground/validator, with empty
// values left untouched and unknown form keys ignored (JSON already ignores
// unknown keys by default).
func New() *Validator {
	v := validator.New()

	d := schema.NewDecoder()
	d.IgnoreUnknownKeys(true)
	d.SetAliasTag("json")

	return &Validator{v: v, d: d}
}

// Engine exposes the underlying validator.Validate instance for advanced
// customization (custom validation funcs, aliases, translations).
func (v *Validator) Engine() *validator.Validate {
	return v.v
}

// Bind parses the request body into out according to its Content-Type. Only
// JSON, urlencoded and multipart/form-data are accepted; anything else (or an
// unparseable body) returns a 400 bad request error.
func (v *Validator) Bind(c *fiber.Ctx, out any) error {
	ctype := c.Get(fiber.HeaderContentType)
	if end := strings.IndexByte(ctype, ';'); end != -1 {
		ctype = strings.TrimSpace(ctype[:end])
	}

	switch ctype {
	case fiber.MIMEApplicationJSON:
		return v.bindJSON(c, out)
	case fiber.MIMEApplicationForm, fiber.MIMEMultipartForm:
		return v.bindForm(c, out)
	case "":
		if len(c.Body()) == 0 {
			return nil
		}
		return customErrors.BadRequest("Missing Content-Type header")
	default:
		return customErrors.BadRequest("Unsupported Content-Type: " + ctype)
	}
}

// Validate checks out against its `validate` tags. It returns
// validator.ValidationErrors, which the application ErrorHandler turns into a
// structured 400/422 VALIDATION_ERROR response.
func (v *Validator) Validate(out any) error {
	if err := v.v.Struct(out); err != nil {
		return err
	}
	return nil
}

// BindAndValidate is the one-call entry point used by handlers.
func (v *Validator) BindAndValidate(c *fiber.Ctx, out any) error {
	if err := v.Bind(c, out); err != nil {
		return err
	}
	return v.Validate(out)
}

// BindQuery parses the request query string into out using its json tags as
// the lookup names. Values are converted to their destination type (int,
// float, bool, slices, ...) exactly like form binding, and an unparseable
// value returns a 400 bad request error.
func (v *Validator) BindQuery(c *fiber.Ctx, out any) error {
	query := make(map[string][]string)
	c.Context().QueryArgs().VisitAll(func(k, val []byte) {
		query[string(k)] = append(query[string(k)], string(val))
	})

	if err := v.decodeForm(out, query); err != nil {
		return customErrors.BadRequest("Unable to parse query string: " + err.Error())
	}
	return nil
}

// BindQueryAndValidate is the one-call entry point for query strings: it
// binds every present parameter and then runs the `validate` tags on the
// struct (so missing required params and out-of-range values are caught by
// the standard validator).
func (v *Validator) BindQueryAndValidate(c *fiber.Ctx, out any) error {
	if err := v.BindQuery(c, out); err != nil {
		return err
	}
	return v.Validate(out)
}

func (v *Validator) bindJSON(c *fiber.Ctx, out any) error {
	if err := c.BodyParser(out); err != nil {
		return customErrors.BadRequest("Request body must be a valid JSON object")
	}
	return nil
}

func (v *Validator) bindForm(c *fiber.Ctx, out any) error {
	if len(c.Body()) == 0 {
		return nil
	}

	// Flatten the incoming form/multipart values; the schema decoder then
	// maps them onto json-tagged fields and converts values to their
	// destination types (int, float, bool, slices, ...).
	var form map[string][]string

	if formMultipart, err := c.MultipartForm(); err == nil {
		form = formMultipart.Value
	} else {
		form = make(map[string][]string)
		c.Context().PostArgs().VisitAll(func(k, val []byte) {
			form[string(k)] = append(form[string(k)], string(val))
		})
	}

	if err := v.decodeForm(out, form); err != nil {
		return customErrors.BadRequest("Unable to parse form body: " + err.Error())
	}
	if err := v.bindFiles(c, out); err != nil {
		return customErrors.BadRequest(err.Error())
	}
	return nil
}

func (v *Validator) decodeForm(out any, form map[string][]string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.d.Decode(out, form)
}

// bindFiles attaches uploaded files to *multipart.FileHeader and
// []*multipart.FileHeader fields using their json tag as the part name.
func (v *Validator) bindFiles(c *fiber.Ctx, out any) error {
	formMultipart, err := c.MultipartForm()
	if err != nil || formMultipart == nil {
		return nil
	}

	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return nil
	}
	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}

		name := field.Tag.Get("json")
		if end := strings.IndexByte(name, ','); end != -1 {
			name = name[:end]
		}
		if name == "" || name == "-" {
			name = field.Name
		}

		switch fv.Interface().(type) {
		case *multipart.FileHeader:
			files := formMultipart.File[name]
			if len(files) > 0 {
				fv.Set(reflect.ValueOf(files[0]))
			}
		case []*multipart.FileHeader:
			if files := formMultipart.File[name]; len(files) > 0 {
				fv.Set(reflect.ValueOf(files))
			}
		}
	}
	return nil
}
