package swagger

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/chatApp/internal/pkg/config"
)

const swaggerLoginPage = `<title>Login</title><link crossorigin=anonymous href=https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css integrity=sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u rel=stylesheet><style>.back{background:#e2e2e2;width:100%;position:absolute;top:0;bottom:0}.div-center{width:400px;height:400px;background-color:#fff;position:absolute;left:0;right:0;top:0;bottom:0;margin:auto;max-width:100%;max-height:100%;overflow:auto;padding:1em 2em;border-bottom:2px solid #ccc;display:table}div.content{display:table-cell;vertical-align:middle}</style><div class=back><div class=div-center><div class=content><h3>Login</h3><hr><form action=/authenticate method=POST><div class=form-group><label for=swaggerUsername>Username</label><input class=form-control id=swaggerUsername name=swaggerUsername placeholder=Username></div><div class=form-group><label for=swaggerPassword>Password</label><input class=form-control id=swaggerPassword name=swaggerPassword placeholder=Password type=password></div><button class="btn btn-success"type=submit>Login</button></form></div></div></div>`
const swaggerLoginErrorPage = `<title>Login</title><link crossorigin=anonymous href=https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css integrity=sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u rel=stylesheet><style>.back{background:#e2e2e2;width:100%;position:absolute;top:0;bottom:0}.invalid-credentials{color:red}.div-center{width:400px;height:400px;background-color:#fff;position:absolute;left:0;right:0;top:0;bottom:0;margin:auto;max-width:100%;max-height:100%;overflow:auto;padding:1em 2em;border-bottom:2px solid #ccc;display:table}div.content{display:table-cell;vertical-align:middle}</style><div class=back><div class=div-center><div class=content><h3>Login</h3><hr><span class=invalid-credentials>Invalid username or password</span><br><br><form action=/authenticate method=POST><div class=form-group><label for=swaggerUsername>Username</label><input class=form-control id=swaggerUsername name=swaggerUsername placeholder=Username></div><div class=form-group><label for=swaggerPassword>Password</label><input class=form-control id=swaggerPassword name=swaggerPassword placeholder=Password type=password></div><button class="btn btn-success"type=submit>Login</button></form></div></div></div>`

const tokenName = "chatApp-auth-token"

var tokens = make(map[string]string)

func getTokenName() (t string) {
	return tokenName
}

func createToken() (t string) {
	token, _ := uuid.NewV4()
	t = token.String()
	tokens[t] = t
	return t
}

func validateToken(token string) (err error) {
	if _, ok := tokens[token]; !ok {
		return err
	}
	return nil
}

func SetupSwagger(cfg config.ChatApiConfig, e *echo.Echo) {
	SwaggerInfo.Host = cfg.SwaggerHostUrl
	SwaggerInfo.Schemes = strings.Split(cfg.SwaggerHostScheme, ",")
	SwaggerInfo.Title = "Chat APP"
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.POST("/swagger/*", echoSwagger.WrapHandler)

	e.GET(
		"/login", func(ctx echo.Context) error {
			token, err := ctx.Cookie(getTokenName())
			if err != nil {
				return ctx.HTML(http.StatusOK, swaggerLoginPage)
			}
			if err := validateToken(token.Value); err != nil {
				return ctx.HTML(http.StatusOK, swaggerLoginPage)
			}
			return ctx.Redirect(http.StatusPermanentRedirect, "/swagger/index.html")
		},
	)

	e.GET(
		"/", func(ctx echo.Context) error {
			token, err := ctx.Cookie(getTokenName())
			if err != nil {
				return ctx.HTML(http.StatusOK, swaggerLoginPage)
			}
			if err := validateToken(token.Value); err != nil {
				return ctx.HTML(http.StatusOK, swaggerLoginPage)
			}
			return ctx.Redirect(http.StatusPermanentRedirect, "/swagger/index.html")
		},
	)

	e.POST(
		"/authenticate", func(ctx echo.Context) error {
			un := ctx.FormValue("swaggerUsername")
			pass := ctx.FormValue("swaggerPassword")

			if un != cfg.SwaggerUsername || pass != cfg.SwaggerPassword {
				return ctx.HTML(http.StatusOK, swaggerLoginErrorPage)
			}

			t := createToken()
			ctx.SetCookie(
				&http.Cookie{
					Name:     getTokenName(),
					Value:    t,
					Expires:  time.Now().Add(time.Hour * 24),
					Secure:   true,
					HttpOnly: false,
				},
			)
			return ctx.Redirect(http.StatusPermanentRedirect, "/swagger/index.html")
		},
	)
}

// RedirectSwagger redirects the user to the login page if they are not logged in
func RedirectSwagger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if ctx.Request().RequestURI == "/swagger/index.html" {
			token, err := ctx.Cookie(getTokenName())
			if err != nil {
				return ctx.Redirect(http.StatusTemporaryRedirect, "/login")
			}
			err = validateToken(token.Value)
			if err != nil {
				return ctx.Redirect(http.StatusTemporaryRedirect, "/login")
			}
			return next(ctx)
		}
		return next(ctx)
	}
}
