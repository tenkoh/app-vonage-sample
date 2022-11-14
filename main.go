package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vonage/vonage-go-sdk"
)

const (
	brandName     = "vonage-go-sample"
	sessionName   = "vonage-go-sample"
	sessionMaxAge = 120
)

var (
	apiKey               string
	apiSecret            string
	errConcurrentRequest = errors.New("there is a concurrent verification request in-progress")
	telValidation        = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	pinValidation        = regexp.MustCompile(`^\d{4}$`)
)

func init() {
	apiKey = os.Getenv("VONAGE_API_KEY")
	if apiKey == "" {
		panic("not found vonage api key in env")
	}
	apiSecret = os.Getenv("VONAGE_API_SECRET")
	if apiSecret == "" {
		panic("not found vonage api secret in env")
	}
}

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

func verify(c echo.Context) error {
	tel := c.FormValue("tel")
	if !isValidTel(tel) {
		return c.HTML(
			http.StatusBadRequest,
			`<html><p>invalid telephone number</p><a href="/">go to home</a></html>`,
		)
	}

	// start verify process using vonage api
	reqID, err := requestVerify(c)
	if err != nil && !errors.Is(err, errConcurrentRequest) {
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
	auth := vonage.CreateAuthFromKeySecret(apiKey, apiSecret)
	verifyClient := vonage.NewVerifyClient(auth)

	resp, respErr, err := verifyClient.Request(c.FormValue("tel"), brandName, vonage.VerifyOpts{
		// https://developer.vonage.com/api/verify
		PinExpiry:  sessionMaxAge,
		Lg:         "ja-jp",
		WorkflowID: 6, //only sms (one time)
	})
	if err != nil {
		return "", err
	}
	if t := respErr.ErrorText; t != "" {
		if id := respErr.RequestId; id != "" {
			return id, errConcurrentRequest
		}
		return "", fmt.Errorf("fail to start verification, status=%s, detail=%s", respErr.Status, t)
	}
	return resp.RequestId, nil
}

func requestCheck(c echo.Context) (verified bool, _ error) {
	auth := vonage.CreateAuthFromKeySecret(apiKey, apiSecret)
	verifyClient := vonage.NewVerifyClient(auth)

	sess, err := session.Get(sessionName, c)
	if err != nil {
		return false, fmt.Errorf("fail to call session: %w", err)
	}
	r := sess.Values["requestID"]
	reqID, ok := r.(string)
	if !ok {
		return false, errors.New("invalid requestID in this session")
	}
	if reqID == "" {
		return false, errors.New("invalid requestID in this session")
	}
	pin := c.FormValue("pin")
	if !isValidPin(pin) {
		return false, errors.New("invalid pin code")
	}
	_, respErr, err := verifyClient.Check(reqID, pin)
	if err != nil {
		return false, err
	}
	if t := respErr.ErrorText; t != "" {
		if id := respErr.RequestId; id != "" {
			return false, errConcurrentRequest
		}
		return false, fmt.Errorf("fail to start verification, status=%s, detail=%s", respErr.Status, t)
	}
	return true, nil
}

func isValidTel(tel string) bool {
	return telValidation.MatchString(tel)
}

func isValidPin(pin string) bool {
	return pinValidation.MatchString(pin)
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
