package validator

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"slices"

	"github.com/poteto0/takibi/constants"
	"github.com/poteto0/takibi/interfaces"
)

// ErrStop aliases constants.ErrStop for convenience — return it from a
// validator fn to halt the handler chain without triggering the error handler.
// The response must be written by the returning handler before returning ErrStop.
var ErrStop = constants.ErrStop

// Target keys used by the built-in validator factories.
const (
	TargetForm     = "form"
	TargetFormFile = "formFile"
	TargetJson     = "json"
	TargetQuery    = "query"
)

// formFileMaxMemory is the in-memory threshold for ParseMultipartForm. Setting
// it to the default body limit keeps the whole form in memory instead of
// spilling to temp files, so FormFile behaves the same on wasm (no filesystem).
const formFileMaxMemory = constants.DefaultMaxBodyBytes

func newValidator[Bindings any, In any, T any](
	key string,
	extract func(interfaces.IContext[Bindings]) (In, error),
	fn func(In, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return func(c interfaces.IContext[Bindings]) error {
		input, err := extract(c)
		if err != nil {
			return err
		}
		result, err := fn(input, c)
		if err != nil {
			return err
		}
		c.SetValidated(key, result)
		return nil
	}
}

// Form returns a HandlerFunc that parses the request's form body and passes
// the values to fn. The returned value is stored under TargetForm ("form").
func Form[Bindings any, T any](
	fn func(url.Values, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetForm, func(c interfaces.IContext[Bindings]) (url.Values, error) {
		if err := c.Req().Raw().ParseForm(); err != nil {
			return nil, err
		}
		return c.Req().Raw().Form, nil
	}, fn)
}

// FormFile returns a HandlerFunc that parses a multipart/form-data request body
// and passes the parsed *multipart.Form (both Value and File maps) to fn. The
// returned value is stored under TargetFormFile ("formFile").
//
// The form is parsed entirely in memory (see formFileMaxMemory) so behaviour is
// identical on native and wasm builds.
func FormFile[Bindings any, T any](
	fn func(*multipart.Form, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetFormFile, func(c interfaces.IContext[Bindings]) (*multipart.Form, error) {
		raw := c.Req().Raw()
		if err := raw.ParseMultipartForm(formFileMaxMemory); err != nil {
			return nil, err
		}
		return raw.MultipartForm, nil
	}, fn)
}

// Reasons reported by FileError.
const (
	FileErrRequired        = "required"
	FileErrTooLarge        = "too_large"
	FileErrUnsupportedType = "unsupported_type"
)

// sniffLen is the number of leading bytes http.DetectContentType inspects.
const sniffLen = 512

// FileConstraint declares the validation rules applied to a single uploaded
// file field by File and FileField.
type FileConstraint struct {
	Field        string   // multipart field name (required)
	Required     bool     // reject when the field carries no file part
	MaxBytes     int64    // reject files larger than this; 0 disables the check
	AllowedTypes []string // permitted (sniffed) Content-Types; empty allows any
}

// UploadedFile is a validated file part. The handler opens it via Open and is
// responsible for closing the returned multipart.File.
type UploadedFile struct {
	Field       string
	Filename    string
	ContentType string // content type detected by sniffing the file's bytes
	Size        int64
	Header      *multipart.FileHeader
}

// Open opens the underlying multipart file for reading. The caller must Close
// the returned file. It returns an error when the file is absent (zero value).
func (u UploadedFile) Open() (multipart.File, error) {
	if u.Header == nil {
		return nil, fmt.Errorf("no file to open for field %q", u.Field)
	}
	return u.Header.Open()
}

// FileError is returned by File/FileField when an uploaded file violates a
// FileConstraint. It flows to app.OnError so the application can render its own
// error response. Reason is one of the FileErr* constants.
type FileError struct {
	Field  string
	Reason string
}

func (e *FileError) Error() string {
	return fmt.Sprintf("file %q: %s", e.Field, e.Reason)
}

// fileInput carries the validated file and the parsed form to the File fn.
type fileInput struct {
	file UploadedFile
	form *multipart.Form
}

// FileErrorHandler maps a constraint violation to a response. It receives the
// *FileError describing the violation and the request context, and returns the
// error that drives the handler chain: return ErrStop after writing a response,
// or return any error to flow to app.OnError.
type FileErrorHandler[Bindings any] func(*FileError, interfaces.IContext[Bindings]) error

// fileOptions collects the configurable behaviour of File/FileField.
type fileOptions[Bindings any] struct {
	onError FileErrorHandler[Bindings]
}

// FileOption customizes File/FileField. Build one with WithFileError.
type FileOption[Bindings any] func(*fileOptions[Bindings])

// WithFileError overrides the default constraint-violation handling. Without it
// File/FileField return the raw *FileError to app.OnError; with it the handler
// decides the response (e.g. write a body and return ErrStop, or transform the
// error). The handler only runs for *FileError violations — parse and I/O
// errors still flow straight to app.OnError.
func WithFileError[Bindings any](h FileErrorHandler[Bindings]) FileOption[Bindings] {
	return func(o *fileOptions[Bindings]) {
		o.onError = h
	}
}

// File parses a multipart/form-data body, validates the file field named by c
// against its constraints, then passes the validated UploadedFile and the full
// *multipart.Form (for the text Value fields) to fn. The returned value is
// stored under TargetFormFile ("formFile").
//
// A constraint violation returns a *FileError (flowing to app.OnError) rather
// than writing a response, so the application controls the error format.
func File[Bindings any, T any](
	c FileConstraint,
	fn func(UploadedFile, *multipart.Form, interfaces.IContext[Bindings]) (T, error),
	opts ...FileOption[Bindings],
) interfaces.HandlerFunc[Bindings] {
	var o fileOptions[Bindings]
	for _, opt := range opts {
		opt(&o)
	}
	return newValidator(TargetFormFile, func(ctx interfaces.IContext[Bindings]) (fileInput, error) {
		raw := ctx.Req().Raw()
		if err := raw.ParseMultipartForm(formFileMaxMemory); err != nil {
			return fileInput{}, err
		}
		form := raw.MultipartForm
		file, err := validateFile(form, c)
		if err != nil {
			if fe, ok := err.(*FileError); ok && o.onError != nil {
				return fileInput{}, o.onError(fe, ctx)
			}
			return fileInput{}, err
		}
		return fileInput{file: file, form: form}, nil
	}, func(in fileInput, ctx interfaces.IContext[Bindings]) (T, error) {
		return fn(in.file, in.form, ctx)
	})
}

// FileField is the no-fn shortcut over File: it validates the named file field
// and stores the resulting UploadedFile under TargetFormFile ("formFile"),
// ready to retrieve with Valid[UploadedFile]. Use File when you also need the
// form's text fields.
func FileField[Bindings any](c FileConstraint, opts ...FileOption[Bindings]) interfaces.HandlerFunc[Bindings] {
	return File(c, func(file UploadedFile, _ *multipart.Form, _ interfaces.IContext[Bindings]) (UploadedFile, error) {
		return file, nil
	}, opts...)
}

// validateFile checks the file part named by c against its constraints. When
// the field is optional and absent it returns the zero UploadedFile with no
// error; a present file is validated for size and (sniffed) content type.
func validateFile(form *multipart.Form, c FileConstraint) (UploadedFile, error) {
	var headers []*multipart.FileHeader
	if form != nil {
		headers = form.File[c.Field]
	}
	if len(headers) == 0 {
		if c.Required {
			return UploadedFile{}, &FileError{Field: c.Field, Reason: FileErrRequired}
		}
		return UploadedFile{}, nil
	}

	fh := headers[0]
	if c.MaxBytes > 0 && fh.Size > c.MaxBytes {
		return UploadedFile{}, &FileError{Field: c.Field, Reason: FileErrTooLarge}
	}

	contentType, err := sniffContentType(fh)
	if err != nil {
		return UploadedFile{}, err
	}
	if len(c.AllowedTypes) > 0 && !slices.Contains(c.AllowedTypes, contentType) {
		return UploadedFile{}, &FileError{Field: c.Field, Reason: FileErrUnsupportedType}
	}

	return UploadedFile{
		Field:       c.Field,
		Filename:    fh.Filename,
		ContentType: contentType,
		Size:        fh.Size,
		Header:      fh,
	}, nil
}

// sniffContentType detects the content type from the file's leading bytes
// instead of trusting the client-declared Content-Type header. Any media-type
// parameters (e.g. "; charset=utf-8") are stripped so the value compares
// cleanly against FileConstraint.AllowedTypes.
func sniffContentType(fh *multipart.FileHeader) (string, error) {
	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, sniffLen)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}
	detected := http.DetectContentType(buf[:n])
	if mediaType, _, err := mime.ParseMediaType(detected); err == nil {
		return mediaType, nil
	}
	return detected, nil
}

// Json returns a HandlerFunc that reads the JSON request body and passes
// the parsed map to fn. The returned value is stored under TargetJson ("json").
func Json[Bindings any, T any](
	fn func(map[string]any, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetJson, func(c interfaces.IContext[Bindings]) (map[string]any, error) {
		return c.Req().Json()
	}, fn)
}

// Query returns a HandlerFunc that passes the request's query parameters to
// fn. The returned value is stored under TargetQuery ("query").
func Query[Bindings any, T any](
	fn func(map[string]string, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetQuery, func(c interfaces.IContext[Bindings]) (map[string]string, error) {
		return c.Req().Query(), nil
	}, fn)
}

// Unmarshall returns a HandlerFunc that decodes the JSON request body into a
// value of type T and passes it to fn. This is the typical typed-body pattern:
// fn receives a fully populated T rather than a map. The returned value is
// stored under TargetJson ("json").
func Unmarshall[Bindings any, T any](
	fn func(T, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetJson, func(c interfaces.IContext[Bindings]) (T, error) {
		var dest T
		if err := c.Req().Unmarshall(&dest); err != nil {
			return dest, err
		}
		return dest, nil
	}, fn)
}

// UnmarshallForm returns a HandlerFunc that binds the form request body
// (urlencoded or multipart Value fields) into a value of type T via `form`
// struct tags and passes it to fn. This is the typed-form counterpart of
// Unmarshall: fn receives a fully populated T rather than url.Values. The
// returned value is stored under TargetForm ("form").
func UnmarshallForm[Bindings any, T any](
	fn func(T, interfaces.IContext[Bindings]) (T, error),
) interfaces.HandlerFunc[Bindings] {
	return newValidator(TargetForm, func(c interfaces.IContext[Bindings]) (T, error) {
		var dest T
		if err := c.Req().UnmarshallForm(&dest); err != nil {
			return dest, err
		}
		return dest, nil
	}, fn)
}

// Valid retrieves the validated value stored under target and type-asserts it
// to T. Returns the zero value and false when the key is absent or the type
// does not match.
func Valid[T any](c interface {
	Validated(string) (any, bool)
}, target string) (T, bool) {
	v, ok := c.Validated(target)
	if !ok {
		var zero T
		return zero, false
	}
	t, ok := v.(T)
	return t, ok
}
