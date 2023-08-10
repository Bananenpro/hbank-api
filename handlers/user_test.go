package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/juho05/hbank-api/bindings"
	"github.com/juho05/hbank-api/config"
	"github.com/juho05/hbank-api/db"
	"github.com/juho05/hbank-api/models"
	"github.com/juho05/hbank-api/responses"
	"github.com/juho05/hbank-api/router"
)

func TestHandler_GetUsers(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user1 := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
	}
	us.Create(user1)

	user2 := &models.User{
		Name:  "peter",
		Email: "peter@gmail.com",
	}
	us.Create(user2)

	handler := New(us, nil, nil)

	tests := []struct {
		tName         string
		user          *models.User
		pageSize      int
		exclude       string
		wantCode      int
		wantSuccess   bool
		wantAllInfo   bool
		wantUserCount int
	}{
		{tName: "All", user: user1, pageSize: 10, exclude: "", wantCode: http.StatusOK, wantSuccess: true, wantUserCount: 2},
		{tName: "Don't include self", user: user1, pageSize: 10, exclude: user1.Id, wantCode: http.StatusOK, wantSuccess: true, wantUserCount: 1},
		{tName: "Only 1 user", user: user1, pageSize: 1, exclude: "", wantCode: http.StatusOK, wantSuccess: true, wantUserCount: 1},
		{tName: "Invalid page size", user: user1, pageSize: -1, exclude: "", wantCode: http.StatusBadRequest, wantSuccess: false},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?pageSize=%d&exclude=%s", tt.pageSize, tt.exclude), nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")

			c.Set("userId", tt.user.Id)

			err := handler.GetUsers(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))

			type usersResp struct {
				responses.Base
				Users []models.User `json:"users"`
			}
			var users usersResp
			json.Unmarshal(rec.Body.Bytes(), &users)
			assert.Equal(t, tt.wantUserCount, len(users.Users))
		})
	}
}

func TestHandler_GetUser(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user1 := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
	}
	us.Create(user1)

	user2 := &models.User{
		Name:  "peter",
		Email: "peter@gmail.com",
	}
	us.Create(user2)

	handler := New(us, nil, nil)

	tests := []struct {
		tName       string
		user        *models.User
		wantCode    int
		wantSuccess bool
		wantMessage string
		wantAllInfo bool
	}{
		{tName: "Doesn't exist", wantCode: http.StatusNotFound, wantSuccess: false, wantMessage: "Resource not found"},
		{tName: "Auth user", user: user1, wantCode: http.StatusOK, wantSuccess: true, wantAllInfo: true},
		{tName: "Not auth user", user: user2, wantCode: http.StatusOK, wantSuccess: true, wantAllInfo: false},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")

			c.SetParamNames("id")
			if tt.user != nil {
				c.SetParamValues(tt.user.Id)
			} else {
				c.SetParamValues(uuid.NewString())
			}
			c.Set("userId", user1.Id)

			err := handler.GetUser(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"id":"%s"`, tt.user.Id))
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"name":"%s"`, tt.user.Name))

				if tt.wantAllInfo {
					assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"email":"%s"`, tt.user.Email))
				} else {
					assert.NotContains(t, rec.Body.String(), "email")
				}
			}
		})
	}
}

func TestHandler_UpdateUser(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user1 := &models.User{
		Name:                    "bob",
		Email:                   "bob@gmail.com",
		DontSendInvitationEmail: true,
	}
	us.Create(user1)

	user2 := &models.User{
		Name:                    "bob2",
		Email:                   "bob2@gmail.com",
		DontSendInvitationEmail: true,
	}
	us.Create(user2)

	handler := New(us, nil, nil)

	tests := []struct {
		tName                   string
		user                    *models.User
		dontSendInvitationEmail bool
		wantCode                int
		wantSuccess             bool
		wantMessage             string
	}{
		{tName: "Success", user: user1, dontSendInvitationEmail: false, wantCode: http.StatusOK, wantSuccess: true},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"dontSendInvitationEmail": %t, "email": "bla@bla.bla", "password": "123456"}`, tt.dontSendInvitationEmail)
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")

			c.Set("userId", tt.user.Id)

			err := handler.UpdateUser(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			user, _ := us.GetById(tt.user.Id)
			if tt.wantSuccess {
				assert.Equal(t, tt.dontSendInvitationEmail, user.DontSendInvitationEmail)
			} else {
				assert.NotEqual(t, tt.dontSendInvitationEmail, user.DontSendInvitationEmail)
			}

			assert.Equal(t, tt.user.Email, user.Email)
		})
	}
}

func TestHandler_GetCurrentCash(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user1 := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
		CashLog: []models.CashLogEntry{
			{ChangeTitle: "Change1", Base: models.Base{Created: time.Now().Unix()}},
			{ChangeTitle: "Change2", Base: models.Base{Created: time.Now().Unix() + 100000}},
			{ChangeTitle: "Change3", Base: models.Base{Created: time.Now().Unix()}},
		},
	}
	us.Create(user1)

	user2 := &models.User{
		Name:  "peter",
		Email: "peter@gmail.com",
	}
	us.Create(user2)

	handler := New(us, nil, nil)

	tests := []struct {
		tName       string
		user        *models.User
		wantCode    int
		wantSuccess bool
		wantMessage string
		wantTitle   string
	}{
		{tName: "Success", user: user1, wantCode: http.StatusOK, wantSuccess: true, wantTitle: "Change2"},
		{tName: "Empty cash log", user: user2, wantCode: http.StatusOK, wantSuccess: true, wantTitle: ""},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")
			c.Set("userId", tt.user.Id)

			err := handler.GetCurrentCash(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"title":"%s"`, tt.wantTitle))
			}
		})
	}
}

func TestHandler_GetCashLogEntryById(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
		CashLog: []models.CashLogEntry{
			{ChangeTitle: "Change1", Base: models.Base{Created: time.Now().Unix()}},
			{ChangeTitle: "Change2", Base: models.Base{Created: time.Now().Unix()}},
			{ChangeTitle: "Change3", Base: models.Base{Created: time.Now().Unix()}},
		},
	}
	us.Create(user)

	handler := New(us, nil, nil)

	tests := []struct {
		tName       string
		entryId     string
		wantCode    int
		wantSuccess bool
		wantMessage string
		wantTitle   string
	}{
		{tName: "Success", entryId: user.CashLog[2].Id, wantCode: http.StatusOK, wantSuccess: true, wantTitle: "Change3"},
		{tName: "Doesn't exist", entryId: uuid.NewString(), wantCode: http.StatusNotFound, wantSuccess: false, wantMessage: "Resource not found"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")
			c.Set("userId", user.Id)
			c.SetParamNames("id")
			c.SetParamValues(tt.entryId)

			err := handler.GetCashLogEntryById(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"title":"%s"`, tt.wantTitle))
			}
		})
	}
}

func TestHandler_GetCashLog(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
		CashLog: []models.CashLogEntry{
			{ChangeTitle: "Change1", Base: models.Base{Created: time.Now().Unix()}},
			{ChangeTitle: "Change2", Base: models.Base{Created: time.Now().Unix()}},
			{ChangeTitle: "Change3", Base: models.Base{Created: time.Now().Unix()}},
		},
	}
	us.Create(user)

	handler := New(us, nil, nil)

	tests := []struct {
		tName       string
		pageSize    int
		wantCode    int
		wantSuccess bool
		wantMessage string
		wantCount   int
	}{
		{tName: "Get all", pageSize: 10, wantCode: http.StatusOK, wantSuccess: true, wantCount: 3},
		{tName: "Get 1", pageSize: 1, wantCode: http.StatusOK, wantSuccess: true, wantCount: 1},
		{tName: "Invalid pageSize", pageSize: -1, wantCode: http.StatusBadRequest, wantSuccess: false, wantMessage: "Unsupported page size"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?pageSize=%d", tt.pageSize), nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")
			c.Set("userId", user.Id)
			c.SetParamNames("id")

			err := handler.GetCashLog(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			if tt.wantSuccess {
				type cashLogResp struct {
					responses.Base
					CashLog []models.CashLogEntry `json:"log"`
				}
				var resp cashLogResp
				json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.Equal(t, tt.wantCount, len(resp.CashLog))
			}
		})
	}
}

func TestHandler_AddCashLogEntry(t *testing.T) {
	t.Parallel()
	config.Data.Debug = true
	r := router.New()

	database, dbId, err := db.NewTestDB()
	if err != nil {
		t.Fatalf("Couldn't create test database")
	}
	defer db.DeleteTestDB(dbId)
	err = db.AutoMigrate(database)
	if err != nil {
		t.Fatalf("Couldn't auto migrate database")
	}

	us := db.NewUserStore(database)

	user1 := &models.User{
		Name:  "bob",
		Email: "bob@gmail.com",
	}
	us.Create(user1)

	user2 := &models.User{
		Name:  "bob2",
		Email: "bob2@gmail.com",
	}
	us.Create(user2)

	handler := New(us, nil, nil)

	tests := []struct {
		tName       string
		user        *models.User
		entry       bindings.AddCashLogEntry
		wantCode    int
		wantSuccess bool
		wantMessage string
	}{
		{tName: "Success", user: user1, entry: bindings.AddCashLogEntry{Title: "Test"}, wantCode: http.StatusCreated, wantSuccess: true, wantMessage: "Successfully added new cash log entry"},
		{tName: "Title too short", user: user2, entry: bindings.AddCashLogEntry{Title: "    hi   "}, wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Title too short"},
		{tName: "Title too long", user: user2, entry: bindings.AddCashLogEntry{Title: "12345678901234567890123456789012"}, wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Title too long"},
		{tName: "Description too long", user: user2, entry: bindings.AddCashLogEntry{Title: "Test", Description: strings.Repeat("a", 257)}, wantCode: http.StatusOK, wantSuccess: false, wantMessage: "Description too long"},
	}
	for _, tt := range tests {
		t.Run(tt.tName, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.entry)
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(string(jsonBody)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			c.Set("lang", "en")

			c.Set("userId", tt.user.Id)

			err := handler.AddCashLogEntry(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"success":%t`, tt.wantSuccess))
			assert.Contains(t, rec.Body.String(), fmt.Sprintf(`"message":"%s"`, tt.wantMessage))

			log, _ := us.GetCashLog(tt.user, "", 0, 10, false)
			if tt.wantSuccess {
				assert.Equal(t, 1, len(log))
			} else {
				assert.Equal(t, 0, len(log))
			}
		})
	}
}
