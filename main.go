package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	sessionName   = "vonage-go-sample"
	sessionMaxAge = 120
)

// implatement echo.Template interface
type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data any, e echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var t = &Template{
	templates: template.Must(template.ParseGlob("views/*.tmpl")),
}

func home(c echo.Context) error {
	sess, err := session.Get(sessionName, c)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("fail to call session: %s", err))
	}
	v, exist := sess.Values["verified"]
	if !exist {
		v = false
	}
	return c.Render(http.StatusOK, "home", v)
}

// verify handler starts verify process using vonage api.
// tel number must be passed via form.
func verify(c echo.Context) error {
	tel := c.FormValue("tel")
	log.Printf("requested telephone number: %s\n", tel)
	// validation
	if !isValidNumber(tel) {
		return c.String(http.StatusBadRequest, "invalid telephone number")
	}
	// start verify process using vonage api
	reqID, err := requestVerify(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// save requestID into session
	sess, err := session.Get(sessionName, c)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("fail to call session: %s", err))
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   true,
	}
	sess.Values["requestID"] = reqID
	sess.Save(c.Request(), c.Response())

	// show form to send pin code
	return c.Render(http.StatusOK, "pinform", nil)
}

func check(c echo.Context) error {
	pin := c.FormValue("pin")
	if !isValidPin(pin) {
		return c.String(http.StatusBadRequest, pin)
	}
	verified, err := requestCheck(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if !verified {
		return c.HTML(
			http.StatusForbidden,
			`<html><p>verification failed</p><a href="/">go to home</a></html>`,
		)
	}

	sess, err := session.Get(sessionName, c)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("fail to call session: %s", err))
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   true,
	}
	sess.Values["verified"] = verified
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusSeeOther, "/")
}

func requestVerify(c echo.Context) (reqID string, _ error) {
	return "foobar", nil
}

func requestCheck(c echo.Context) (verified bool, _ error) {
	return false, nil
}

func isValidNumber(tel string) bool {
	if tel == "" {
		return false
	}
	return true
}

func isValidPin(pin string) bool {
	return true
}

func cleanTel(tel string) string {
	return tel
}

func main() {
	authKey := os.Getenv("SESSION_AUTH_KEY")

	e := echo.New()
	e.Renderer = t
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(authKey))))

	e.GET("/", home)
	e.POST("/verify", verify)
	e.POST("/check", check)

	e.Logger.Fatal(e.Start(":1323"))
}
