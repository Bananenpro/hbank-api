package middlewares

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/Bananenpro/oidc-client/oidc"

	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/responses"
)

func JWT(oidcClient *oidc.Client, userStore models.UserStore) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			lang := c.Get("lang").(string)
			idToken := ""
			idTokenCookie, err := c.Cookie("ID-Token")
			if err == nil {
				idToken = idTokenCookie.Value
			}
			idTokenSignature := ""
			idTokenSignatureCookie, err := c.Cookie("ID-Token-Signature")
			if err == nil {
				idTokenSignature = idTokenSignatureCookie.Value
			}

			token, err := oidcClient.VerifyIDToken(idToken + "." + idTokenSignature)
			if err != nil {
				if errors.Is(err, oidc.ErrExpiredToken) || idToken == "" || idTokenSignature == "" {
					refreshToken, err := c.Cookie("Refresh-Token")
					if err != nil {
						return c.JSON(http.StatusUnauthorized, responses.New(false, "Missing or expired ID token", lang))
					}
					userID, access, refresh, id, err := oidcClient.RefreshTokens(refreshToken.Value)
					if err != nil {
						return c.JSON(http.StatusUnauthorized, responses.New(false, "Invalid refresh token", lang))
					}
					info, err := oidcClient.FetchUserInfo(userID, access)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
					}

					user, err := userStore.GetById(userID)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
					}
					if user == nil {
						return c.JSON(http.StatusUnauthorized, responses.New(false, "The user does not longer exist", lang))
					}
					user.Name = info.Name
					user.Email = info.Email
					err = userStore.Update(user)
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

					c.Set("userId", userID)
				} else {
					return c.JSON(http.StatusUnauthorized, responses.New(false, "Invalid JWT", lang))
				}
			} else {
				c.Set("userId", token.Subject())
			}

			return next(c)
		}
	}
}
