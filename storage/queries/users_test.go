package queries_test

import (
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"testing"
	"time"
)

func TestIsUsernameExists(t *testing.T) {
	us := queries.NewUserStore(database.GetSession())

	if us.IsUsernameExists(t.Context(), "InvalidUsername") {
		t.Errorf("expected false, but it return true for existing of invalid username")
	}
	if !us.IsUsernameExists(t.Context(), "adminUser") {
		t.Errorf("expected true, but it return false for existing of valid username")
	}
}

func TestCreateGetDetailsUser(t *testing.T) {
	session := database.GetSession()
	us := queries.NewUserStore(session)

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

	userDetails, err := us.GetDetails(t.Context(), user.Username)
	if err != nil {
		t.Errorf("failed to get user details, %s", err)
	}
	const isUserProfileExistsQuery = `SELECT EXISTS(SELECT 1 FROM user_profile WHERE user_id = $1)`
	var isExists bool
	if err := session.QueryRow(t.Context(), isUserProfileExistsQuery, userDetails.ID).Scan(&isExists); err != nil || !isExists {
		t.Errorf("failed to query existing of user_profile on user creation, %s", err)
	}
}

func TestSetEmailAndSetPhoneNumber(t *testing.T) {
	const getIDFromUsernameQuery = `SELECT id FROM users WHERE username = $1`
	session := database.GetSession()
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

	if err := us.VerifyEmail(t.Context(), id, true); err != nil {
		t.Errorf(`failed to verify email, %s`, err)
	}

	phoneNumber := "123393"
	if err := us.SetPhoneNumber(t.Context(), id, &queries.PhoneNumberRequest{phoneNumber}); err != nil {
		t.Errorf("failed to set phone number, %s", err)
	}

	if err := us.VerifyPhoneNumber(t.Context(), id, true); err != nil {
		t.Errorf(`failed to verify phone number, %s`, err)
	}
}

func TestGetUser(t *testing.T) {
	us := queries.NewUserStore(database.GetSession())

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
	us := queries.NewUserStore(database.GetSession())

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
	us := queries.NewUserStore(database.GetSession())

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

	if err := us.Delete(t.Context(), gUser.ID, user.Username); err != nil {
		t.Errorf(`failed to delete user "%s, %s`, user.Username, err)
	}
	if us.IsUsernameExists(t.Context(), user.Username) {
		t.Errorf(`expected to user be deleted, but got existing of user "%s"`, user.Username)
	}
}

func TestUserProfileStore(t *testing.T) {
	ctx := t.Context()

	session := database.GetSession()
	us := queries.NewUserStore(session)
	ups := queries.NewUserProfileStore(session)

	username := "adminUser"
	userDetails, err := us.GetDetails(ctx, username)
	if err != nil {
		t.Errorf("failed to get user details, %s", err)
	}
	upr := &queries.UserProfileRequest{
		Birthday: time.Now().Format(time.DateOnly),
		Bio:      "some descriptions",
	}
	if err = ups.Upsert(ctx, userDetails.ID, upr); err != nil {
		t.Errorf("failed to upsert birthday and bio on user_profile, %s", err)
	}
	imgPath := "/some/path/to/img.img"
	if err := ups.SetImgPath(ctx, userDetails.ID, imgPath); err != nil {
		t.Errorf("failed to set img path, %s", err)
	}

	if gotImgPatherr, err := ups.GetImgPath(ctx, userDetails.ID); err != nil || gotImgPatherr != imgPath {
		t.Errorf(`expected imgPath="%s", but got "%s", %s`, imgPath, gotImgPatherr, err)
	}
	userDetails, err = us.GetDetails(ctx, username)
	if err != nil {
		t.Errorf("failed to get user Details after updating user profile fields, %s", err)
	}
	if userDetails.Birthday != upr.Birthday || userDetails.Bio != upr.Bio {
		t.Errorf(`unexpected output birthday="%s", bio="%s", phoneNumber="%s"`, userDetails.Birthday, userDetails.Bio, userDetails.PhoneNumber)
	}
}
