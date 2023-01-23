package dbrepo

import (
	"database/sql"
	"errors"
	"time"
	"webapp/pkg/data"
)

type MockDBRepo struct {
}

func mockUser() data.User {
	return data.User{
		ID:        1,
		Email:     "admin@example.com",
		Password:  "secret",
		FirstName: "Admin",
		LastName:  "User",
	}
}

func (m *MockDBRepo) Connection() *sql.DB {
	return nil
}

// AllUsers returns all users as a slice of *data.User
func (m *MockDBRepo) AllUsers() ([]*data.User, error) {

	var users []*data.User

	return users, nil
}

// GetUser returns one user by id
func (m *MockDBRepo) GetUser(id int) (*data.User, error) {
	if id == 1 {
		u := mockUser()
		return &u, nil
	}

	return nil, errors.New("user not found")
}

// GetUserByEmail returns one user by email address
func (m *MockDBRepo) GetUserByEmail(email string) (*data.User, error) {

	if email == "admin@example.com" {
		return &data.User{
			ID:        1,
			FirstName: "Admin",
			LastName:  "User",
			Email:     "admin@example.com",
			Password:  "$2a$14$ajq8Q7fbtFRQvXpdCq7Jcuy.Rx1h/L4J60Otx.gyNLbAYctGMJ9tK",
			IsAdmin:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	if email == "invalid@sql.com" {
		return nil, errors.New("invalid response from sql")
	}
	u := mockUser()

	return &u, nil
}

// UpdateUser updates one user in the database
func (m *MockDBRepo) UpdateUser(u data.User) error {
	if u.ID == 1 {
		return nil
	}
	return errors.New("update failed")
}

// DeleteUser deletes one user from the database, by id
func (m *MockDBRepo) DeleteUser(id int) error {
	if id != 1 {
		return errors.New("user not found")
	}
	return nil
}

// InsertUser inserts a new user into the database, and returns the ID of the newly inserted row
func (m *MockDBRepo) InsertUser(user data.User) (int, error) {
	if user.Email == "neo@example.com" {
		return 1, nil
	}

	return 0, errors.New("unable to insert user")
}

// ResetPassword is the method we will use to change a user's password.
func (m *MockDBRepo) ResetPassword(id int, password string) error {
	return nil
}

// InsertUserImage inserts a user profile image into the database.
func (m *MockDBRepo) InsertUserImage(i data.UserImage) (int, error) {
	return 1, nil
}
