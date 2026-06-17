package queries_test

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"testing"
	"time"
)

func TestUserStore(t *testing.T) {
	ctx := t.Context()
	session := database.GetSession()
	log := logger.GetLogger()
	store := queries.NewUserStore(session, log)
	s := testUserStore{ctx, session, store}

	s.isUsernameExists(t)
	s.createGetDetailsUser(t)
	s.setEmailAndSetPhoneNumber(t)
	s.getUser(t)
	s.updateUserPermission(t)
	s.deleteUser(t)
}

type testUserStore struct {
	ctx     context.Context
	session database.Session
	store   queries.UserStore
}

func (s *testUserStore) isUsernameExists(t *testing.T) {
	if s.store.IsUsernameExists(s.ctx, "InvalidUsername") {
		t.Errorf("expected false, but it return true for existing of invalid username")
	}
	if !s.store.IsUsernameExists(s.ctx, "adminUser") {
		t.Errorf("expected true, but it return false for existing of valid username")
	}
}

func (s *testUserStore) createGetDetailsUser(t *testing.T) {
	user := &queries.CreateUserRequest{
		queries.LoginRequest{"validUser", "secure_password"},
		queries.UserPermissionRequest{queries.Customer, true},
	}
	if err := s.store.Create(s.ctx, user); err != nil {
		t.Errorf("expected to creating valid user, but got: %s", err)
	}
	if !s.store.IsUsernameExists(s.ctx, user.Username) {
		t.Errorf("expected to existing of created user, but got false")
	}

	if err := s.store.Create(s.ctx, user); err == nil {
		t.Errorf("error duplicate user created")
	}

	userDetails, err := s.store.GetDetails(s.ctx, user.Username)
	if err != nil {
		t.Errorf("failed to get user details, %s", err)
	}
	const isUserProfileExistsQuery = `SELECT EXISTS(SELECT 1 FROM user_profile WHERE user_id = $1)`
	var isExists bool
	if err := s.session.QueryRow(s.ctx, isUserProfileExistsQuery, userDetails.ID).Scan(&isExists); err != nil || !isExists {
		t.Errorf("failed to query existing of user_profile on user creation, %s", err)
	}
}

func (s *testUserStore) setEmailAndSetPhoneNumber(t *testing.T) {
	const getIDFromUsernameQuery = `SELECT id FROM users WHERE username = $1`
	var id int32
	if err := s.session.QueryRow(s.ctx, getIDFromUsernameQuery, "customerUser").Scan(&id); err != nil {
		t.Errorf("failed to query id from username, %s", err)
	}

	email := &queries.EmailAddrRequest{"someaddr@email.blab"}
	if err := s.store.SetEmail(s.ctx, id, email); err != nil {
		t.Errorf("bad SetEmail operation, %s", err)
	}
	var newEmail string
	const getEmailQuery = `SELECT email FROM users WHERE id = $1`
	if err := s.session.QueryRow(s.ctx, getEmailQuery, id).Scan(&newEmail); err != nil {
		t.Errorf("failed to query email address with id, %s", err)
	}

	if newEmail != email.Email {
		t.Errorf(`expected email="%s", but got "%s"`, email.Email, newEmail)
	}

	if err := s.store.VerifyEmail(s.ctx, id, true); err != nil {
		t.Errorf(`failed to verify email, %s`, err)
	}

	phoneNumber := "123393"
	if err := s.store.SetPhoneNumber(s.ctx, id, &queries.PhoneNumberRequest{phoneNumber}); err != nil {
		t.Errorf("failed to set phone number, %s", err)
	}

	if err := s.store.VerifyPhoneNumber(s.ctx, id, true); err != nil {
		t.Errorf(`failed to verify phone number, %s`, err)
	}
}

func (s *testUserStore) getUser(t *testing.T) {
	user := queries.CreateUserRequest{
		queries.LoginRequest{"NewSimpleUser", "secure password"},
		queries.UserPermissionRequest{queries.Customer, true},
	}
	if err := s.store.Create(s.ctx, &user); err != nil {
		t.Errorf("failed to creating new user, %s", err)
	}

	gUser, err := s.store.Get(s.ctx, user.Username)
	if err != nil {
		t.Errorf(`failed to get user "%s", %s`, user.Username, err)
	}

	if gUser.Password != user.Password && !gUser.IsActive && gUser.PermissionType != queries.Customer {
		t.Errorf("failed to match fields")
	}
}

func (s *testUserStore) updateUserPermission(t *testing.T) {
	username := "blockUser"
	user, err := s.store.Get(s.ctx, username)
	if err != nil {
		t.Errorf(`failed to get user "%s", %s`, username, err)
	}

	if err := s.store.UpdatePermission(s.ctx, user.ID, &queries.UserPermissionRequest{queries.BlockUser, true}); err != nil {
		t.Errorf(`failed to update user "%s", %s`, username, err)
	}

	newUser, err := s.store.Get(s.ctx, username)
	if err != nil {
		t.Errorf(`failed to get user "%s", %s`, username, err)
	}

	if !newUser.IsActive && newUser.PermissionType != queries.BlockUser {
		t.Errorf("bad user update method")
	}
}

func (s *testUserStore) deleteUser(t *testing.T) {
	user := &queries.CreateUserRequest{
		queries.LoginRequest{"deleteUser", "secure password"},
		queries.UserPermissionRequest{queries.BlockUser, false},
	}
	if err := s.store.Create(s.ctx, user); err != nil {
		t.Errorf(`failed to create user "%s", %s`, user.Username, err)
	}

	gUser, err := s.store.Get(s.ctx, user.Username)
	if err != nil {
		t.Errorf(`failed to get user "%s", %s`, user.Username, err)
	}

	if err := s.store.Delete(s.ctx, gUser.ID, user.Username); err != nil {
		t.Errorf(`failed to delete user "%s, %s`, user.Username, err)
	}
	if s.store.IsUsernameExists(s.ctx, user.Username) {
		t.Errorf(`expected to user be deleted, but got existing of user "%s"`, user.Username)
	}
}

func TestUserProfileStore(t *testing.T) {
	ctx := t.Context()
	session := database.GetSession()
	log := logger.GetLogger()
	userStore := queries.NewUserStore(session, log)
	store := queries.NewUserProfileStore(session, log)

	username := "adminUser"
	userDetails, err := userStore.GetDetails(ctx, username)
	if err != nil {
		t.Errorf("failed to get user details, %s", err)
	}

	s := testUserProfileStore{ctx, session, userStore, store, userDetails}

	s.upsert(t)
	s.setGetImgPath(t)
	s.checkNewUserDetails(t)
}

type testUserProfileStore struct {
	ctx         context.Context
	session     database.Session
	userStore   queries.UserStore
	store       queries.UserProfileStore
	userDetails queries.UserDetailsResponse
}

func (s *testUserProfileStore) upsert(t *testing.T) {
	upr := &queries.UserProfileRequest{
		Birthday: time.Now().Format(time.DateOnly),
		Bio:      "some descriptions",
	}
	if err := s.store.Upsert(s.ctx, s.userDetails.ID, upr); err != nil {
		t.Errorf("failed to upsert birthday and bio on user_profile, %s", err)
	}
	s.userDetails.Birthday = upr.Birthday
	s.userDetails.Bio = upr.Bio
}

func (s *testUserProfileStore) setGetImgPath(t *testing.T) {
	imgPath := "/some/path/to/img.img"
	if err := s.store.SetImgPath(s.ctx, s.userDetails.ID, imgPath); err != nil {
		t.Errorf("failed to set img path, %s", err)
	}

	if gotImgPatherr, err := s.store.GetImgPath(s.ctx, s.userDetails.ID); err != nil || gotImgPatherr != imgPath {
		t.Errorf(`expected imgPath="%s", but got "%s", %s`, imgPath, gotImgPatherr, err)
	}
}

func (s *testUserProfileStore) checkNewUserDetails(t *testing.T) {
	userDetails, err := s.userStore.GetDetails(s.ctx, s.userDetails.Username)
	if err != nil {
		t.Errorf("failed to get user Details after updating user profile fields, %s", err)
	}
	if userDetails.Birthday != s.userDetails.Birthday || userDetails.Bio != s.userDetails.Bio {
		t.Errorf(`unexpected output birthday="%s", bio="%s", phoneNumber="%s"`, userDetails.Birthday, userDetails.Bio, userDetails.PhoneNumber)
	}
}
