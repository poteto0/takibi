package validator_test

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/poteto0/takibi"
	"github.com/poteto0/takibi/interfaces"
	"github.com/poteto0/takibi/validator"
	"github.com/stretchr/testify/assert"
)

type Bindings struct{}

type MyContext = interfaces.IContext[Bindings]

type PostForm struct {
	Body string
}

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestForm_StoresValidatedData(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.Post("/posts",
		validator.Form(func(v url.Values, c MyContext) (PostForm, error) {
			return PostForm{Body: v.Get("body")}, nil
		}),
		func(c MyContext) error {
			called = true
			data, ok := validator.Valid[PostForm](c, "form")
			assert.True(t, ok)
			assert.Equal(t, "hello", data.Body)
			return c.Status(http.StatusCreated).Text("created")
		},
	)

	form := url.Values{"body": {"hello"}}
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestForm_ErrStopHaltsChain(t *testing.T) {
	nextCalled := false

	app := takibi.New(&Bindings{})
	app.Post("/posts",
		validator.Form(func(v url.Values, c MyContext) (PostForm, error) {
			if v.Get("body") == "" {
				c.Status(http.StatusUnprocessableEntity).Text("Invalid!")
				return PostForm{}, validator.ErrStop
			}
			return PostForm{Body: v.Get("body")}, nil
		}),
		func(c MyContext) error {
			nextCalled = true
			return c.Status(http.StatusCreated).Text("created")
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "Invalid!", w.Body.String())
}

func TestUnmarshall_StoresTypedBody(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.Post("/users",
		validator.Unmarshall(func(u User, c MyContext) (User, error) {
			if u.Name == "" {
				c.Status(http.StatusUnprocessableEntity).Text("name required")
				return User{}, validator.ErrStop
			}
			return u, nil
		}),
		func(c MyContext) error {
			called = true
			u, ok := validator.Valid[User](c, validator.TargetJson)
			assert.True(t, ok)
			assert.Equal(t, "alice", u.Name)
			assert.Equal(t, 30, u.Age)
			return c.Status(http.StatusCreated).Json(u)
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"alice","age":30}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestUnmarshall_ErrStopHaltsChain(t *testing.T) {
	nextCalled := false

	app := takibi.New(&Bindings{})
	app.Post("/users",
		validator.Unmarshall(func(u User, c MyContext) (User, error) {
			if u.Name == "" {
				c.Status(http.StatusUnprocessableEntity).Text("name required")
				return User{}, validator.ErrStop
			}
			return u, nil
		}),
		func(c MyContext) error {
			nextCalled = true
			return c.Status(http.StatusCreated).Text("created")
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"age":30}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "name required", w.Body.String())
}

type SignUp struct {
	Name string `form:"name"`
	Age  int    `form:"age"`
}

func TestUnmarshallForm_StoresValidatedData(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.Post("/signup",
		validator.UnmarshallForm(func(in SignUp, c MyContext) (SignUp, error) {
			if in.Name == "" {
				c.Status(http.StatusUnprocessableEntity).Text("name required")
				return SignUp{}, validator.ErrStop
			}
			return in, nil
		}),
		func(c MyContext) error {
			called = true
			in, ok := validator.Valid[SignUp](c, validator.TargetForm)
			assert.True(t, ok)
			assert.Equal(t, "alice", in.Name)
			assert.Equal(t, 30, in.Age)
			return c.Status(http.StatusCreated).Text("ok")
		},
	)

	form := url.Values{"name": {"alice"}, "age": {"30"}}
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestUnmarshallForm_ErrStopHaltsChain(t *testing.T) {
	nextCalled := false

	app := takibi.New(&Bindings{})
	app.Post("/signup",
		validator.UnmarshallForm(func(in SignUp, c MyContext) (SignUp, error) {
			if in.Name == "" {
				c.Status(http.StatusUnprocessableEntity).Text("name required")
				return SignUp{}, validator.ErrStop
			}
			return in, nil
		}),
		func(c MyContext) error {
			nextCalled = true
			return c.Status(http.StatusCreated).Text("ok")
		},
	)

	form := url.Values{"age": {"30"}}
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "name required", w.Body.String())
}

func TestUnmarshallForm_DecodeErrorReachesErrorHandler(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.OnError(func(c MyContext, err error) error {
		called = true
		return c.Status(http.StatusBadRequest).Text("decode failed")
	})
	app.Post("/signup",
		validator.UnmarshallForm(func(in SignUp, c MyContext) (SignUp, error) {
			return in, nil
		}),
		func(c MyContext) error {
			return c.Status(http.StatusCreated).Text("ok")
		},
	)

	// age is not a number -> conversion error bubbles to OnError
	form := url.Values{"name": {"alice"}, "age": {"abc"}}
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJson_StoresValidatedData(t *testing.T) {
	app := takibi.New(&Bindings{})
	app.Post("/json",
		validator.Json(func(m map[string]any, c MyContext) (User, error) {
			name, _ := m["name"].(string)
			return User{Name: name}, nil
		}),
		func(c MyContext) error {
			u, ok := validator.Valid[User](c, validator.TargetJson)
			assert.True(t, ok)
			assert.Equal(t, "bob", u.Name)
			return c.Status(http.StatusCreated).Text(u.Name)
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/json", strings.NewReader(`{"name":"bob"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "bob", w.Body.String())
}

func TestQuery_StoresValidatedData(t *testing.T) {
	app := takibi.New(&Bindings{})
	app.Get("/search",
		validator.Query(func(q map[string]string, c MyContext) (string, error) {
			return q["q"], nil
		}),
		func(c MyContext) error {
			term, ok := validator.Valid[string](c, validator.TargetQuery)
			assert.True(t, ok)
			return c.Text("search: " + term)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/search?q=golang", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "search: golang", w.Body.String())
}

// A validator fn that returns a non-ErrStop error must NOT be swallowed —
// it flows to the user's errorHandler so BadRequest etc. can be controlled there.
func TestValidator_NonStopErrorReachesErrorHandler(t *testing.T) {
	nextCalled := false

	app := takibi.New(&Bindings{})
	app.OnError(func(c MyContext, err error) error {
		return c.Status(http.StatusBadRequest).Text("bad request: " + err.Error())
	})
	app.Post("/posts",
		validator.Form(func(v url.Values, c MyContext) (PostForm, error) {
			return PostForm{}, errors.New("invalid form")
		}),
		func(c MyContext) error {
			nextCalled = true
			return c.Text("unreachable")
		},
	)

	form := url.Values{"body": {"x"}}
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "bad request: invalid form", w.Body.String())
}

// A malformed url-encoded body makes ParseForm fail; the error must reach
// the errorHandler rather than being swallowed.
func TestForm_ParseErrorReachesErrorHandler(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.OnError(func(c MyContext, err error) error {
		called = true
		return c.Status(http.StatusBadRequest).Text("parse failed")
	})
	app.Post("/posts",
		validator.Form(func(v url.Values, c MyContext) (PostForm, error) {
			return PostForm{Body: v.Get("body")}, nil
		}),
		func(c MyContext) error {
			return c.Text("unreachable")
		},
	)

	// stray "%zz" is an invalid percent-encoding -> ParseForm returns an error
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader("body=%zz"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// When the extract step itself fails (malformed JSON body), the error must
// reach the errorHandler rather than being swallowed.
func TestUnmarshall_DecodeErrorReachesErrorHandler(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.OnError(func(c MyContext, err error) error {
		called = true
		return c.Status(http.StatusBadRequest).Text("decode failed")
	})
	app.Post("/users",
		validator.Unmarshall(func(u User, c MyContext) (User, error) {
			return u, nil
		}),
		func(c MyContext) error {
			return c.Status(http.StatusCreated).Text("ok")
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

type Upload struct {
	Title    string
	Filename string
	Content  string
}

// buildMultipart creates a multipart/form-data body with one value field and
// one file field, returning the body and its Content-Type header.
func buildMultipart(t *testing.T, field, value, fileField, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	if err := w.WriteField(field, value); err != nil {
		t.Fatal(err)
	}
	fw, err := w.CreateFormFile(fileField, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return body, w.FormDataContentType()
}

func TestFormFile_StoresValidatedData(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.Post("/upload",
		validator.FormFile(func(form *multipart.Form, c MyContext) (Upload, error) {
			fhs := form.File["avatar"]
			if len(fhs) == 0 {
				c.Status(http.StatusUnprocessableEntity).Text("avatar required")
				return Upload{}, validator.ErrStop
			}
			f, err := fhs[0].Open()
			if err != nil {
				return Upload{}, err
			}
			defer f.Close()
			content, err := io.ReadAll(f)
			if err != nil {
				return Upload{}, err
			}
			return Upload{
				Title:    form.Value["title"][0],
				Filename: fhs[0].Filename,
				Content:  string(content),
			}, nil
		}),
		func(c MyContext) error {
			called = true
			up, ok := validator.Valid[Upload](c, validator.TargetFormFile)
			assert.True(t, ok)
			assert.Equal(t, "hello", up.Title)
			assert.Equal(t, "a.txt", up.Filename)
			assert.Equal(t, "file-body", up.Content)
			return c.Status(http.StatusCreated).Text("created")
		},
	)

	body, contentType := buildMultipart(t, "title", "hello", "avatar", "a.txt", "file-body")
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestFormFile_ErrStopHaltsChain(t *testing.T) {
	nextCalled := false

	app := takibi.New(&Bindings{})
	app.Post("/upload",
		validator.FormFile(func(form *multipart.Form, c MyContext) (Upload, error) {
			if len(form.File["avatar"]) == 0 {
				c.Status(http.StatusUnprocessableEntity).Text("avatar required")
				return Upload{}, validator.ErrStop
			}
			return Upload{}, nil
		}),
		func(c MyContext) error {
			nextCalled = true
			return c.Status(http.StatusCreated).Text("created")
		},
	)

	// multipart body with no file field
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	_ = mw.WriteField("title", "x")
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "avatar required", w.Body.String())
}

func TestFormFile_ParseErrorReachesErrorHandler(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.OnError(func(c MyContext, err error) error {
		called = true
		return c.Status(http.StatusBadRequest).Text("parse failed")
	})
	app.Post("/upload",
		validator.FormFile(func(form *multipart.Form, c MyContext) (Upload, error) {
			return Upload{}, nil
		}),
		func(c MyContext) error {
			return c.Text("unreachable")
		},
	)

	// Declares multipart but body is not a valid multipart payload.
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("not-multipart"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// pngBytes returns a minimal byte slice whose signature makes
// http.DetectContentType report "image/png".
func pngBytes() string {
	return string([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0})
}

func TestFileField_StoresValidatedFile(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.Post("/upload",
		validator.FileField[Bindings](validator.FileConstraint{
			Field:        "avatar",
			Required:     true,
			MaxBytes:     1 << 20,
			AllowedTypes: []string{"image/png"},
		}),
		func(c MyContext) error {
			called = true
			file, ok := validator.Valid[validator.UploadedFile](c, validator.TargetFormFile)
			assert.True(t, ok)
			assert.Equal(t, "avatar", file.Field)
			assert.Equal(t, "a.png", file.Filename)
			assert.Equal(t, "image/png", file.ContentType)
			assert.Equal(t, int64(len(pngBytes())), file.Size)

			f, err := file.Open()
			assert.NoError(t, err)
			defer f.Close()
			content, _ := io.ReadAll(f)
			assert.Equal(t, pngBytes(), string(content))
			return c.Status(http.StatusCreated).Text("created")
		},
	)

	body, contentType := buildMultipart(t, "title", "hello", "avatar", "a.png", pngBytes())
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestFileField_RequiredMissing_ReturnsFileError(t *testing.T) {
	var captured error

	app := takibi.New(&Bindings{})
	app.OnError(func(c MyContext, err error) error {
		captured = err
		return c.Status(http.StatusUnprocessableEntity).Text("err")
	})
	app.Post("/upload",
		validator.FileField[Bindings](validator.FileConstraint{Field: "avatar", Required: true}),
		func(c MyContext) error { return c.Text("unreachable") },
	)

	// multipart body with no file field
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	_ = mw.WriteField("title", "x")
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var fe *validator.FileError
	assert.True(t, errors.As(captured, &fe))
	assert.Equal(t, "avatar", fe.Field)
	assert.Equal(t, validator.FileErrRequired, fe.Reason)
}

func TestFileField_TooLarge_ReturnsFileError(t *testing.T) {
	var captured error

	app := takibi.New(&Bindings{})
	app.OnError(func(c MyContext, err error) error {
		captured = err
		return c.Status(http.StatusRequestEntityTooLarge).Text("err")
	})
	app.Post("/upload",
		validator.FileField[Bindings](validator.FileConstraint{Field: "avatar", MaxBytes: 4}),
		func(c MyContext) error { return c.Text("unreachable") },
	)

	body, contentType := buildMultipart(t, "title", "x", "avatar", "a.png", pngBytes())
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	var fe *validator.FileError
	assert.True(t, errors.As(captured, &fe))
	assert.Equal(t, validator.FileErrTooLarge, fe.Reason)
}

func TestFileField_UnsupportedType_ReturnsFileError(t *testing.T) {
	var captured error

	app := takibi.New(&Bindings{})
	app.OnError(func(c MyContext, err error) error {
		captured = err
		return c.Status(http.StatusUnsupportedMediaType).Text("err")
	})
	app.Post("/upload",
		validator.FileField[Bindings](validator.FileConstraint{
			Field:        "avatar",
			AllowedTypes: []string{"image/png"},
		}),
		func(c MyContext) error { return c.Text("unreachable") },
	)

	// plain text content sniffs to text/plain, not image/png — even though the
	// client declares it as image/png the validator rejects it.
	body, contentType := buildMultipart(t, "title", "x", "avatar", "a.png", "just text content")
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	var fe *validator.FileError
	assert.True(t, errors.As(captured, &fe))
	assert.Equal(t, validator.FileErrUnsupportedType, fe.Reason)
}

func TestFileField_OptionalAbsent_StoresZeroValue(t *testing.T) {
	called := false

	app := takibi.New(&Bindings{})
	app.Post("/upload",
		validator.FileField[Bindings](validator.FileConstraint{Field: "avatar"}),
		func(c MyContext) error {
			called = true
			file, ok := validator.Valid[validator.UploadedFile](c, validator.TargetFormFile)
			assert.True(t, ok)
			assert.Equal(t, validator.UploadedFile{}, file)
			return c.Status(http.StatusOK).Text("ok")
		},
	)

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	_ = mw.WriteField("title", "x")
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFile_CombinesFileAndTextFields(t *testing.T) {
	type Avatar struct {
		Title    string
		Filename string
	}
	called := false

	app := takibi.New(&Bindings{})
	app.Post("/upload",
		validator.File(validator.FileConstraint{Field: "avatar", Required: true},
			func(file validator.UploadedFile, form *multipart.Form, c MyContext) (Avatar, error) {
				return Avatar{Title: form.Value["title"][0], Filename: file.Filename}, nil
			}),
		func(c MyContext) error {
			called = true
			a, ok := validator.Valid[Avatar](c, validator.TargetFormFile)
			assert.True(t, ok)
			assert.Equal(t, "hello", a.Title)
			assert.Equal(t, "a.png", a.Filename)
			return c.Status(http.StatusCreated).Text("created")
		},
	)

	body, contentType := buildMultipart(t, "title", "hello", "avatar", "a.png", pngBytes())
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestValid_ReturnsZeroWhenMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := takibi.NewContext[Bindings](w, req, &Bindings{}, nil)

	data, ok := validator.Valid[PostForm](ctx, "form")
	assert.False(t, ok)
	assert.Equal(t, PostForm{}, data)
}
