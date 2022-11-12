package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	sessionName = "vonage-go-sample"
)

// PROCESS.
// create session manager
// create session
// start verify process
// save requestID in the session
// check
//
// VIEW.
// home
//   if session != nil -> show "VERIFIED USER"
//   else -> "NOT VERIFIED USER"
// send
// verify

func home(c echo.Context) error {
	sess, err := session.Get(sessionName, c)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("fail to call session: %s", err))
	}
	v, exist := sess.Values["verifiedNumber"]
	if !exist {
		return c.String(http.StatusOK, "New Comer")
	}
	if v == "" {
		return c.String(http.StatusOK, "New Comer")
	}
	return c.String(http.StatusOK, fmt.Sprintf("Verified person: %s", v))
}

func send(c echo.Context) error {
	return nil
}

func verify(c echo.Context) error {
	sess, err := session.Get(sessionName, c)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("fail to call session: %s", err))
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60,
		HttpOnly: true,
		Secure:   true,
	}
	sess.Values["verifiedNumber"] = "0120100100"
	sess.Save(c.Request(), c.Response())
	return c.NoContent(http.StatusOK)
}

func main() {
	authKey := os.Getenv("SESSION_AUTH_KEY")

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(authKey))))

	e.GET("/", home)
	e.GET("/send", send)
	e.GET("/verify", verify)

	e.Logger.Fatal(e.Start(":1323"))
}
