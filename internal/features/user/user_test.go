package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type testCase struct {
	name     string
	path     string
	method   string
	payload  RegisterUserRequestDTO
	expected int
}

const (
	registerPath = "/register"
	loginPath    = "/login"
)

var (
	validPayload = RegisterUserRequestDTO{
		FirstName: "Lime",
		LastName:  "Peters",
		Email:     "lime@gmail.com",
		Password:  "12345",
	}
	invalidPayload = RegisterUserRequestDTO{
		FirstName: "Lime",
		LastName:  "Peters",
		Email:     "line.com",
		Password:  "12",
	}
)

var testCases = []testCase{
	{
		name:     "should fail to register a new user if payload is invalid",
		path:     registerPath,
		method:   http.MethodPost,
		payload:  invalidPayload,
		expected: http.StatusBadRequest,
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
	userService := NewService(userStore)
	userHandler := NewHandler(userService)

	// create router
	router := chi.NewRouter()

	router.MethodFunc(
		http.MethodPost,
		registerPath,
		userHandler.registerUserHandler,
	)
	router.MethodFunc(
		http.MethodPost,
		loginPath,
		userHandler.loginUserHandler,
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

type mockUserStore struct {
	User map[string]*User
}

func newMockUserStore() *mockUserStore {
	return &mockUserStore{
		User: make(map[string]*User),
	}
}

func (m *mockUserStore) create(ctx context.Context, user *User) error {
	m.User = make(map[string]*User)
	m.User[user.Email] = user
	return nil
}

func (m *mockUserStore) findByEmail(ctx context.Context, email string) (*User, error) {
	user, exists := m.User[email]

	// since no user was found and it is not an error from the database
	// we can return nil and nil
	//but if an actual error occurs when accessing the database we can return an error
	if !exists {
		return nil, nil
	}

	return user, nil
}

func (m *mockUserStore) findByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	return nil, nil
}
