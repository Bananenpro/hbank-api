package handlers

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Bananenpro/hbank-api/bindings"
	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/responses"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// /v1/user?except=uuid,uuid,…&page=int&pageSize=int&descending=bool (GET)
func (h *Handler) GetUsers(c echo.Context) error {
	lang := c.Get("lang").(string)
	authUserId := c.Get("userId").(uuid.UUID)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if authUser == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	page := 0
	pageSize := 20

	if c.QueryParam("page") != "" {
		page, err = strconv.Atoi(c.QueryParam("page"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, responses.New(false, "'page' query parameter not a number", lang))
		}
	}

	if c.QueryParam("pageSize") != "" {
		pageSize, err = strconv.Atoi(c.QueryParam("pageSize"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, responses.New(false, "'pageSize' query parameter not a number", lang))
		}
		if pageSize > config.Data.MaxPageSize || pageSize < 1 {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Unsupported page size", lang))
		}
	}

	descending := services.StrToBool(c.QueryParam("descending"))

	ids := []uuid.UUID{}
	for _, idStr := range strings.Split(c.QueryParams().Get("exclude"), ",") {
		if idStr != "" {
			id, err := uuid.Parse(idStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid 'exclude' query parameter", lang))
			}
			ids = append(ids, id)
		}
	}

	users, err := h.userStore.GetAll(ids, c.QueryParam("search"), page, pageSize, descending)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	count, err := h.userStore.Count()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewUsers(users, count))
}

// /v1/user/:id (GET)
func (h *Handler) GetUser(c echo.Context) error {
	lang := c.Get("lang").(string)
	authUserId := c.Get("userId").(uuid.UUID)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if authUser == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	if bytes.Equal(authUserId[:], userId[:]) {
		return c.JSON(http.StatusOK, responses.NewAuthUser(authUser))
	}
	return c.JSON(http.StatusOK, responses.NewUser(user))
}

// /v1/user/delete (POST)
func (h *Handler) DeleteUser(c echo.Context) error {
	lang := c.Get("lang").(string)
	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	var body bindings.DeleteUser
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}

	twoFAToken, err := h.userStore.GetTwoFATokenByCode(user, body.TwoFAToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if twoFAToken == nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}
	if twoFAToken.ExpirationTime < time.Now().Unix() {
		h.userStore.DeleteTwoFAToken(twoFAToken)
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}

	groups, err := h.groupStore.GetAllByUser(user, -1, -1, false)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	ids := make([]uuid.UUID, 0)
	for _, g := range groups {
		isAdmin, err := h.groupStore.IsAdmin(&g, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if isAdmin {
			userCount, err := h.groupStore.GetUserCount(&g)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}

			admins, err := h.groupStore.GetAdmins(nil, "", &g, 0, 2, false)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}

			if userCount > 1 && len(admins) == 1 {
				ids = append(ids, g.Id)
			} else if userCount == 1 {
				h.groupStore.Delete(&g)
			}
		}
	}
	if len(ids) > 0 {
		return c.JSON(http.StatusOK, responses.NewDeleteFailedBecauseOfSoleGroupAdmin(ids, lang))
	}

	h.userStore.DeleteTwoFAToken(twoFAToken)
	h.userStore.Delete(user)

	return c.JSON(http.StatusOK, responses.New(true, "Successfully deleted account", lang))
}

// /v1/user/:id?token=string (DELETE)
func (h *Handler) DeleteUserByDeleteToken(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	token := c.QueryParam("token")

	if subtle.ConstantTimeCompare([]byte(token), []byte(user.DeleteToken)) == 1 {
		h.userStore.Delete(user)
		return c.JSON(http.StatusOK, responses.New(true, "Successfully deleted account", lang))
	}

	return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
}

// /v1/user/picture (POST)
func (h *Handler) SetProfilePicture(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	file, err := c.FormFile("profilePicture")
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing profile picture file", lang))
	}

	if file.Size > config.Data.MaxProfilePictureFileSize {
		return c.JSON(http.StatusBadRequest, responses.New(false, fmt.Sprintf(services.Tr("File too big (max %s)", lang), services.SizeInBytesToStr(config.Data.MaxProfilePictureFileSize)), ""))
	}

	mimeType := file.Header.Get("Content-Type")
	if !services.SupportedPictureMimeType(mimeType) {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Unsupported MIME type", lang))
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	defer src.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	pic, err := services.NewPicture(buf.Bytes(), mimeType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	user.ProfilePictureId = uuid.New()
	err = h.userStore.UpdateProfilePicture(user, &models.ProfilePicture{
		Tiny:   pic.Tiny,
		Small:  pic.Small,
		Medium: pic.Medium,
		Large:  pic.Large,
		Huge:   pic.Huge,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.Id{
		Base: responses.Base{
			Success: true,
			Message: services.Tr("Successfully updated profile picture", lang),
		},
		Id: user.ProfilePictureId.String(),
	})
}

// /v1/user/picture (DELETE)
func (h *Handler) RemoveProfilePicture(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	user.ProfilePictureId = uuid.New()
	err = h.userStore.UpdateProfilePicture(user, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.Id{
		Base: responses.Base{
			Success: true,
			Message: services.Tr("Successfully updated profile picture", lang),
		},
		Id: user.ProfilePictureId.String(),
	})
}

// /v1/user/:id/picture?id=uuid&size=tiny/small/medium/large/huge (GET)
func (h *Handler) GetProfilePicture(c echo.Context) error {
	lang := c.Get("lang").(string)

	authUserId := c.Get("userId").(uuid.UUID)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if authUser == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	switch user.ProfilePicturePrivacy {
	case models.ProfilePictureNobody:
		if user.Id.String() != authUser.Id.String() {
			data, err := os.ReadFile("assets/fallback-profile-picture.svg")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}
			return c.Blob(http.StatusOK, "image/svg", data)
		}
	}

	if c.QueryParam("id") != "" && c.QueryParam("id") != user.ProfilePictureId.String() {
		return c.JSON(http.StatusNotFound, responses.New(false, "Wrong profile picture id", lang))
	}

	size := services.PictureSize(c.QueryParam("size"))
	if c.QueryParam("size") != "" {
		if !size.Validate() {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid 'size' query parameter", lang))
		}
	} else {
		size = services.PictureHuge
	}

	profilePicture, err := h.userStore.GetProfilePicture(user, size)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if len(profilePicture) == 0 {
		data, err := os.ReadFile("assets/fallback-profile-picture.svg")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		return c.Blob(http.StatusOK, "image/svg", data)
	}

	return c.Blob(http.StatusOK, "image/jpeg", profilePicture)
}

// /v1/user (PUT)
func (h *Handler) UpdateUser(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	var body bindings.UpdateUser
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user.DontSendInvitationEmail = body.DontSendInvitationEmail

	body.ProfilePicturePrivacy = strings.ToLower(body.ProfilePicturePrivacy)
	switch body.ProfilePicturePrivacy {
	case models.ProfilePictureEverybody, models.ProfilePictureNobody:
		user.ProfilePicturePrivacy = body.ProfilePicturePrivacy
	default:
		return c.JSON(http.StatusOK, responses.New(false, "Invalid profile picture privacy", lang))
	}

	h.userStore.Update(user)

	return c.JSON(http.StatusOK, responses.NewAuthUser(user))
}

// /v1/user/cash/current (GET)
func (h *Handler) GetCurrentCash(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	entry, err := h.userStore.GetLastCashLogEntry(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if entry == nil {
		entry = &models.CashLogEntry{
			Base: models.Base{
				Created: user.Created,
			},
		}
	}

	return c.JSON(http.StatusOK, responses.NewCashLogEntry(entry))
}

// /v1/user/cash/:id (GET)
func (h *Handler) GetCashLogEntryById(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}

	entry, err := h.userStore.GetCashLogEntryById(user, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if entry == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	return c.JSON(http.StatusOK, responses.NewCashLogEntry(entry))
}

// /v1/user/cash?page=int&pageSize=int&oldestFirst=bool (GET)
func (h *Handler) GetCashLog(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	page := 0
	pageSize := 20

	if c.QueryParam("page") != "" {
		page, err = strconv.Atoi(c.QueryParam("page"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, responses.New(false, "'page' query parameter not a number", lang))
		}
	}

	if c.QueryParam("pageSize") != "" {
		pageSize, err = strconv.Atoi(c.QueryParam("pageSize"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, responses.New(false, "'pageSize' query parameter not a number", lang))
		}
		if pageSize > config.Data.MaxPageSize || pageSize < 1 {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Unsupported page size", lang))
		}
	}

	oldestFirst := services.StrToBool(c.QueryParam("oldestFirst"))

	entries, err := h.userStore.GetCashLog(user, c.QueryParam("search"), page, pageSize, oldestFirst)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	count, err := h.userStore.CashLogEntryCount(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewCashLog(entries, count))
}

// /v1/user/cash (POST)
func (h *Handler) AddCashLogEntry(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	var body bindings.AddCashLogEntry
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	body.Title = strings.TrimSpace(body.Title)
	body.Description = strings.TrimSpace(body.Description)

	if utf8.RuneCountInString(body.Title) > config.Data.MaxNameLength {
		return c.JSON(http.StatusOK, responses.New(false, "Title too long", lang))
	}

	if utf8.RuneCountInString(body.Title) < config.Data.MinNameLength {
		return c.JSON(http.StatusOK, responses.New(false, "Title too short", lang))
	}

	if utf8.RuneCountInString(body.Description) > config.Data.MaxDescriptionLength {
		return c.JSON(http.StatusOK, responses.New(false, "Description too long", lang))
	}

	if utf8.RuneCountInString(body.Description) < config.Data.MinDescriptionLength {
		return c.JSON(http.StatusOK, responses.New(false, "Description too short", lang))
	}

	cashLogEntry := models.CashLogEntry{
		ChangeTitle:       body.Title,
		ChangeDescription: body.Description,
		Ct1:               int(body.Ct1),
		Ct2:               int(body.Ct2),
		Ct5:               int(body.Ct5),
		Ct10:              int(body.Ct10),
		Ct20:              int(body.Ct20),
		Ct50:              int(body.Ct50),
		Eur1:              int(body.Eur1),
		Eur2:              int(body.Eur2),
		Eur5:              int(body.Eur5),
		Eur10:             int(body.Eur10),
		Eur20:             int(body.Eur20),
		Eur50:             int(body.Eur50),
		Eur100:            int(body.Eur100),
		Eur200:            int(body.Eur200),
		Eur500:            int(body.Eur500),
	}

	err = h.userStore.AddCashLogEntry(user, &cashLogEntry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusCreated, responses.New(true, "Successfully added new cash log entry", lang))
}
