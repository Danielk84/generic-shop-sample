package queries_test

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"testing"
)

func TestIsUsernameExists(t *testing.T) {
	session := db.Session()
	um := queries.NewUserManager(session)

	if um.IsUsernameExists(t.Context(), "InvalidUsername") {
		t.Errorf("expected false, but it return true for existing of invalid username")
	}
	if !um.IsUsernameExists(t.Context(), "adminUser") {
		t.Errorf("expected true, but it return false for existing of valid username")
	}
}

func TestCreateUser(t *testing.T) {
	session := db.Session()
	um := queries.NewUserManager(session)

	user := &queries.User{
		Username:       "validUser",
		PermissionType: queries.Customer,
		IsActive:       true,
	}
	if err := um.Create(t.Context(), user); err != nil {
		t.Errorf("expected to creating validUser, but got: %s", err)
	}
	if !um.IsUsernameExists(t.Context(), user.Username) {
		t.Errorf("expected to existing of created user, but got false")
	}

	if err := um.Create(t.Context(), user); err == nil {
		t.Errorf("error duplicate user created")
	}
}

const getIDFromUsernameQuery = `SELECT id FROM users WHERE username = $1`
const getEmailQuery = `SELECT email FROM users WHERE id = $1`

func TestSetEmail(t *testing.T) {
	session := db.Session()
	um := queries.NewUserManager(session)

	var id int32
	if err := session.QueryRow(t.Context(), getIDFromUsernameQuery, "customerUser").Scan(&id); err != nil {
		t.Errorf("failed to query id from username, %s", err)
	}

	email := "someaddr@email.blab"
	user := &queries.User{
		ID:    id,
		Email: &email,
	}
	if err := um.SetEmail(t.Context(), user); err != nil {
		t.Errorf("bad SetEmail operation, %s", err)
	}

	if err := session.QueryRow(t.Context(), getEmailQuery, user.ID).Scan(&email); err != nil {
		t.Errorf("failed to query email address with id, %s", err)
	}

	if email != *user.Email {
		t.Errorf(`expected email="%s", but got "%s"`, *user.Email, email)
	}
}

func TestReadUser(t *testing.T) {
	session := db.Session()
	um := queries.NewUserManager(session)

	passwordHash := "simpleHash"
	user := queries.User{
		Username:       "NewSimpleUser",
		PermissionType: queries.Customer,
		PasswordHash:   &passwordHash,
		IsActive:       true,
	}
	if err := um.Create(t.Context(), &user); err != nil {
		t.Errorf("failed to creating new user, %s", err)
	}

	rUser, err := um.Read(t.Context(), user.Username)
	if err != nil {
		t.Errorf(`failed to reading user "%s", %s`, user.Username, err)
	}

	if rUser.PasswordHash != user.PasswordHash && !rUser.IsActive && rUser.PermissionType != queries.Customer {
		t.Errorf("failed to math fields")
	}
}

func TestUpdateUser(t *testing.T) {
	um := queries.NewUserManager(db.Session())

	username := "blockUser"
	user, err := um.Read(t.Context(), username)
	if err != nil {
		t.Errorf(`failed to read user "%s", %s`, username, err)
	}

	user.IsActive = true
	user.PermissionType = queries.BlockUser
	if err := um.Update(t.Context(), user); err != nil {
		t.Errorf(`failed to update user "%s", %s`, username, err)
	}

	newUser, err := um.Read(t.Context(), username)
	if err != nil {
		t.Errorf(`failed to re-read user "%s", %s`, username, err)
	}

	if !newUser.IsActive && newUser.PermissionType != queries.BlockUser {
		t.Errorf("bad user update method")
	}
}

func TestDeleteUser(t *testing.T) {
	um := queries.NewUserManager(db.Session())

	user := &queries.User{
		Username:       "deleteUser",
		PermissionType: queries.BlockUser,
		IsActive:       false,
	}
	if err := um.Create(t.Context(), user); err != nil {
		t.Errorf(`failed to create user "%s", %s`, user.Username, err)
	}

	rUser, err := um.Read(t.Context(), user.Username)
	if err != nil {
		t.Errorf(`failed to read user "%s", %s`, user.Username, err)
	}

	if err := um.Delete(t.Context(), rUser.ID); err != nil {
		t.Errorf(`failed to delete user "%s, %s`, user.Username, err)
	}
	if um.IsUsernameExists(t.Context(), user.Username) {
		t.Errorf(`expected to user be deleted, but got existing of user "%s"`, user.Username)
	}
}
