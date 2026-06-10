package api

import (
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/db/cache"
	"generic-shop-sample/db/database"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/payment"
	md "generic-shop-sample/middlewares"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
)

func PaymentRouter(ctx context.Context, router *gin.RouterGroup) {
	config := internal.GetConfig()

	addr := payment.ZPSandboxAddr
	if config.Opt.Mode == gin.ReleaseMode {
		addr = payment.ZPGatewayAddr
	}

	session := database.GetSession()
	ph := paymentHandler{
		cache:       cache.GetCache(cache.PaymentCache),
		userStore:   queries.NewUserStore(session),
		orderStore:  queries.NewOrderStore(session),
		zpGateway:   payment.NewZarinPalGateway(addr, &http.Client{Timeout: 10 * time.Second}),
		merchandID:  config.Opt.ZPMerchantID,
		callbackURL: config.Opt.PaymentCallbackURL,
	}

	rl := md.NewRateLimiter(ctx, 10, 30*time.Minute, 60*time.Second)
	router.Use(rl.RateLimiterMiddleware())
	router.GET("/callback", ph.callback)
	router.POST("/:id", md.AuthMiddleware(), ph.init)
}

type UserPayment struct {
	UserID  int32  `json:"user_id"`
	OrderID string `json:"order_id"`
	Amount  int64  `json:"Amount"`
}

type paymentHandler struct {
	cache       cache.CacheClient
	userStore   queries.UserStore
	orderStore  queries.OrderStore
	zpGateway   payment.ZPGateway
	merchandID  string
	callbackURL string
}

// @Summary		Initiate payment
// @Description	Initializes a payment request for the authenticated user's order
// @Tags			payment
// @Accept			json
// @Produce		json
// @Param			id	path		string				true	"Order ID"
// @Success		301	{string}	string				"Redirect to payment gateway"
// @Failure		400	{object}	map[string]string	"Bad Request"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string	"Not Found"
// @Failure		422	{object}	map[string]string	"Unprocessable Entity"
// @Security		CookieAuth
// @Router			/payment/{id} [post]
func (ph *paymentHandler) init(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	order, err := ph.orderStore.Get(ctx, id, claims.ID)
	if err != nil {
		NotFound(c, "Order not found")
		return
	}
	if !order.IsConfirmed {
		Forbidden(c, "Unconfirmed order")
		return
	}
	user, err := ph.userStore.GetDetails(ctx, claims.Username)
	if err != nil {
		slog.Error("unexpected error in UserStore.GetDetails in payment init",
			"user_id", claims.ID,
			"error", err)
		NotFound(c, "")
		return
	}

	init, err := ph.zpGateway.InitReq(ctx, &payment.ZPRequest{
		MerchantID:  ph.merchandID,
		Amount:      order.TotalBill,
		Currency:    "IRT",
		Description: fmt.Sprintf("%s-%s", user.Username, order.StartedAt),
		CallbackURL: ph.callbackURL,
		Metadata: payment.ZPReqMetadata{
			Mobile:  user.PhoneNumber,
			Email:   user.Email,
			OrderID: order.ID,
		},
	})
	if err != nil {
		slog.Error("failed to init payment gateway", "error", err)
		Forbidden(c, "")
		return
	}
	if len(init.Errs) > 0 || init.Data.Code != 100 {
		slog.Error("unexpected error happend", "init_output", init)
		BadRequest(c, "")
		return
	}

	output, err := json.Marshal(UserPayment{UserID: user.ID, OrderID: order.ID, Amount: order.TotalBill})
	if err != nil {
		slog.Error("unexpected error in UserPayment encoding", "error", err)
		Unprocessable(c, "")
		return
	}
	if _, err := ph.cache.SetEx(ctx, init.Data.Authority, output, 30*time.Minute).Result(); err != nil {
		slog.Error("failed to cache paymeny authority",
			"user_id", user.ID,
			"error", err)
		Unprocessable(c, "Failed to init gateway")
		return
	}

	gatewayURL := fmt.Sprintf("%s/pg/StartPay/%s", ph.zpGateway.Addr, init.Data.Authority)
	c.Redirect(http.StatusMovedPermanently, gatewayURL)
}

// @Summary		Payment callback
// @Description	Handles payment gateway callback and verifies payment status
// @Tags			payment
// @Accept			json
// @Produce		json
// @Param			Authority	query		string				true	"Payment Authority"
// @Param			Status		query		string				true	"Payment Status"
// @Success		202			{object}	map[string]string	"Accepted"
// @Failure		400			{object}	map[string]string	"Bad Request"
// @Failure		401			{object}	map[string]string	"Unauthorized"
// @Failure		403			{object}	map[string]string	"Forbidden"
// @Failure		404			{object}	map[string]string	"Not Found"
// @Failure		422			{object}	map[string]string	"Unprocessable Entity"
// @Security		CookieAuth
// @Router			/payment/callback [get]
func (ph *paymentHandler) callback(c *gin.Context) {
	var input payment.ZPGatewayStatus
	if err := c.ShouldBindQuery(&input); err != nil {
		BadRequest(c, "")
		return
	}
	if !ph.zpGateway.CheckStatus(input.Status) {
		Forbidden(c, "")
		return
	}

	ctx := c.Request.Context()
	reverse := &payment.ZPReverseRequest{MerchantID: ph.merchandID, Authority: input.Authority}
	data, err := ph.cache.Get(ctx, input.Authority).Result()
	if err != nil {
		ph.reverseWithLog(ctx, "failed to get payment authority related data", err, reverse)
		BadRequest(c, "Invliad authority")
		return
	}
	var up UserPayment
	if err := json.Unmarshal([]byte(data), &up); err != nil {
		ph.reverseWithLog(ctx, "failed to decode api.paymeny.UserPayment", err, reverse)
		Unprocessable(c, "")
		return
	}
	verfiedPayment, err := ph.zpGateway.VerifyReq(ctx, &payment.ZPVerifyRequest{
		MerchantID: ph.merchandID,
		Amount:     up.Amount,
		Authority:  input.Authority,
	})
	if err != nil {
		ph.reverseWithLog(ctx, "failed to verify payment", err, reverse)
		Unauthorized(c, "")
		return
	}

	status := &queries.PaymentStatus{}
	output, err := json.Marshal(verfiedPayment)
	if err != nil {
		slog.Warn("failed to encode internal.payment.ZPVerifyRequest", "error", err)
		status.PaymentSummary = fmt.Sprintf("%v", &verfiedPayment)
	}
	status.PaymentSummary = string(output)

	if slices.Contains([]int{100, 101}, verfiedPayment.Data.Code) {
		status.IsPaid = true
	}
	if err := ph.orderStore.SetPaymentStatus(c.Request.Context(), up.OrderID, up.UserID, status); err != nil {
		ph.reverseWithLog(ctx, "failed to set payment summary", err, reverse)
		NotFound(c, "Order not found")
		return
	}
	Accepted(c, "")
}

func (ph *paymentHandler) reverseWithLog(ctx context.Context, reason string, err error, payload *payment.ZPReverseRequest) {
	result, rerr := ph.zpGateway.ReverseReq(ctx, payload)
	slog.Error(fmt.Sprintf(`transaction reversed: %s`, reason),
		"error", err,
		"reverse_res", result,
		"reverse_error", rerr,
	)
}
