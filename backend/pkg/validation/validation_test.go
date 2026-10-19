package validation

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
	"testing"

	apperrors "golang/pkg/errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

type userDTO struct {
	Name   string                `json:"name" validate:"required,min=3,max=50"`
	Age    int                   `json:"age" validate:"required,gte=0,lte=130"`
	Active bool                  `json:"active"`
	Score  float64               `json:"score"`
	Avatar *multipart.FileHeader `json:"avatar"`
}

type validatedDTO struct {
	Email string `json:"email" validate:"required,email"`
	Count int    `json:"count" validate:"required,gte=1"`
}

type queryDTO struct {
	Page  int    `json:"page" validate:"gte=1"`
	Limit int    `json:"limit" validate:"required,gte=1,lte=100"`
	Role  string `json:"role"`
}

func newCtx(app *fiber.App, ctype []byte, body []byte) *fiber.Ctx {
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx.Request().Header.SetContentType(string(ctype))
	ctx.Request().SetBody(body)
	return ctx
}

func newCtxWithQuery(app *fiber.App, query string) *fiber.Ctx {
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx.Request().Header.SetMethod(fiber.MethodGet)
	ctx.Request().SetRequestURI("/?" + query)
	return ctx
}

func TestBindAndValidateJSON(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	body := []byte(`{"name":"Andi","age":28,"active":true,"score":9.5}`)
	c := newCtx(app, []byte(fiber.MIMEApplicationJSON), body)
	defer app.ReleaseCtx(c)

	var dto userDTO
	if err := Default.BindAndValidate(c, &dto); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if dto.Name != "Andi" || dto.Age != 28 || !dto.Active || dto.Score != 9.5 {
		t.Fatalf("unexpected binding result: %+v", dto)
	}
}

func TestBindAndValidateJSONRejectsBadPayload(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	body := []byte(`not-json`)
	c := newCtx(app, []byte(fiber.MIMEApplicationJSON), body)
	defer app.ReleaseCtx(c)

	var dto userDTO
	err := Default.BindAndValidate(c, &dto)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if _, ok := err.(*apperrors.AppError); !ok {
		t.Fatalf("expected *apperrors.AppError, got %T: %v", err, err)
	}
}

func TestBindAndValidateJSONValidationErrors(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	body := []byte(`{"email":"not-an-email","count":0}`)
	c := newCtx(app, []byte(fiber.MIMEApplicationJSON), body)
	defer app.ReleaseCtx(c)

	var dto validatedDTO
	err := Default.BindAndValidate(c, &dto)
	if err == nil {
		t.Fatal("expected validation errors")
	}
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		t.Fatalf("expected validator.ValidationErrors, got %T: %v", err, err)
	}
	if len(ve) < 2 {
		t.Fatalf("expected at least 2 validation errors, got %d", len(ve))
	}
}

func TestBindMultipartFormAllTypes(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("name", "Budi")
	_ = mw.WriteField("age", "31")
	_ = mw.WriteField("active", "true")
	_ = mw.WriteField("score", "7.25")
	mw.Close()

	c := newCtx(app, []byte(mw.FormDataContentType()), buf.Bytes())
	defer app.ReleaseCtx(c)

	var dto userDTO
	if err := Default.BindAndValidate(c, &dto); err != nil {
		t.Fatalf("expected multipart bound OK, got error: %v", err)
	}
	if dto.Name != "Budi" || dto.Age != 31 || !dto.Active || dto.Score != 7.25 {
		t.Fatalf("unexpected DTO from multipart: %+v", dto)
	}
}

func TestBindURLEncodedForm(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	body := []byte("name=Cici&age=25&active=false&score=8.5")
	c := newCtx(app, []byte(fiber.MIMEApplicationForm), body)
	defer app.ReleaseCtx(c)

	var dto userDTO
	if err := Default.BindAndValidate(c, &dto); err != nil {
		t.Fatalf("expected urlencoded OK, got error: %v", err)
	}
	if dto.Name != "Cici" || dto.Age != 25 || dto.Active || dto.Score != 8.5 {
		t.Fatalf("unexpected DTO from urlencoded: %+v", dto)
	}
}

func TestBindMultipartUploads(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("name", "Dedi")
	_ = mw.WriteField("age", "22")

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="avatar"; filename="me.png"`)
	h.Set("Content-Type", "image/png")
	part, _ := mw.CreatePart(h)
	_, _ = part.Write([]byte("fake-png-bytes"))
	mw.Close()

	c := newCtx(app, []byte(mw.FormDataContentType()), buf.Bytes())
	defer app.ReleaseCtx(c)

	var dto userDTO
	if err := Default.BindAndValidate(c, &dto); err != nil {
		t.Fatalf("expected multipart upload OK, got error: %v", err)
	}
	if dto.Avatar == nil {
		t.Fatal("expected Avatar file header to be bound")
	}
	f, err := dto.Avatar.Open()
	if err != nil {
		t.Fatalf("failed to open uploaded file: %v", err)
	}
	defer f.Close()
	data, _ := io.ReadAll(f)
	if string(data) != "fake-png-bytes" {
		t.Fatalf("unexpected uploaded content: %q", string(data))
	}
}

func TestBindQuery(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	c := newCtxWithQuery(app, "page=2&limit=50&role=admin")
	defer app.ReleaseCtx(c)

	var dto queryDTO
	if err := Default.BindQueryAndValidate(c, &dto); err != nil {
		t.Fatalf("expected valid query, got error: %v", err)
	}
	if dto.Page != 2 || dto.Limit != 50 || dto.Role != "admin" {
		t.Fatalf("unexpected DTO from query: %+v", dto)
	}
}

func TestBindQueryRejectsOutOfRange(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	c := newCtxWithQuery(app, "page=0&limit=1000")
	defer app.ReleaseCtx(c)

	var dto queryDTO
	err := Default.BindQueryAndValidate(c, &dto)
	if err == nil {
		t.Fatal("expected validation errors")
	}
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		t.Fatalf("expected validator.ValidationErrors, got %T: %v", err, err)
	}
	if len(ve) < 2 {
		t.Fatalf("expected at least 2 validation errors, got %d", len(ve))
	}
}

func TestBindQueryRejectsMissingRequired(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	c := newCtxWithQuery(app, "page=1")
	defer app.ReleaseCtx(c)

	var dto queryDTO
	err := Default.BindQueryAndValidate(c, &dto)
	if err == nil {
		t.Fatal("expected error for missing required limit")
	}
	if _, ok := err.(validator.ValidationErrors); !ok {
		t.Fatalf("expected validator.ValidationErrors, got %T: %v", err, err)
	}
}

func TestBindQueryRejectsInvalidValue(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	c := newCtxWithQuery(app, "page=abc&limit=10")
	defer app.ReleaseCtx(c)

	var dto queryDTO
	err := Default.BindQuery(c, &dto)
	if err == nil {
		t.Fatal("expected error for non-numeric query value")
	}
	if _, ok := err.(*apperrors.AppError); !ok {
		t.Fatalf("expected *apperrors.AppError, got %T: %v", err, err)
	}
}

func TestBindQueryMultipleValues(t *testing.T) {
	app := fiber.New()
	defer func() { _ = app.Shutdown() }()

	var dto struct {
		Tags []string `json:"tags" validate:"required,min=2"`
	}

	c := newCtxWithQuery(app, "tags=a&tags=b&tags=c")
	defer app.ReleaseCtx(c)

	if err := Default.BindQueryAndValidate(c, &dto); err != nil {
		t.Fatalf("expected multi-value query OK, got error: %v", err)
	}
	if len(dto.Tags) != 3 || dto.Tags[0] != "a" || dto.Tags[2] != "c" {
		t.Fatalf("unexpected tags: %+v", dto.Tags)
	}
}
