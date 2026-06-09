package cmd

import (
	"fmt"
	"generic-shop-sample/db/database"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	"log"

	"github.com/spf13/cobra"
)

var (
	username string
	password string
)

var newAdminCmd = &cobra.Command{
	Use:   "new-admin",
	Short: "create new admin",
	Run:   newAdmin,
}

func init() {
	newAdminCmd.Flags().StringVarP(&username, "username", "u", "", "admin username")
	newAdminCmd.Flags().StringVarP(&password, "password", "p", "", "admin password")

	rootCmd.AddCommand(newAdminCmd)
}

func newAdmin(cmd *cobra.Command, args []string) {
	us := queries.NewUserStore(database.GetSession())
	user := queries.CreateUserRequest{
		LoginRequest: queries.LoginRequest{
			Username: username,
			Password: password,
		},
		UserPermissionRequest: queries.UserPermissionRequest{
			PermissionType: queries.Admin,
			IsActive:       true,
		},
	}
	var err error
	validate := internal.GetValidator()
	if err = validate.Struct(user); err != nil {
		if username == "" || password == "" {
			log.Fatalln("username and password required.")
			return
		}
		log.Fatalf("failed to validate input, %s", err)
		return
	}
	if user.Password, err = auth.PasswordHash(password); err != nil {
		log.Fatalf("failed to hash password, %s", err)
		return
	}
	if err = us.Create(cmd.Context(), &user); err != nil {
		log.Fatalf("failed to create admin, %s", err)
		return
	}
	fmt.Println("admin created.")
}
