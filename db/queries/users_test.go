package queries_test

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"testing"
)

func TestIsUsernameExists(t *testing.T) {
	us := queries.NewUserStore(db.NewSession())

	if us.IsUsernameExists(t.Context(), "InvalidUsername") {
		t.Errorf("expected false, but it return true for existing of invalid username")
	}
	if !us.IsUsernameExists(t.Context(), "adminUser") {
		t.Errorf("expected true, but it return false for existing of valid username")
	}
}

func TestCreateUser(t *testing.T) {
	us := queries.NewUserStore(db.NewSession())

	user := &queries.CreateUserRequest{
		queries.LoginRequest{"validUser", "secure_password"},
		queries.UserPermissionRequest{queries.Customer, true},
	}
	if err := us.Create(t.Context(), user); err != nil {
		t.Errorf("expected to creating valid user, but got: %s", err)
	}
	if !us.IsUsernameExists(t.Context(), user.Username) {
		t.Errorf("expected to existing of created user, but got false")
	}

	if err := us.Create(t.Context(), user); err == nil {
		t.Errorf("error duplicate user created")
	}
}

func TestSetEmail(t *testing.T) {
	const getIDFromUsernameQuery = `SELECT id FROM users WHERE username = $1`
	session := db.NewSession()
	us := queries.NewUserStore(session)
	var id int32
	if err := session.QueryRow(t.Context(), getIDFromUsernameQuery, "customerUser").Scan(&id); err != nil {
		t.Errorf("failed to query id from username, %s", err)
	}

	email := &queries.EmailAddrRequest{"someaddr@email.blab"}
	if err := us.SetEmail(t.Context(), id, email); err != nil {
		t.Errorf("bad SetEmail operation, %s", err)
	}
	var newEmail string
	const getEmailQuery = `SELECT email FROM users WHERE id = $1`
	if err := session.QueryRow(t.Context(), getEmailQuery, id).Scan(&newEmail); err != nil {
		t.Errorf("failed to query email address with id, %s", err)
	}

	if newEmail != email.Email {
		t.Errorf(`expected email="%s", but got "%s"`, email.Email, newEmail)
	}
}

func TestGetUser(t *testing.T) {
	us := queries.NewUserStore(db.NewSession())

	user := queries.CreateUserRequest{
		queries.LoginRequest{"NewSimpleUser", "secure password"},
		queries.UserPermissionRequest{queries.Customer, true},
	}
	if err := us.Create(t.Context(), &user); err != nil {
		t.Errorf("failed to creating new user, %s", err)
	}

	gUser, err := us.Get(t.Context(), user.Username)
	if err != nil {
		t.Errorf(`failed to get user "%s", %s`, user.Username, err)
	}

	if gUser.Password != user.Password && !gUser.IsActive && gUser.PermissionType != queries.Customer {
		t.Errorf("failed to match fields")
	}
}

func TestUpdateUserPermission(t *testing.T) {
	us := queries.NewUserStore(db.NewSession())

	username := "blockUser"
	user, err := us.Get(t.Context(), username)
	if err != nil {
		t.Errorf(`failed to get user "%s", %s`, username, err)
	}

	if err := us.UpdatePermission(t.Context(), user.ID, &queries.UserPermissionRequest{queries.BlockUser, true}); err != nil {
		t.Errorf(`failed to update user "%s", %s`, username, err)
	}

	newUser, err := us.Get(t.Context(), username)
	if err != nil {
		t.Errorf(`failed to get user "%s", %s`, username, err)
	}

	if !newUser.IsActive && newUser.PermissionType != queries.BlockUser {
		t.Errorf("bad user update method")
	}
}

func TestDeleteUser(t *testing.T) {
	us := queries.NewUserStore(db.NewSession())

	user := &queries.CreateUserRequest{
		queries.LoginRequest{"deleteUser", "secure password"},
		queries.UserPermissionRequest{queries.BlockUser, false},
	}
	if err := us.Create(t.Context(), user); err != nil {
		t.Errorf(`failed to create user "%s", %s`, user.Username, err)
	}

	gUser, err := us.Get(t.Context(), user.Username)
	if err != nil {
		t.Errorf(`failed to get user "%s", %s`, user.Username, err)
	}

	if err := us.Delete(t.Context(), gUser.ID); err != nil {
		t.Errorf(`failed to delete user "%s, %s`, user.Username, err)
	}
	if us.IsUsernameExists(t.Context(), user.Username) {
		t.Errorf(`expected to user be deleted, but got existing of user "%s"`, user.Username)
	}
}
