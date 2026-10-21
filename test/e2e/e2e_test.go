package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	r := call(t, http.MethodGet, "/_info", "", nil)

	require.Equal(t, http.StatusOK, r.Status)
	require.JSONEq(t, `{"status":"ok"}`, string(r.Body))
}

func TestAuth(t *testing.T) {
	t.Run("dummyLogin выдаёт токены для обеих ролей", func(t *testing.T) {
		require.NotEmpty(t, dummyToken(t, "user"))
		require.NotEmpty(t, dummyToken(t, "admin"))
	})

	t.Run("dummyLogin с неизвестной ролью — 400", func(t *testing.T) {
		r := call(t, http.MethodPost, "/dummyLogin", "", map[string]string{"role": "hacker"})

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	t.Run("регистрация и вход", func(t *testing.T) {
		email := "e2e-" + uuid.NewString() + "@example.com"
		creds := map[string]string{"email": email, "password": "StrongPassword123"}

		r := call(t, http.MethodPost, "/register", "", map[string]string{
			"email": email, "password": "StrongPassword123", "role": "user",
		})
		require.Equal(t, http.StatusCreated, r.Status, "тело ответа: %s", r.Body)

		registered := decode[registerResp](t, r)
		require.Equal(t, email, registered.User.Email)
		require.Equal(t, "user", registered.User.Role)
		require.NotEmpty(t, registered.User.ID)

		r = call(t, http.MethodPost, "/login", "", creds)
		require.Equal(t, http.StatusOK, r.Status, "тело ответа: %s", r.Body)
		require.NotEmpty(t, decode[tokenResp](t, r).Token)
	})

	t.Run("повторная регистрация с тем же email — 400", func(t *testing.T) {
		email := "e2e-" + uuid.NewString() + "@example.com"
		body := map[string]string{"email": email, "password": "StrongPassword123", "role": "user"}

		require.Equal(t, http.StatusCreated, call(t, http.MethodPost, "/register", "", body).Status)

		r := call(t, http.MethodPost, "/register", "", body)
		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	t.Run("регистрация с невалидным email — 400", func(t *testing.T) {
		r := call(t, http.MethodPost, "/register", "", map[string]string{
			"email": "not-an-email", "password": "StrongPassword123", "role": "user",
		})

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	t.Run("неверный пароль — 401", func(t *testing.T) {
		email := "e2e-" + uuid.NewString() + "@example.com"
		require.Equal(t, http.StatusCreated, call(t, http.MethodPost, "/register", "", map[string]string{
			"email": email, "password": "StrongPassword123", "role": "user",
		}).Status)

		r := call(t, http.MethodPost, "/login", "", map[string]string{
			"email": email, "password": "wrong-password",
		})

		requireError(t, r, http.StatusUnauthorized, "UNAUTHORIZED")
	})

	t.Run("вход несуществующего пользователя — 401", func(t *testing.T) {
		r := call(t, http.MethodPost, "/login", "", map[string]string{
			"email": "nobody-" + uuid.NewString() + "@example.com", "password": "StrongPassword123",
		})

		requireError(t, r, http.StatusUnauthorized, "UNAUTHORIZED")
	})
}

func TestAccessControl(t *testing.T) {
	admin := dummyToken(t, "admin")
	user := dummyToken(t, "user")

	t.Run("без токена — 401", func(t *testing.T) {
		r := call(t, http.MethodGet, "/rooms/list", "", nil)

		requireError(t, r, http.StatusUnauthorized, "UNAUTHORIZED")
	})

	t.Run("мусорный токен — 401", func(t *testing.T) {
		r := call(t, http.MethodGet, "/rooms/list", "garbage.token.value", nil)

		requireError(t, r, http.StatusUnauthorized, "UNAUTHORIZED")
	})

	t.Run("user не может создавать комнаты — 403", func(t *testing.T) {
		r := call(t, http.MethodPost, "/rooms/create", user, map[string]any{"name": "x", "capacity": 2})

		requireError(t, r, http.StatusForbidden, "FORBIDDEN")
	})

	t.Run("user не может создавать расписание — 403", func(t *testing.T) {
		r := call(t, http.MethodPost, "/rooms/"+uuid.NewString()+"/schedule/create", user, map[string]any{
			"daysOfWeek": []int{1}, "startTime": "09:00", "endTime": "10:00",
		})

		requireError(t, r, http.StatusForbidden, "FORBIDDEN")
	})

	t.Run("user не видит общий список броней — 403", func(t *testing.T) {
		r := call(t, http.MethodGet, "/bookings/list", user, nil)

		requireError(t, r, http.StatusForbidden, "FORBIDDEN")
	})

	t.Run("admin не может бронировать — 403", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/create", admin, map[string]any{"slotId": uuid.NewString()})

		requireError(t, r, http.StatusForbidden, "FORBIDDEN")
	})

	t.Run("admin не имеет своих броней — 403", func(t *testing.T) {
		r := call(t, http.MethodGet, "/bookings/my", admin, nil)

		requireError(t, r, http.StatusForbidden, "FORBIDDEN")
	})
}

func TestRooms(t *testing.T) {
	admin := dummyToken(t, "admin")
	user := dummyToken(t, "user")

	t.Run("admin создаёт комнату, и она появляется в списке", func(t *testing.T) {
		r := call(t, http.MethodPost, "/rooms/create", admin, map[string]any{
			"name": "Большая переговорная", "description": "На 8 человек", "capacity": 8,
		})
		require.Equal(t, http.StatusCreated, r.Status, "тело ответа: %s", r.Body)

		created := decode[roomResp](t, r).Room
		require.NotEmpty(t, created.ID)
		require.Equal(t, "Большая переговорная", created.Name)
		require.Equal(t, 8, created.Capacity)

		r = call(t, http.MethodGet, "/rooms/list", user, nil)
		require.Equal(t, http.StatusOK, r.Status)

		var found bool
		for _, room := range decode[roomsResp](t, r).Rooms {
			if room.ID == created.ID {
				found = true
			}
		}
		require.True(t, found, "созданная комната должна быть в списке")
	})

	t.Run("пустое название — 400", func(t *testing.T) {
		r := call(t, http.MethodPost, "/rooms/create", admin, map[string]any{"name": "", "capacity": 4})

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	t.Run("вместимость меньше 1 — 400", func(t *testing.T) {
		r := call(t, http.MethodPost, "/rooms/create", admin, map[string]any{"name": "Тесная", "capacity": 0})

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})
}

func TestSchedule(t *testing.T) {
	admin := dummyToken(t, "admin")

	valid := map[string]any{
		"daysOfWeek": []int{1, 2, 3, 4, 5},
		"startTime":  "09:00",
		"endTime":    "18:00",
	}

	t.Run("создание расписания и запрет повторного — 201, затем 409", func(t *testing.T) {
		roomID := newRoom(t, admin)

		r := call(t, http.MethodPost, "/rooms/"+roomID+"/schedule/create", admin, valid)
		require.Equal(t, http.StatusCreated, r.Status, "тело ответа: %s", r.Body)

		created := decode[scheduleResp](t, r)
		require.Equal(t, roomID, created.Schedule.RoomID)
		require.Equal(t, "09:00", created.Schedule.StartTime)
		require.Equal(t, "18:00", created.Schedule.EndTime)

		r = call(t, http.MethodPost, "/rooms/"+roomID+"/schedule/create", admin, valid)
		requireError(t, r, http.StatusConflict, "SCHEDULE_EXISTS")
	})

	t.Run("несуществующая комната — 404", func(t *testing.T) {
		r := call(t, http.MethodPost, "/rooms/"+uuid.NewString()+"/schedule/create", admin, valid)

		requireError(t, r, http.StatusNotFound, "ROOM_NOT_FOUND")
	})

	t.Run("roomId не UUID — 400", func(t *testing.T) {
		r := call(t, http.MethodPost, "/rooms/not-a-uuid/schedule/create", admin, valid)

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	invalid := map[string]map[string]any{
		"конец раньше начала":      {"daysOfWeek": []int{1}, "startTime": "18:00", "endTime": "09:00"},
		"интервал меньше 30 минут": {"daysOfWeek": []int{1}, "startTime": "09:00", "endTime": "09:15"},
		"интервал не кратен 30":    {"daysOfWeek": []int{1}, "startTime": "09:00", "endTime": "10:10"},
		"день недели вне 1..7":     {"daysOfWeek": []int{0}, "startTime": "09:00", "endTime": "10:00"},
		"пустой список дней":       {"daysOfWeek": []int{}, "startTime": "09:00", "endTime": "10:00"},
		"неверный формат времени":  {"daysOfWeek": []int{1}, "startTime": "9am", "endTime": "10:00"},
	}

	for name, body := range invalid {
		t.Run("невалидное расписание: "+name+" — 400", func(t *testing.T) {
			roomID := newRoom(t, admin)

			r := call(t, http.MethodPost, "/rooms/"+roomID+"/schedule/create", admin, body)

			requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
		})
	}
}

func TestSlots(t *testing.T) {
	admin := dummyToken(t, "admin")
	user := dummyToken(t, "user")

	roomID := newRoomWithSchedule(t, admin)

	t.Run("слоты на завтра: 18 получасовых слотов с 09:00 до 18:00", func(t *testing.T) {
		slots := listSlots(t, user, roomID, tomorrow())

		require.Len(t, slots, 18)
		for _, s := range slots {
			require.Equal(t, roomID, s.RoomID)
			require.Equal(t, 30.0, s.End.Sub(s.Start).Minutes())
		}
		require.Equal(t, 9, slots[0].Start.Hour())
		require.Equal(t, 17, slots[len(slots)-1].Start.Hour())
	})

	t.Run("дата дальше окна генерации — пустой список", func(t *testing.T) {
		slots := listSlots(t, user, roomID, "2099-01-01")

		require.Empty(t, slots)
	})

	t.Run("комната без расписания — пустой список", func(t *testing.T) {
		emptyRoom := newRoom(t, admin)

		require.Empty(t, listSlots(t, user, emptyRoom, tomorrow()))
	})

	t.Run("без параметра date — 400", func(t *testing.T) {
		r := call(t, http.MethodGet, "/rooms/"+roomID+"/slots/list", user, nil)

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	t.Run("дата в неверном формате — 400", func(t *testing.T) {
		r := call(t, http.MethodGet, "/rooms/"+roomID+"/slots/list?date=20.10.2026", user, nil)

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	t.Run("roomId не UUID — 400", func(t *testing.T) {
		r := call(t, http.MethodGet, "/rooms/abc/slots/list?date="+tomorrow(), user, nil)

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	t.Run("несуществующая комната — 404", func(t *testing.T) {
		r := call(t, http.MethodGet, "/rooms/"+uuid.NewString()+"/slots/list?date="+tomorrow(), user, nil)

		requireError(t, r, http.StatusNotFound, "ROOM_NOT_FOUND")
	})
}

func TestBookingFlow(t *testing.T) {
	admin := dummyToken(t, "admin")
	alice := dummyToken(t, "user")
	bob := newUserToken(t)

	roomID := newRoomWithSchedule(t, admin)
	date := tomorrow()

	slots := listSlots(t, alice, roomID, date)
	require.NotEmpty(t, slots)
	target := slots[0]

	var created booking

	t.Run("alice бронирует слот", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/create", alice, map[string]any{"slotId": target.ID})
		require.Equal(t, http.StatusCreated, r.Status, "тело ответа: %s", r.Body)

		created = decode[bookingResp](t, r).Booking
		require.NotEmpty(t, created.ID)
		require.Equal(t, target.ID, created.SlotID)
		require.Equal(t, "active", created.Status)
		require.Nil(t, created.ConferenceLink, "ссылка не запрашивалась")
	})

	t.Run("занятый слот пропал из списка свободных", func(t *testing.T) {
		after := listSlots(t, alice, roomID, date)

		require.Len(t, after, len(slots)-1)
		for _, s := range after {
			require.NotEqual(t, target.ID, s.ID)
		}
	})

	t.Run("повторная бронь того же слота — 409", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/create", bob, map[string]any{"slotId": target.ID})

		requireError(t, r, http.StatusConflict, "SLOT_ALREADY_BOOKED")
	})

	t.Run("бронь видна в «моих» у alice и не видна у bob", func(t *testing.T) {
		r := call(t, http.MethodGet, "/bookings/my", alice, nil)
		require.Equal(t, http.StatusOK, r.Status)
		require.True(t, containsBooking(decode[bookingsResp](t, r).Bookings, created.ID))

		r = call(t, http.MethodGet, "/bookings/my", bob, nil)
		require.Equal(t, http.StatusOK, r.Status)
		require.False(t, containsBooking(decode[bookingsResp](t, r).Bookings, created.ID))
	})

	t.Run("admin видит бронь в общем списке с пагинацией", func(t *testing.T) {
		r := call(t, http.MethodGet, "/bookings/list?page=1&pageSize=100", admin, nil)
		require.Equal(t, http.StatusOK, r.Status, "тело ответа: %s", r.Body)

		list := decode[bookingsResp](t, r)
		require.Equal(t, 1, list.Pagination.Page)
		require.Equal(t, 100, list.Pagination.PageSize)
		require.GreaterOrEqual(t, list.Pagination.Total, 1)
		require.LessOrEqual(t, len(list.Bookings), 100)
	})

	t.Run("bob не может отменить чужую бронь — 403", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/"+created.ID+"/cancel", bob, nil)

		requireError(t, r, http.StatusForbidden, "FORBIDDEN")
	})

	t.Run("alice отменяет свою бронь", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/"+created.ID+"/cancel", alice, nil)
		require.Equal(t, http.StatusOK, r.Status, "тело ответа: %s", r.Body)

		require.Equal(t, "cancelled", decode[bookingResp](t, r).Booking.Status)
	})

	t.Run("повторная отмена идемпотентна — 200", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/"+created.ID+"/cancel", alice, nil)
		require.Equal(t, http.StatusOK, r.Status, "тело ответа: %s", r.Body)

		require.Equal(t, "cancelled", decode[bookingResp](t, r).Booking.Status)
	})

	t.Run("слот снова свободен и его можно забронировать", func(t *testing.T) {
		after := listSlots(t, bob, roomID, date)
		require.Len(t, after, len(slots))

		r := call(t, http.MethodPost, "/bookings/create", bob, map[string]any{"slotId": target.ID})
		require.Equal(t, http.StatusCreated, r.Status, "тело ответа: %s", r.Body)
	})
}

func TestBookingConferenceLink(t *testing.T) {
	admin := dummyToken(t, "admin")
	user := dummyToken(t, "user")

	roomID := newRoomWithSchedule(t, admin)
	slots := listSlots(t, user, roomID, tomorrow())
	require.NotEmpty(t, slots)

	r := call(t, http.MethodPost, "/bookings/create", user, map[string]any{
		"slotId": slots[0].ID, "createConferenceLink": true,
	})
	require.Equal(t, http.StatusCreated, r.Status, "тело ответа: %s", r.Body)

	b := decode[bookingResp](t, r).Booking
	require.NotNil(t, b.ConferenceLink)
	require.Contains(t, *b.ConferenceLink, "https://meet.example.com/conf-")
}

func TestBookingValidation(t *testing.T) {
	user := dummyToken(t, "user")

	t.Run("slotId не UUID — 400", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/create", user, map[string]any{"slotId": "abc"})

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	t.Run("пустое тело — 400", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/create", user, map[string]any{})

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	t.Run("несуществующий слот — 404", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/create", user, map[string]any{"slotId": uuid.NewString()})

		requireError(t, r, http.StatusNotFound, "SLOT_NOT_FOUND")
	})

	t.Run("отмена несуществующей брони — 404", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/"+uuid.NewString()+"/cancel", user, nil)

		requireError(t, r, http.StatusNotFound, "BOOKING_NOT_FOUND")
	})

	t.Run("отмена с bookingId не UUID — 400", func(t *testing.T) {
		r := call(t, http.MethodPost, "/bookings/abc/cancel", user, nil)

		requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
	})

	admin := dummyToken(t, "admin")

	for name, query := range map[string]string{
		"page=0":        "?page=0",
		"pageSize=0":    "?pageSize=0",
		"pageSize=101":  "?pageSize=101",
		"page не число": "?page=abc",
	} {
		t.Run("список броней: "+name+" — 400", func(t *testing.T) {
			r := call(t, http.MethodGet, "/bookings/list"+query, admin, nil)

			requireError(t, r, http.StatusBadRequest, "INVALID_REQUEST")
		})
	}
}

func containsBooking(list []booking, id string) bool {
	for _, b := range list {
		if b.ID == id {
			return true
		}
	}

	return false
}
