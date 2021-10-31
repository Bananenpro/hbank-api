package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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
		if pageSize > config.Data.MaxPageSize || pageSize < 1 {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Unsupported page size", lang))
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
		if pageSize > config.Data.MaxPageSize || pageSize < 1 {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Unsupported page size", lang))
		}
	}

	descending := services.StrToBool(c.QueryParam("descending"))
	includeSelf := services.StrToBool(c.QueryParam("includeSelf"))

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

	var members []models.User
	if includeSelf {
		members, err = h.groupStore.GetMembers(nil, group, page, pageSize, descending)
	} else {
		members, err = h.groupStore.GetMembers(user, group, page, pageSize, descending)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewUsers(members))
}

// /v1/group/:id/member (DELETE)
func (h *Handler) LeaveGroup(c echo.Context) error {
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
	if !isMember {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not a member of the group", lang))
	}

	err = h.groupStore.RemoveMember(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully left group", lang))
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
		if pageSize > config.Data.MaxPageSize || pageSize < 1 {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Unsupported page size", lang))
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

	includeSelf := services.StrToBool(c.QueryParam("includeSelf"))
	var members []models.User
	if includeSelf {
		members, err = h.groupStore.GetAdmins(nil, group, page, pageSize, descending)
	} else {
		members, err = h.groupStore.GetAdmins(user, group, page, pageSize, descending)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewUsers(members))
}

// /v1/group/:id/admin (POST)
func (h *Handler) AddGroupAdmin(c echo.Context) error {
	lang := c.Get("lang").(string)
	authUserId := c.Get("userId").(uuid.UUID)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if authUser == nil {
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

	authIsAdmin, err := h.groupStore.IsAdmin(group, authUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if !authIsAdmin {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
	}

	var body bindings.Id
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	userId, err := uuid.Parse(body.Id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusOK, responses.New(false, "The user doesn't exist", lang))
	}

	isMember, err := h.groupStore.IsMember(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if !isMember {
		return c.JSON(http.StatusOK, responses.New(false, "The user is not a member of the group", lang))
	}

	isAdmin, err := h.groupStore.IsAdmin(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if isAdmin {
		return c.JSON(http.StatusOK, responses.New(false, "The user already is an admin of the group", lang))
	}

	err = h.groupStore.AddAdmin(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully made user an admin", lang))
}

// /v1/group/:id/admin (DELETE)
func (h *Handler) RemoveAdminRights(c echo.Context) error {
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

	isAdmin, err := h.groupStore.IsAdmin(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if !isAdmin {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
	}

	userCount, err := h.groupStore.GetUserCount(group)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	isMember, err := h.groupStore.IsMember(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	admins, err := h.groupStore.GetAdmins(nil, group, 0, 2, false)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if (userCount > 1 && len(admins) == 1) || (userCount == 1 && isMember) {
		return c.JSON(http.StatusOK, responses.New(false, "Cannot remove admin rights of sole admin of group", lang))
	}

	if userCount == 1 {
		err = h.groupStore.Delete(group)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		return c.JSON(http.StatusOK, responses.New(true, "Successfully deleted group", lang))
	}

	err = h.groupStore.RemoveAdmin(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully removed admin rights", lang))
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

// /v1/group/:id/picture?id=uuid (POST)
func (h *Handler) SetGroupPicture(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	file, err := c.FormFile("group_picture")
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing group picture file", lang))
	}

	if file.Size > config.Data.MaxProfilePictureFileSize {
		return c.JSON(http.StatusBadRequest, responses.New(false, fmt.Sprintf(services.Tr("File too big (max %d bytes)", lang), config.Data.MaxProfilePictureFileSize), ""))
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	defer src.Close()

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	data := buffer.Bytes()

	img := bimg.NewImage(data)
	if img.Type() == "unknown" {
		return c.JSON(http.StatusBadRequest, responses.New(false, "File is not an image", lang))
	}

	data, err = img.Convert(bimg.JPEG)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	data, err = bimg.NewImage(data).AutoRotate()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	data, err = bimg.NewImage(data).Thumbnail(config.Data.ProfilePictureSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	isAdmin, err := h.groupStore.IsAdmin(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if !isAdmin {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
	}

	group.GroupPicture = data
	group.GroupPictureId = uuid.New()
	h.groupStore.Update(group)

	return c.JSON(http.StatusOK, responses.New(true, "Successfully updated group picture", lang))
}

// /v1/group/:id/transaction/balance (GET)
func (h *Handler) GetBalance(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	isMember, err := h.groupStore.IsMember(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if !isMember {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not a member of the group", lang))
	}

	balance, err := h.groupStore.GetUserBalance(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.Balance{
		Base: responses.Base{
			Success: true,
		},
		Balance: balance,
	})
}

// /v1/group/:id/transaction/:transactionId (GET)
func (h *Handler) GetTransactionById(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	transactionId, err := uuid.Parse(c.Param("transactionId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing transactionId parameter", lang))
	}

	transaction, err := h.groupStore.GetTransactionLogEntryById(group, transactionId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if transaction == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	isSender := bytes.Equal(user.Id[:], transaction.SenderId[:])
	isReceiver := bytes.Equal(user.Id[:], transaction.ReceiverId[:])

	if isSender || isReceiver {
		isMember, err := h.groupStore.IsMember(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if !isMember {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not a member of the group", lang))
		}

		return c.JSON(http.StatusOK, responses.NewTransaction(transaction, user))
	} else if transaction.SenderIsBank || transaction.ReceiverIsBank {
		isAdmin, err := h.groupStore.IsAdmin(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if !isAdmin {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
		}

		return c.JSON(http.StatusOK, responses.NewBankTransaction(transaction))
	}

	return c.JSON(http.StatusForbidden, responses.New(false, "User not allowed to view transaction", lang))

}

// /v1/group/:id/transaction?bank=bool&page=int&pageSize=int&oldestFirst=bool (GET)
func (h *Handler) GetTransactionLog(c echo.Context) error {
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

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	bank := services.StrToBool(c.QueryParam("bank"))

	if !bank {
		isMember, err := h.groupStore.IsMember(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		if !isMember {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not a member of the group", lang))
		}

		log, err := h.groupStore.GetTransactionLog(group, user, page, pageSize, oldestFirst)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		return c.JSON(http.StatusOK, responses.NewTransactionLog(log, user))
	} else {
		isAdmin, err := h.groupStore.IsAdmin(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		if !isAdmin {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
		}

		log, err := h.groupStore.GetBankTransactionLog(group, page, pageSize, oldestFirst)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		return c.JSON(http.StatusOK, responses.NewBankTransactionLog(log))
	}
}

// /v1/group/:id/transaction (POST)
func (h *Handler) CreateTransaction(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	var body bindings.CreateTransaction
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}
	if body.Amount <= 0 {
		return c.JSON(http.StatusOK, responses.New(false, "Amount must be >0", lang))
	}

	body.Title = strings.TrimSpace(body.Title)
	body.Description = strings.TrimSpace(body.Description)

	if len(body.Title) > config.Data.MaxNameLength {
		return c.JSON(http.StatusOK, responses.New(false, "Title too long", lang))
	}

	if utf8.RuneCountInString(body.Title) < config.Data.MinNameLength {
		return c.JSON(http.StatusOK, responses.New(false, "Title too short", lang))
	}

	if len(body.Description) > config.Data.MaxDescriptionLength {
		return c.JSON(http.StatusOK, responses.New(false, "Description too long", lang))
	}

	if utf8.RuneCountInString(body.Description) < config.Data.MinDescriptionLength {
		return c.JSON(http.StatusOK, responses.New(false, "Description too short", lang))
	}

	if !body.FromBank {
		isMember, err := h.groupStore.IsMember(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if !isMember {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not a member of the group", lang))
		}

		balanceSender, err := h.groupStore.GetUserBalance(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		if balanceSender-int(body.Amount) < 0 {
			return c.JSON(http.StatusOK, responses.New(false, "Not enough money", lang))
		}
	}

	if strings.EqualFold(body.ReceiverId, "bank") {
		if body.FromBank {
			return c.JSON(http.StatusOK, responses.New(false, "Cannot send money from bank to bank", lang))
		}
		err = h.groupStore.CreateTransaction(group, false, true, user, nil, body.Title, body.Description, int(body.Amount))
	} else {
		receiverId, err := uuid.Parse(body.ReceiverId)
		if err != nil {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid receiver id", lang))
		}

		receiver, err := h.userStore.GetById(receiverId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if receiver == nil {
			return c.JSON(http.StatusNotFound, responses.New(false, "Couldn't find receiver", lang))
		}
		isReceiverMember, err := h.groupStore.IsMember(group, receiver)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if !isReceiverMember {
			return c.JSON(http.StatusForbidden, responses.New(false, "Receiver not a member of the group", lang))
		}

		if body.FromBank {
			isAdmin, err := h.groupStore.IsAdmin(group, user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}
			if !isAdmin {
				return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
			}
			err = h.groupStore.CreateTransaction(group, true, false, nil, receiver, body.Title, body.Description, int(body.Amount))
		} else {
			if bytes.Equal(user.Id[:], receiverId[:]) {
				return c.JSON(http.StatusOK, responses.New(false, "Sender is the receiver", lang))
			}
			err = h.groupStore.CreateTransaction(group, false, false, user, receiver, body.Title, body.Description, int(body.Amount))
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}
		}
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully completed transaction", lang))
}

// /v1/group/invitation?page=int&pageSize=int&oldestFirst=bool (GET)
func (h *Handler) GetInvitationsByUser(c echo.Context) error {
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

	invitations, err := h.groupStore.GetInvitationsByUser(user, page, pageSize, oldestFirst)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewInvitations(invitations))
}

// /v1/group/:id/invitation?page=int&pageSize=int&oldestFirst=bool (GET)
func (h *Handler) GetInvitationsByGroup(c echo.Context) error {
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

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	isAdmin, err := h.groupStore.IsAdmin(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if !isAdmin {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
	}

	invitations, err := h.groupStore.GetInvitationsByGroup(group, page, pageSize, oldestFirst)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewInvitations(invitations))
}

// /v1/group/invitation/:id (GET)
func (h *Handler) GetInvitationById(c echo.Context) error {
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

	invitation, err := h.groupStore.GetInvitationById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if invitation == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	group, err := h.groupStore.GetById(invitation.GroupId)
	if err != nil || group == nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if !bytes.Equal(userId[:], invitation.UserId[:]) {
		isAdmin, err := h.groupStore.IsAdmin(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if !isAdmin {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
		}
	}

	return c.JSON(http.StatusOK, responses.NewInvitation(invitation))
}

// /v1/group/:id/invitation (POST)
func (h *Handler) CreateInvitation(c echo.Context) error {
	lang := c.Get("lang").(string)

	authUserId := c.Get("userId").(uuid.UUID)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if authUser == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	var body bindings.CreateInvitation
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if len(body.Message) > config.Data.MaxDescriptionLength {
		return c.JSON(http.StatusOK, responses.New(false, "Message too long", lang))
	}

	if utf8.RuneCountInString(body.Message) < config.Data.MinDescriptionLength {
		return c.JSON(http.StatusOK, responses.New(false, "Message too short", lang))
	}

	userId, err := uuid.Parse(body.UserId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if bytes.Equal(userId[:], authUserId[:]) {
		return c.JSON(http.StatusOK, responses.New(false, "You can't invite yourself", lang))
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusOK, responses.New(false, "The user doesn't exist", lang))
	}

	userIsInGroup, err := h.groupStore.IsInGroup(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if userIsInGroup {
		return c.JSON(http.StatusOK, responses.New(false, "The user is already a member/an admin of the group", lang))
	}

	authUserIsAdmin, err := h.groupStore.IsAdmin(group, authUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if !authUserIsAdmin {
		return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
	}

	invitation, err := h.groupStore.GetInvitationByGroupAndUser(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if invitation != nil {
		return c.JSON(http.StatusOK, responses.New(false, "The user was already invited", lang))
	}

	err = h.groupStore.CreateInvitation(group, user, body.Message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if !user.DontSendInvitationEmail && config.Data.EmailEnabled {
		type templateData struct {
			Name           string
			GroupName      string
			InvitationsUrl string
		}
		body, err := services.ParseEmailTemplate("invitation", c.Get("lang").(string), templateData{
			Name:           user.Name,
			GroupName:      group.Name,
			InvitationsUrl: fmt.Sprintf("https://%s/invitations", config.Data.DomainName),
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		go services.SendEmail([]string{user.Email}, services.Tr("H-Bank Invitation", lang), body)
	}

	return c.JSON(http.StatusCreated, responses.New(true, "Successfully invited user", lang))
}

// /v1/group/invitation/:id (POST)
func (h *Handler) AcceptInvitation(c echo.Context) error {
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

	invitation, err := h.groupStore.GetInvitationById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if invitation == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	group, err := h.groupStore.GetById(invitation.GroupId)
	if err != nil || group == nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if !bytes.Equal(userId[:], invitation.UserId[:]) {
		return c.JSON(http.StatusForbidden, responses.New(false, "User is not the receiver of the invitation", lang))
	}

	isInGroup, err := h.groupStore.IsInGroup(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if isInGroup {
		return c.JSON(http.StatusOK, responses.New(false, "The user is already a member/an admin of the group", lang))
	}

	err = h.groupStore.AddMember(group, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	err = h.groupStore.DeleteInvitation(invitation)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully accepted invitation", lang))
}

// /v1/group/invitation/:id (DELETE)
func (h *Handler) DenyInvitation(c echo.Context) error {
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

	invitation, err := h.groupStore.GetInvitationById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if invitation == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	group, err := h.groupStore.GetById(invitation.GroupId)
	if err != nil || group == nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if !bytes.Equal(userId[:], invitation.UserId[:]) {
		return c.JSON(http.StatusForbidden, responses.New(false, "User is not the receiver of the invitation", lang))
	}

	err = h.groupStore.DeleteInvitation(invitation)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully denied invitation", lang))
}

// /v1/group/:id/paymentPlan/:paymentPlanId (GET)
func (h *Handler) GetPaymentPlanById(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	paymentPlanId, err := uuid.Parse(c.Param("paymentPlanId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing paymentPlanId parameter", lang))
	}

	paymentPlan, err := h.groupStore.GetPaymentPlanById(group, paymentPlanId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if paymentPlan == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	isSender := bytes.Equal(user.Id[:], paymentPlan.SenderId[:])
	isReceiver := bytes.Equal(user.Id[:], paymentPlan.ReceiverId[:])

	if isSender || isReceiver {
		isMember, err := h.groupStore.IsMember(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if !isMember {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not a member of the group", lang))
		}

		return c.JSON(http.StatusOK, responses.NewPaymentPlan(paymentPlan))
	} else if paymentPlan.SenderIsBank || paymentPlan.ReceiverIsBank {
		isAdmin, err := h.groupStore.IsAdmin(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if !isAdmin {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
		}

		return c.JSON(http.StatusOK, responses.NewPaymentPlan(paymentPlan))
	}

	return c.JSON(http.StatusForbidden, responses.New(false, "User not allowed to view payment plan", lang))

}

// /v1/group/:id/paymentPlan?bank=bool&page=int&pageSize=int&oldestFirst=bool (GET)
func (h *Handler) GetPaymentPlans(c echo.Context) error {
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

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	bank := services.StrToBool(c.QueryParam("bank"))

	if !bank {
		isMember, err := h.groupStore.IsMember(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		if !isMember {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not a member of the group", lang))
		}

		paymentPlans, err := h.groupStore.GetPaymentPlans(group, user, page, pageSize, oldestFirst)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		return c.JSON(http.StatusOK, responses.NewPaymentPlans(paymentPlans))
	} else {
		isAdmin, err := h.groupStore.IsAdmin(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		if !isAdmin {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
		}

		paymentPlans, err := h.groupStore.GetBankPaymentPlans(group, page, pageSize, oldestFirst)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}

		return c.JSON(http.StatusOK, responses.NewPaymentPlans(paymentPlans))
	}
}

// /v1/group/:id/paymentPlan/nextPayment?id=uuid&firstPayment=int&schedule=int&scheduleUnit=string&count=int
func (h *Handler) GetPaymentPlanNextPayments(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	count := 1
	if c.QueryParam("count") != "" {
		count, err = strconv.Atoi(c.QueryParam("count"))
		if err != nil || count < 1 {
			return c.JSON(http.StatusBadRequest, responses.New(false, "'count' query parameter not a number or <1", lang))
		}
		if count > config.Data.MaxPageSize {
			return c.JSON(http.StatusBadRequest, responses.New(false, "'count' query parameter too big", lang))
		}
	}

	schedule := -1
	scheduleUnit := ""
	firstPayment := int64(-1)

	if c.QueryParam("id") != "" {
		id, err := uuid.Parse(c.QueryParam("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid 'id' query parameter", lang))
		}

		paymentPlan, err := h.groupStore.GetPaymentPlanById(group, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if paymentPlan == nil {
			return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
		}

		isSender := bytes.Equal(user.Id[:], paymentPlan.SenderId[:])
		isReceiver := bytes.Equal(user.Id[:], paymentPlan.ReceiverId[:])

		if isSender || isReceiver {
			isMember, err := h.groupStore.IsMember(group, user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}
			if !isMember {
				return c.JSON(http.StatusForbidden, responses.New(false, "Not a member of the group", lang))
			}
		} else if paymentPlan.SenderIsBank || paymentPlan.ReceiverIsBank {
			isAdmin, err := h.groupStore.IsAdmin(group, user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}
			if !isAdmin {
				return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
			}
		} else {
			return c.JSON(http.StatusForbidden, responses.New(false, "User not allowed to view payment plan", lang))
		}

		schedule = paymentPlan.Schedule
		scheduleUnit = paymentPlan.ScheduleUnit
		firstPayment = paymentPlan.NextExecute
	} else {
		if c.QueryParam("schedule") != "" {
			schedule, err = strconv.Atoi(c.QueryParam("schedule"))
			if err != nil || schedule < 1 {
				return c.JSON(http.StatusBadRequest, responses.New(false, "'schedule' query parameter not a number or <1", lang))
			}
		} else {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Missing 'schedule' or 'id' query parameter", lang))
		}

		scheduleUnit = strings.ToLower(c.QueryParam("scheduleUnit"))
		if scheduleUnit != models.ScheduleUnitDay && scheduleUnit != models.ScheduleUnitWeek && scheduleUnit != models.ScheduleUnitMonth && scheduleUnit != models.ScheduleUnitYear {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid schedule unit", lang))
		}

		if c.QueryParam("firstPayment") != "" {
			firstPayment, err = strconv.ParseInt(c.QueryParam("firstPayment"), 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, responses.New(false, "'firstPayment' query parameter not a number", lang))
			}
		} else {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Missing 'firstPayment' or 'id' query parameter", lang))
		}
	}

	executionTimes := make([]int64, count)
	for i := 0; i < count; i++ {
		executionTimes[i] = firstPayment
		firstPayment = services.AddTime(firstPayment, schedule, scheduleUnit)
	}

	return c.JSON(http.StatusOK, responses.PaymentPlanExecutionTimes{
		Base: responses.Base{
			Success: true,
		},
		ExecutionTimes: executionTimes,
	})
}

// /v1/group/:id/paymentPlan (POST)
func (h *Handler) CreatePaymentPlan(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	var body bindings.CreatePaymentPlan
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if body.Amount <= 0 {
		return c.JSON(http.StatusOK, responses.New(false, "Amount must be >0", lang))
	}

	if body.Schedule <= 0 {
		return c.JSON(http.StatusOK, responses.New(false, "Schedule must be >0", lang))
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

	body.ScheduleUnit = strings.ToLower(body.ScheduleUnit)

	if body.ScheduleUnit != models.ScheduleUnitDay && body.ScheduleUnit != models.ScheduleUnitWeek && body.ScheduleUnit != models.ScheduleUnitMonth && body.ScheduleUnit != models.ScheduleUnitYear {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid schedule unit", lang))
	}

	firstPayment, err := time.Parse("2006-01-02", body.FirstPayment)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid date string", lang))
	}
	if firstPayment.Before(time.Now()) {
		return c.JSON(http.StatusOK, responses.New(false, "First payment can't be in the past", lang))
	}

	if body.PaymentCount == 0 {
		body.PaymentCount = -1
	}

	if !body.FromBank {
		isMember, err := h.groupStore.IsMember(group, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if !isMember {
			return c.JSON(http.StatusForbidden, responses.New(false, "Not a member of the group", lang))
		}
	}

	if strings.EqualFold(body.ReceiverId, "bank") {
		if body.FromBank {
			return c.JSON(http.StatusOK, responses.New(false, "Cannot send money from bank to bank", lang))
		}
		err = h.groupStore.CreatePaymentPlan(group, false, true, user, nil, body.Name, body.Description, int(body.Amount), body.PaymentCount, int(body.Schedule), body.ScheduleUnit, firstPayment.Unix())
	} else {
		receiverId, err := uuid.Parse(body.ReceiverId)
		if err != nil {
			return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid receiver id", lang))
		}

		receiver, err := h.userStore.GetById(receiverId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if receiver == nil {
			return c.JSON(http.StatusNotFound, responses.New(false, "Couldn't find receiver", lang))
		}
		isReceiverMember, err := h.groupStore.IsMember(group, receiver)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
		}
		if !isReceiverMember {
			return c.JSON(http.StatusForbidden, responses.New(false, "Receiver not a member of the group", lang))
		}

		if body.FromBank {
			isAdmin, err := h.groupStore.IsAdmin(group, user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}
			if !isAdmin {
				return c.JSON(http.StatusForbidden, responses.New(false, "Not an admin of the group", lang))
			}
			err = h.groupStore.CreatePaymentPlan(group, true, false, nil, receiver, body.Name, body.Description, int(body.Amount), body.PaymentCount, int(body.Schedule), body.ScheduleUnit, firstPayment.Unix())
		} else {
			if bytes.Equal(user.Id[:], receiverId[:]) {
				return c.JSON(http.StatusOK, responses.New(false, "Sender is the receiver", lang))
			}
			err = h.groupStore.CreatePaymentPlan(group, false, false, user, receiver, body.Name, body.Description, int(body.Amount), body.PaymentCount, int(body.Schedule), body.ScheduleUnit, firstPayment.Unix())
			if err != nil {
				return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
			}
		}
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully created payment plan", lang))
}

// /v1/group/:id/paymentPlan/:paymentPlanId (DELETE)
func (h *Handler) DeletePaymentPlan(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	paymentPlanId, err := uuid.Parse(c.Param("paymentPlanId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing paymentPlanId parameter", lang))
	}

	paymentPlan, err := h.groupStore.GetPaymentPlanById(group, paymentPlanId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if paymentPlan == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	isSender := bytes.Equal(user.Id[:], paymentPlan.SenderId[:])
	if !isSender {
		return c.JSON(http.StatusForbidden, responses.New(false, "User not the sender of the payment plan", lang))
	}

	err = h.groupStore.DeletePaymentPlan(paymentPlan)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully deleted payment plan", lang))
}

// /v1/group/:id/paymentPlan/:paymentPlanId (PUT)
func (h *Handler) UpdatePaymentPlan(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	groupId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}
	group, err := h.groupStore.GetById(groupId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if group == nil {
		return c.JSON(http.StatusNotFound, responses.New(false, "Group not found", lang))
	}

	paymentPlanId, err := uuid.Parse(c.Param("paymentPlanId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing paymentPlanId parameter", lang))
	}

	paymentPlan, err := h.groupStore.GetPaymentPlanById(group, paymentPlanId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if paymentPlan == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	isSender := bytes.Equal(user.Id[:], paymentPlan.SenderId[:])
	if !isSender {
		return c.JSON(http.StatusForbidden, responses.New(false, "User not the sender of the payment plan", lang))
	}

	var body bindings.UpdatePaymentPlan
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if body.Amount <= 0 {
		return c.JSON(http.StatusOK, responses.New(false, "Amount must be >0", lang))
	}

	if body.Schedule <= 0 {
		return c.JSON(http.StatusOK, responses.New(false, "Schedule must be >0", lang))
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

	body.ScheduleUnit = strings.ToLower(body.ScheduleUnit)

	if body.ScheduleUnit != models.ScheduleUnitDay && body.ScheduleUnit != models.ScheduleUnitWeek && body.ScheduleUnit != models.ScheduleUnitMonth && body.ScheduleUnit != models.ScheduleUnitYear {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid schedule unit", lang))
	}

	paymentPlan.Amount = int(body.Amount)
	paymentPlan.Name = body.Name
	paymentPlan.Description = body.Description
	paymentPlan.Schedule = int(body.Schedule)
	paymentPlan.ScheduleUnit = body.ScheduleUnit

	err = h.groupStore.UpdatePaymentPlan(paymentPlan)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.New(true, "Successfully updated payment plan", lang))
}
