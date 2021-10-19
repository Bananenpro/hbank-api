package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Bananenpro/hbank-api/bindings"
	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/responses"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"github.com/labstack/echo/v4"
)

// /v1/group?page=int&pageSize=int&descending=bool (GET)
func (h *Handler) GetGroups(c echo.Context) error {
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
	}

	descending := services.StrToBool(c.QueryParam("descending"))

	groups, err := h.groupStore.GetAllByUser(user, page, pageSize, descending)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewGroups(groups))
}

// /v1/group/:id (GET)
func (h *Handler) GetGroupById(c echo.Context) error {
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

	group, err := h.groupStore.GetById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	isMember, err := h.groupStore.IsMember(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	isAdmin, err := h.groupStore.IsAdmin(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if isMember || isAdmin {
		return c.JSON(http.StatusOK, responses.NewGroup(group, isMember, isAdmin))
	} else {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not a member/admin of the group", lang))
	}
}

// /v1/group (POST)
func (h *Handler) CreateGroup(c echo.Context) error {
	lang := c.Get("lang").(string)
	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	var body bindings.CreateGroup
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Description = strings.TrimSpace(body.Description)

	if len(body.Name) > config.Data.MaxNameLength {
		return c.JSON(http.StatusOK, responses.New(false, "Name too long", lang))
	}

	if utf8.RuneCountInString(body.Name) < config.Data.MinNameLength {
		return c.JSON(http.StatusOK, responses.New(false, "Name too short", lang))
	}

	if len(body.Description) > config.Data.MaxDescriptionLength {
		return c.JSON(http.StatusOK, responses.New(false, "Description too long", lang))
	}

	if utf8.RuneCountInString(body.Description) < config.Data.MinDescriptionLength {
		return c.JSON(http.StatusOK, responses.New(false, "Description too short", lang))
	}

	group := &models.Group{
		Name:           body.Name,
		Description:    body.Description,
		GroupPictureId: uuid.New(),
	}

	err = h.groupStore.Create(group)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	err = h.groupStore.AddAdmin(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if !body.OnlyAdmin {
		err = h.groupStore.AddMember(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
	}

	return c.JSON(http.StatusCreated, responses.CreateGroupSuccess{
		Base: responses.Base{
			Success: true,
			Message: services.Tr("Successfully created new group", lang),
		},
		Id: group.Id.String(),
	})
}

// /v1/group/:id/member (GET)
func (h *Handler) GetGroupMembers(c echo.Context) error {
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
	}

	descending := services.StrToBool(c.QueryParam("descending"))

	group, err := h.groupStore.GetById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	isInGroup, err := h.groupStore.IsInGroup(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if !isInGroup {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not a member/admin of the group", lang))
	}

	members, err := h.groupStore.GetMembers(group, page, pageSize, descending)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewUsers(members))
}

// /v1/group/:id/admin (GET)
func (h *Handler) GetGroupAdmins(c echo.Context) error {
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
	}

	descending := services.StrToBool(c.QueryParam("descending"))

	group, err := h.groupStore.GetById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	isInGroup, err := h.groupStore.IsInGroup(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if !isInGroup {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not a member/admin of the group", lang))
	}

	members, err := h.groupStore.GetAdmins(group, page, pageSize, descending)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewUsers(members))
}

// /v1/group/:id/picture?id=uuid (GET)
func (h *Handler) GetGroupPicture(c echo.Context) error {
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

	group, err := h.groupStore.GetById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	isInGroup, err := h.groupStore.IsInGroup(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if !isInGroup {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not a member/admin of the group", lang))
	}

	if c.QueryParam("id") != "" && c.QueryParam("id") != user.ProfilePictureId.String() {
		return c.JSON(http.StatusNotFound, responses.New(false, "Wrong group picture id", lang))
	}

	size := config.Data.ProfilePictureSize
	if c.QueryParam("size") != "" {
		size, err = strconv.Atoi(c.QueryParam("size"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, responses.New(false, "The 'size' query parameter is not a number", lang))
		}

		if size > config.Data.ProfilePictureSize {
			return c.JSON(http.StatusBadRequest, responses.New(false, fmt.Sprintf(services.Tr("Max allowed size is %d", lang), config.Data.ProfilePictureSize), ""))
		}
	}

	groupPicture, err := h.groupStore.GetGroupPicture(group)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if len(groupPicture) == 0 {
		return c.JSON(http.StatusNotFound, responses.New(false, "No group picture set", lang))
	}

	data, err := bimg.NewImage(groupPicture).Thumbnail(size)

	return c.Blob(http.StatusOK, "image/jpeg", data)
}
