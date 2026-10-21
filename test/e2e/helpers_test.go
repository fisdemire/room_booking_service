package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var baseURL string

func TestMain(m *testing.M) {
	baseURL = os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(baseURL + "/_info")
	if err != nil {
		fmt.Fprintf(os.Stderr, "сервис недоступен по адресу %s: %v\nсначала выполни `make up`\n", baseURL, err)
		os.Exit(1)
	}
	resp.Body.Close()

	os.Exit(m.Run())
}

type response struct {
	Status int
	Body   []byte
}

func call(t *testing.T, method, path, token string, body any) response {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, baseURL+path, reader)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return response{Status: resp.StatusCode, Body: raw}
}

func decode[T any](t *testing.T, r response) T {
	t.Helper()

	var out T
	require.NoError(t, json.Unmarshal(r.Body, &out), "тело ответа: %s", r.Body)

	return out
}

func requireError(t *testing.T, r response, status int, code string) {
	t.Helper()

	require.Equal(t, status, r.Status, "тело ответа: %s", r.Body)

	type errBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}

	require.Equal(t, code, decode[errBody](t, r).Error.Code)
}

type tokenResp struct {
	Token string `json:"token"`
}

type registerResp struct {
	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
}

type scheduleResp struct {
	Schedule struct {
		RoomID    string `json:"roomId"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	} `json:"schedule"`
}

type roomResp struct {
	Room struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Capacity int    `json:"capacity"`
	} `json:"room"`
}

type roomsResp struct {
	Rooms []struct {
		ID string `json:"id"`
	} `json:"rooms"`
}

type slot struct {
	ID     string    `json:"id"`
	RoomID string    `json:"roomId"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type slotsResp struct {
	Slots []slot `json:"slots"`
}

type booking struct {
	ID             string  `json:"id"`
	SlotID         string  `json:"slotId"`
	UserID         string  `json:"userId"`
	Status         string  `json:"status"`
	ConferenceLink *string `json:"conferenceLink"`
}

type bookingResp struct {
	Booking booking `json:"booking"`
}

type bookingsResp struct {
	Bookings   []booking `json:"bookings"`
	Pagination struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
		Total    int `json:"total"`
	} `json:"pagination"`
}

func dummyToken(t *testing.T, role string) string {
	t.Helper()

	r := call(t, http.MethodPost, "/dummyLogin", "", map[string]string{"role": role})
	require.Equal(t, http.StatusOK, r.Status, "тело ответа: %s", r.Body)

	return decode[tokenResp](t, r).Token
}

func newUserToken(t *testing.T) string {
	t.Helper()

	email := fmt.Sprintf("e2e-%s@example.com", uuid.NewString())

	r := call(t, http.MethodPost, "/register", "", map[string]string{
		"email": email, "password": "StrongPassword123", "role": "user",
	})
	require.Equal(t, http.StatusCreated, r.Status, "тело ответа: %s", r.Body)

	r = call(t, http.MethodPost, "/login", "", map[string]string{
		"email": email, "password": "StrongPassword123",
	})
	require.Equal(t, http.StatusOK, r.Status, "тело ответа: %s", r.Body)

	return decode[tokenResp](t, r).Token
}

func newRoom(t *testing.T, adminToken string) string {
	t.Helper()

	r := call(t, http.MethodPost, "/rooms/create", adminToken, map[string]any{
		"name": "e2e-room-" + uuid.NewString()[:8], "description": "e2e", "capacity": 6,
	})
	require.Equal(t, http.StatusCreated, r.Status, "тело ответа: %s", r.Body)

	return decode[roomResp](t, r).Room.ID
}

func newRoomWithSchedule(t *testing.T, adminToken string) string {
	t.Helper()

	roomID := newRoom(t, adminToken)

	r := call(t, http.MethodPost, "/rooms/"+roomID+"/schedule/create", adminToken, map[string]any{
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "09:00",
		"endTime":    "18:00",
	})
	require.Equal(t, http.StatusCreated, r.Status, "тело ответа: %s", r.Body)

	return roomID
}

func tomorrow() string {
	return time.Now().UTC().AddDate(0, 0, 1).Format(time.DateOnly)
}

func listSlots(t *testing.T, token, roomID, date string) []slot {
	t.Helper()

	r := call(t, http.MethodGet, "/rooms/"+roomID+"/slots/list?date="+date, token, nil)
	require.Equal(t, http.StatusOK, r.Status, "тело ответа: %s", r.Body)

	return decode[slotsResp](t, r).Slots
}
