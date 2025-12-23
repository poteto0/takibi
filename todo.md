## Subset

- [ ] Middleware

  - [x] create trie route for Middleware
  - [x] support star(`*`) pattern

    ```go
    app.Use("*", someMiddleware) // use middleware on all routes
    app.Use("/users/*", authMiddleware) // use middleware on users/*
    app.Use("/users/queued", queueMiddleware) // use middleware on users/queued/*
    ```

    search middleware algorithms;
    if request to /users/queued/batch-update

    1. append someMiddleware
    2. trie search start
       a. go to next node
       b. append middleware (if exits)
       c. loop to end
    3. return all middlewares

    middleware type;
    path: interfaces/

    ```go
    func (ctx interfaces.IContext[Bindings any], next interfaces.HandlerFunc) interfaces.HandlerFunc
    ```

    - [x] sample middlewares

      path: middlewares/

      - [x] timeout
      - [x] cors: use golang std cors on middleware

      ```go
      func SampleMiddleware(SampleMiddlewareConfig SampleMiddlewareConfig) interfaces.Middleware {
        func (ctx interfaces.IContext[Bindings], next interfaces.HandlerFunc) interfaces.HandlerFunc {
          return func(ctx interfaces.IContext) error {
            // do something
            return next(c)
          }
        }
      }
      ```

      - [x] factory pattern

      factory pattern temporary use Bindings;

    bindings defines on TakibiApp

    secured middleware

    but not sure for bindings is equal to App's binding: impl if you can

    ```go
    func CreateMiddleware[Bindings any](interfaces.IContext[Bindings], next interfaces.HandlerFunc) interfaces.HandlerFunc {
      return func(ctx interfaces.IContext) error {
        return next(c)
      }
    }
    ```

    ```go
    type Bindings struct {
      Foo: string
    }

    middleware := factory.CreateMiddleware(
      func (ctx interfaces.IContext[Bindings], next interfaces.HandlerFunc) interfaces.HandlerFunc {
        return func(ctx interfaces.IContext) error {
          ctx.Env().Foo = "bar"
          return next(c)
        }
      }
    )

    // in your handler
    func handler(ctx interfaces.IContext[Bindings]) error {
      foo := ctx.Env().Foo // bar
      return nil
    }
    ```

- [x] ErrorHandler

  add user custom errorHandler:
  mapping error -> Response

  ```go
  func errorHandler(ctx interfaces.IContext[Bindings], err error) error {
    if errors.Is(err, YourError) {
      return ctx.Status(400).Json(
        map[string]string{
          "message": "badRequest"
        },
      )
    }

    return ctx.Status(500).Json(
      map[string]string{
        "message": "internalServerError"
      },
    )
  }

  func main() {
    ...
    app.OnError(errorHandler)
    ...
  }

  // in your handler
  func handler(ctx interfaces.IContext[Bindings]) error {
    return YourError // => 400
  }
  ```

## Response

- [x] ctx.Status
- [x] ctx.Text
- [x] ctx.Json
- [x] ctx.Redirect
- [ ] ctx.Html: pending
- [x] Render

  check: Content-Type: text/html

  path: fixtures/hello.templ/hello_templ.go

  ```templ
  package component

  templ Hello(name string) {
    <div>
      hello { name }
    </div>
  }
  ```

  ```go
  type TemplArgs struct {
    name string,
  }

  helper.Render(ctx.Status(200), templ.Component, TemplArgs{
    name: "takibi",
  }) // => <div>hello takibi</div>
  ```

  library implement

  ```go
  package helper

  func Render[TemplateArg any](
    ctx interfaces.IContext,
    componentFunc func(args TemplateArg) templ.Component,
    args: TemplateArg,
  ) error
  ```
