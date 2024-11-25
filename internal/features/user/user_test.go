package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/handlerutils"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type testCase struct {
	name     string
	path     string
	method   string
	payload  RegisterUserRequest
	expected int
}

const (
	registerPath = "/register"
	loginPath    = "/login"
)

var (
	validPayload = RegisterUserRequest{
		FirstName: "Lime",
		LastName:  "Peters",
		Email:     "limepeter@gmail.com",
		Password:  "12345",
	}
	invalidPayloadOne = RegisterUserRequest{
		FirstName: "Lime",
		LastName:  "Peters",
		Email:     "line.com",
		Password:  "12",
	}
	invalidPayloadTwo = RegisterUserRequest{
		FirstName: "",
		LastName:  "Peters",
		Email:     "poster.com",
		Password:  "12345",
	}
)

var testCases = []testCase{
	{
		name:     "should fail to register a new user if payload is invalid",
		path:     registerPath,
		method:   http.MethodPost,
		payload:  invalidPayloadOne,
		expected: http.StatusUnprocessableEntity,
	},
	{
		name:     "should fail to register user cause because firstName is all repeating characters",
		path:     registerPath,
		method:   http.MethodPost,
		payload:  invalidPayloadTwo,
		expected: http.StatusUnprocessableEntity,
	},
	{
		name:     "should successfully register a new user",
		path:     registerPath,
		method:   http.MethodPost,
		payload:  validPayload,
		expected: http.StatusCreated,
	},
	{
		name:     "should fail to register user if user already exists",
		path:     registerPath,
		method:   http.MethodPost,
		payload:  validPayload,
		expected: http.StatusConflict,
	},
}

func TestUserRoutes(t *testing.T) {
	userStore := newMockUserStore()
	userService := NewService(userStore, nil) // todo: no token service
	userHandler := NewHandler(userService)

	// create router
	router := chi.NewRouter()

	router.MethodFunc(
		http.MethodPost,
		registerPath,
		handlerutils.MakeHandler(userHandler.registerUserHandler),
	)
	router.MethodFunc(
		http.MethodPost,
		loginPath,
		handlerutils.MakeHandler(userHandler.loginUserHandler),
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatal(err)
			}

			// create request
			req, err := http.NewRequest(
				tc.method,
				tc.path,
				bytes.NewBuffer(payload))
			if err != nil {
				t.Fatal(err)
			}

			// create response recorder
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code != tc.expected {
				t.Errorf(
					"expected status code %d, got %d",
					tc.expected, rr.Code,
				)
			}

		})
	}

}

type mockStore struct {
	Users map[string]*User
}

func newMockUserStore() *mockStore {
	return &mockStore{
		Users: make(map[string]*User),
	}
}

func (m *mockStore) create(ctx context.Context, user *User) error {
	user.UserID = uuid.New()
	m.Users[user.Email] = user
	return nil
}

func (m *mockStore) findByEmail(ctx context.Context, email string) (*User, error) {
	user, exists := m.Users[email]

	if !exists {
		user = new(User) // initialize to zero values
		return user, nil
	}

	return user, nil
}

func (m *mockStore) findByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	return nil, nil
}

func (m *mockStore) createSession(ctx context.Context, session *Session) error {
	panic("unimplemented")
}

func (m *mockStore) findSessionByUserIDAndUserAgent(ctx context.Context, userID uuid.UUID, UserAgent string) (*Session, error) {
	panic("unimplemented")
}

func (m *mockStore) deleteSessionByID(ctx context.Context, sessionID uuid.UUID) error {
	panic("unimplemented")
}

func (m *mockStore) findSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	panic("unimplemented")
}
