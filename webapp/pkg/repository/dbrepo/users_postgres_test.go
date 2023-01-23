//go:build integration

package dbrepo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"webapp/pkg/data"
	"webapp/pkg/repository"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "postgres"
	dbName   = "users_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var (
	resource *dockertest.Resource
	pool     *dockertest.Pool
	testDB   *sql.DB
	testRepo repository.DatabaseRepo
)

func TestMain(m *testing.M) {
	// connect to docker; fail if docker not running
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker; is it running? %s", err)
	}

	pool = p

	// set up our docker options, specifying the image and so forth
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.5",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	// get a resource (docker image)
	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource: %s", err)
	}

	// start the image and wait until it's ready
	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			log.Println("Error:", err)
			return err
		}
		return testDB.Ping()
	}); err != nil {
		log.Printf("could not connect to database: %s", err)
		cleanUp()
	}

	// populate the database with empty tables
	err = createTables()
	if err != nil {
		pool.Purge(resource)
		log.Fatalf("unable to create tables: %s", err)
	}

	testRepo = &PostgresDBRepo{DB: testDB}

	// run tests
	code := m.Run()

	// clean up
	cleanUp()

	os.Exit(code)
}

func cleanUp() {
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("unable to purge resource: %s", err)
	}
}
func createTables() error {

	tableSql, err := os.ReadFile("./testdata/users.sql")
	if err != nil {
		return err
	}

	_, err = testDB.Exec(string(tableSql))

	return err
}

func Test_ping_db(t *testing.T) {
	err := testDB.Ping()
	if err != nil {
		t.Error("unable to ping database")
	}
}

func Test_PostgresDBRepo_InsertUser(t *testing.T) {
	testUser := data.User{
		FirstName: "Admin",
		LastName:  "User",
		Password:  "password",
		Email:     "admin@localhost.com",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertUser(testUser)
	if err != nil {
		t.Errorf("unable to insert user: %s", err)
	}

	if id != 1 {
		t.Errorf("insertUser returned wrong id; expect 1 got %d", id)
	}
}

func Test_PostgresDBRepo_AllUsers(t *testing.T) {

	users, err := testRepo.AllUsers()
	if err != nil {
		t.Error(err)
	}

	c := len(users)

	if c != 1 {
		t.Errorf("expect count of user to be 1; got %d", c)
	}

	testUser := data.User{
		FirstName: "Admin2",
		LastName:  "User2",
		Password:  "password2",
		Email:     "admin2@localhost.com",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, _ = testRepo.InsertUser(testUser)

	users, err = testRepo.AllUsers()
	if err != nil {
		t.Error(err)
	}
	c = len(users)

	if c != 2 {
		t.Errorf("expect count of users to be 2; got %d", c)
	}

	if users[1].FirstName != testUser.FirstName {
		t.Errorf("expect %s; got %s", testUser.FirstName, users[1].FirstName)
	}
}

func Test_PostgresDBRepo_GetUser(t *testing.T) {

	user, err := testRepo.GetUser(1)
	if err != nil {
		t.Error(err)
	}

	if user == nil {
		t.Fatalf("expect GetUser(1) to return a user object")
	}

	if user.ID != 1 {
		t.Error("expect GetUser(1) to return a user object with ID of 1")
	}

	if user.Email != "admin@localhost.com" {
		t.Error("wrong email returned by get user; we expect admin@localhost.com")
	}

}

func Test_PostgresDBRepo_GetUserByEmail(t *testing.T) {

	testCases := []struct {
		user      data.User
		expectNil bool
	}{
		{data.User{Email: "admin@localhost.com"}, false},
		{data.User{Email: "admin2@localhost.com"}, false},
		{data.User{Email: "invalid-email@localhost.com"}, true},
	}

	for _, tt := range testCases {
		t.Run(tt.user.Email, func(t *testing.T) {
			u, err := testRepo.GetUserByEmail(tt.user.Email)
			if tt.expectNil {

				if err == nil {
					t.Errorf("expect GetUserByEmail(%s) to return nil", tt.user.Email)
				}

				if u != nil {
					t.Errorf("expect GetUserByEmail(%s) to return nil", tt.user.Email)
				}

			} else {

				if err != nil {
					t.Fatal(err)
				}

				if u == nil {
					t.Fatal(err)
				}

				if tt.user.Email != u.Email {
					t.Errorf("GetUserByEmail(%s) return the wrong user; expect %s; got %s", tt.user.Email, tt.user.Email, u.Email)
				}

			}

		})
	}

}

func Test_PostgresDBRepo_UpdateUser(t *testing.T) {

	user, err := testRepo.GetUser(1)
	if err != nil {
		t.Fatalf("GetUser(1) returned an error: %s", err)
	}

	user.FirstName = "AJ"
	user.LastName = "AB"
	user.Email = "aj@admin.com"

	err = testRepo.UpdateUser(*user)

	if err != nil {
		t.Errorf("UpdateUser() returned an error: %s", err)
	}

	newData, _ := testRepo.GetUser(1)
	if newData.FirstName != user.FirstName {
		t.Errorf("failed to update user;")
	}

	if newData.LastName != user.LastName {
		t.Errorf("failed to update user;")
	}

	if newData.Email != user.Email {
		t.Errorf("failed to update user;")
	}
}

func Test_PostgresDBRepo_DeleteUser(t *testing.T) {

	err := testRepo.DeleteUser(2)
	if err != nil {
		t.Errorf("DeleteUser(2) returned an error when it shouldn`t")
	}

	_, err = testRepo.GetUser(2)
	if err == nil {
		t.Errorf("user with ID 2 was retrived even though it was deleted")
	}
}

func Test_PostgresDBRepo_ResetPassword(t *testing.T) {

	err := testRepo.ResetPassword(1, "password")
	if err != nil {
		t.Error(err)
	}

	u, _ := testRepo.GetUser(1)

	ok, err := u.PasswordMatches("password")
	if err != nil {
		t.Error(err)
	}

	if !ok {
		t.Error("password mismatch!")
	}
}

func Test_PostgresDBRepo_InsertUserImage(t *testing.T) {

	img := data.UserImage{
		UserID:    1,
		FileName:  "avatar.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	id, err := testRepo.InsertUserImage(img)
	if err != nil {
		t.Error(err)
	}

	if id != 1 {
		t.Errorf("expected the first id to be 1; got %d", id)
	}

	img.UserID = 100 // invalid user id

	id, err = testRepo.InsertUserImage(img)
	if err == nil {
		t.Errorf("expect InsertUserImage() to return an error for invalid user ID")
	}

	if id != 0 {
		t.Errorf("expect InsertUserImage() to return 0 for an id of invalid user; got %d", id)
	}
}
