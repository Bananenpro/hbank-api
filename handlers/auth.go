package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/juho05/h-bank/config"
	"github.com/juho05/h-bank/models"
	"github.com/juho05/h-bank/responses"
)

// /api/auth/login (GET)
func (h *Handler) Login(c echo.Context) error {
	h.oidcClient.InitiateAuthFlowWithData(c.Response().Writer, c.Request(), []string{"openid", "profile", "email"}, c.QueryParam("redirect"))
	return nil
}

// /api/auth/logout (POST)
func (h *Handler) Logout(c echo.Context) error {
	sameSite := http.SameSiteStrictMode

	if config.Data.Debug {
		sameSite = http.SameSiteNoneMode
	}

	c.SetCookie(&http.Cookie{
		Name:     "Refresh-Token",
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: sameSite,
		Domain:   config.Data.DomainName,
		Path:     "/",
	})

	c.SetCookie(&http.Cookie{
		Name:     "ID-Token",
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: false,
		SameSite: sameSite,
		Domain:   config.Data.DomainName,
		Path:     "/",
	})

	c.SetCookie(&http.Cookie{
		Name:     "ID-Token-Signature",
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: sameSite,
		Domain:   config.Data.DomainName,
		Path:     "/",
	})
	return nil
}

// /api/auth/login (GET)
func (h *Handler) LoginCallback(c echo.Context) error {
	lang := c.Get("lang").(string)
	userID, access, refresh, id, data, err := h.oidcClient.FinishAuthFlowWithData(c.Response().Writer, c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewUnexpectedError(err, lang))
	}

	info, err := h.oidcClient.FetchUserInfo(userID, access)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	user, err := h.userStore.GetById(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		err = h.userStore.Create(&models.User{
			Base: models.Base{
				Id: userID,
			},
			Name:                    info.Name,
			Email:                   info.Email,
			PubliclyVisible:         true,
			DontSendInvitationEmail: false,
		})
	} else {
		user.Name = info.Name
		user.Email = info.Email
		err = h.userStore.Update(user)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	sameSite := http.SameSiteStrictMode

	if config.Data.Debug {
		sameSite = http.SameSiteNoneMode
	}

	c.SetCookie(&http.Cookie{
		Name:     "Refresh-Token",
		Value:    refresh,
		MaxAge:   int((12 * 7 * 24 * time.Hour).Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: sameSite,
		Domain:   config.Data.DomainName,
		Path:     "/",
	})

	idTokenParts := strings.Split(id, ".")

	c.SetCookie(&http.Cookie{
		Name:     "ID-Token",
		Value:    strings.Join(idTokenParts[:2], "."),
		MaxAge:   int((30 * time.Minute).Seconds()),
		Secure:   true,
		HttpOnly: false,
		SameSite: sameSite,
		Domain:   config.Data.DomainName,
		Path:     "/",
	})

	c.SetCookie(&http.Cookie{
		Name:     "ID-Token-Signature",
		Value:    idTokenParts[2],
		MaxAge:   int((30 * time.Minute).Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: sameSite,
		Domain:   config.Data.DomainName,
		Path:     "/",
	})

	c.Redirect(http.StatusSeeOther, config.Data.BaseURL+"/"+strings.TrimPrefix(data.(string), "/"))
	return nil
}
