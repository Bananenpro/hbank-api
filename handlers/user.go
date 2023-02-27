package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v4"

	"github.com/juho05/hbank-api/bindings"
	"github.com/juho05/hbank-api/config"
	"github.com/juho05/hbank-api/models"
	"github.com/juho05/hbank-api/responses"
	"github.com/juho05/hbank-api/services"
)

// /api/user?except=uuid,uuid,…&page=int&pageSize=int&descending=bool (GET)
func (h *Handler) GetUsers(c echo.Context) error {
	lang := c.Get("lang").(string)
	authUserId := c.Get("userId").(string)
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

	ids := []string{}
	for _, id := range strings.Split(c.QueryParams().Get("exclude"), ",") {
		if id != "" {
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

// /api/user/:id (GET)
func (h *Handler) GetUser(c echo.Context) error {
	lang := c.Get("lang").(string)
	authUserId := c.Get("userId").(string)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if authUser == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	userId := c.Param("id")
	if userId == "" {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Missing id parameter", lang))
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	if authUserId == userId {
		return c.JSON(http.StatusOK, responses.NewAuthUser(authUser))
	}
	return c.JSON(http.StatusOK, responses.NewUser(user))
}

// /api/user/delete (POST)
func (h *Handler) DeleteUser(c echo.Context) error {
	// TODO
	return echo.ErrNotFound
}

// /api/user (PUT)
func (h *Handler) UpdateUser(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(string)
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
	user.PubliclyVisible = body.PubliclyVisible
	h.userStore.Update(user)

	return c.JSON(http.StatusOK, responses.NewAuthUser(user))
}

// /api/user/cash/current (GET)
func (h *Handler) GetCurrentCash(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(string)
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

// /api/user/cash/:id (GET)
func (h *Handler) GetCashLogEntryById(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(string)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Missing id parameter", lang))
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

// /api/user/cash?page=int&pageSize=int&oldestFirst=bool (GET)
func (h *Handler) GetCashLog(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(string)
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

// /api/user/cash (POST)
func (h *Handler) AddCashLogEntry(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(string)
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
