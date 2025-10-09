package api

import (
	"fmt"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal/auth"
	md "generic-shop-sample/middlewares"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func UsersRouter(router *gin.RouterGroup) {
	uh := usersHandler{
		queries.NewUserStore(db.NewSession()),
		db.NewCache(db.UsersCache),
	}

	router.Use(md.AuthMiddleware())
	router.GET("/", uh.list)
	router.GET("/:username", uh.get)
	router.POST("/", uh.createUserByAdmin)
	router.PUT("/:id", uh.updateUserPermission)
	router.PUT("/set-email", uh.setEmail)
	router.DELETE("/", uh.delete)
}

type usersHandler struct {
	us    queries.UserStore
	cache db.CacheClient
}

func (uh *usersHandler) createUserByAdmin(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	var json queries.CreateUserRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "invalid user data")
		return
	}

	if uh.us.IsUsernameExists(c.Request.Context(), json.Username) {
		c.JSON(http.StatusConflict, gin.H{"msg": "username already exists"})
		return
	}
	var err error
	json.Password, err = auth.PasswordHash(json.Password)
	if err != nil {
		BadRequest(c, "invalid password string")
		return
	}
	if err = uh.us.Create(c.Request.Context(), &json); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (uh *usersHandler) list(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}
	users, err := uh.us.List(c.Request.Context(), defaultPagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, users)
}

func (uh *usersHandler) get(c *gin.Context) {
	username := c.Param("username")
	user, err := uh.us.GetDetails(c.Request.Context(), username)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, user)
}

func (uh *usersHandler) updateUserPermission(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid id")
		return
	}
	var json queries.UserPermissionRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		slog.Debug(err.Error())
		BadRequest(c, "invalid permission_id or is_active")
		return
	}

	if err := uh.us.UpdatePermission(c.Request.Context(), int32(id), &json); err != nil {
		NotFound(c, "")
		return
	}
	_ = uh.cache.Del(c.Request.Context(), fmt.Sprintf("login:%d", claims.ID))
	Accepted(c, "")
}

func (uh *usersHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if err := uh.us.Delete(c.Request.Context(), claims.ID); err != nil {
		NotFound(c, "")
		return
	}

	c.Status(http.StatusNoContent)
}

func (uh *usersHandler) setEmail(c *gin.Context) {
	var json queries.EmailAddrRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		slog.Debug(err.Error())
		BadRequest(c, "invalid email address")
		return
	}
	claims := md.GetUserClaims(c)
	if err := uh.us.SetEmail(c.Request.Context(), claims.ID, &json); err != nil {
		BadRequest(c, "email already exists")
		return
	}
	Accepted(c, "")
}

func UserProfileRouter(router *gin.RouterGroup) {
	uph := userProfileHandler{queries.NewUserProfileStore(db.NewSession())}

	router.Use(md.AuthMiddleware())
	router.POST("/", uph.upsert)
	router.POST("/upload", uph.uploadProfileImg)
	router.PUT("/set-phone-number", uph.setPhoneNumber)
	router.DELETE("/", uph.deleteImgPath)
}

type userProfileHandler struct {
	ups queries.UserProfileStore
}

func (uph *userProfileHandler) upsert(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json queries.UserProfileRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		slog.Debug(err.Error())
		BadRequest(c, "")
		return
	}
	if err := uph.ups.Upsert(c.Request.Context(), claims.ID, &json); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (uph *userProfileHandler) uploadProfileImg(c *gin.Context) {
	claims := md.GetUserClaims(c)
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "")
		return
	}
	dst := ""
	if fpath, err := uph.ups.GetImgPath(c.Request.Context(), claims.ID); err == nil {
		dst = fpath
	}
	resultPath, err := UploadFile(file, claims, "user-profile", dst)
	if err != nil {
		slog.Debug(err.Error())
		BadRequest(c, "failed to process file")
		return
	}
	slog.Debug(resultPath)
	if resultPath != dst {
		if err := uph.ups.SetImgPath(c.Request.Context(), claims.ID, resultPath); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "failed to save file"})
			return
		}
	}
	Accepted(c, "")
}

func (upr *userProfileHandler) deleteImgPath(c *gin.Context) {
	claims := md.GetUserClaims(c)
	imgPath, err := upr.ups.DeleteImgPath(c.Request.Context(), claims.ID)
	if err != nil {
		slog.Debug(err.Error())
		NotFound(c, "")
		return
	}
	if err := os.Remove(imgPath); err != nil {
		slog.Info(`failed to remove file "%s", %s`, imgPath, err)
	}
	c.Status(http.StatusNoContent)
}

func (uph *userProfileHandler) setPhoneNumber(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json queries.PhoneNumberRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		slog.Debug(err.Error())
		BadRequest(c, "")
		return
	}
	if err := uph.ups.SetPhoneNumber(c.Request.Context(), claims.ID, &json); err != nil {
		slog.Debug(err.Error())
		NotFound(c, "")
		return
	}
	c.Status(http.StatusAccepted)
}
