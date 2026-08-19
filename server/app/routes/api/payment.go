package api

import (
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/app"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/internal/payment"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
)

func PaymentRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	addr := payment.ZPSandboxAddr
	if deps.Config.App.Mode == gin.ReleaseMode {
		addr = payment.ZPGatewayAddr
	}

	session := deps.DB.GetSession()
	log := logger.GetLogger()
	ph := paymentHandler{
		cache:       deps.Cache.GetCache(cache.PaymentCache),
		userStore:   queries.NewUserStore(session, log),
		orderStore:  queries.NewOrderStore(session, log),
		zpGateway:   payment.NewZarinPalGateway(addr, &http.Client{Timeout: 10 * time.Second}),
		merchandID:  deps.Config.Payment.ZPMerchantID,
		callbackURL: deps.Config.Payment.PaymentCallbackURL,
		log:         log,
	}

	rl := md.NewRateLimiter(deps.Ctx, 10, 30*time.Minute, 60*time.Second)
	router.Use(rl.RateLimiterMiddleware())
	router.GET("/callback", ph.callback)
	router.POST("/:id", md.AuthMiddleware(deps, log), ph.init)
}

type UserPayment struct {
	UserID  string `json:"user_id"`
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
	log         logger.Logger
}

func (h *paymentHandler) init(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	order, err := h.orderStore.Get(ctx, queries.OrderID{
		ID:     id,
		UserID: claims.ID,
	})
	if err != nil {
		NotFound(c, "Order not found")
		return
	}
	if !order.IsConfirmed {
		Forbidden(c, "Unconfirmed order")
		return
	}
	user, err := h.userStore.Get(ctx, claims.ID)
	if err != nil {
		h.log.Error("unexpected error in UserStore.GetDetails in payment init",
			"user_id", claims.ID,
			"error", err)
		NotFound(c, "")
		return
	}

	init, err := h.zpGateway.InitReq(ctx, payment.ZPRequest{
		MerchantID:  h.merchandID,
		Amount:      order.TotalBill,
		Currency:    "IRT",
		Description: fmt.Sprintf("%s-%s", user.ID, order.StartedAt),
		CallbackURL: h.callbackURL,
		Metadata: payment.ZPReqMetadata{
			Mobile:  user.PhoneNumber,
			Email:   user.Email,
			OrderID: order.ID,
		},
	})
	if err != nil {
		h.log.Error("failed to init payment gateway", "error", err)
		Forbidden(c, "")
		return
	}
	if len(init.Errs) > 0 || init.Data.Code != 100 {
		h.log.Error("unexpected error happend", "init_output", init)
		BadRequest(c, "")
		return
	}

	output, err := json.Marshal(UserPayment{UserID: user.ID, OrderID: order.ID, Amount: order.TotalBill})
	if err != nil {
		h.log.Error("unexpected error in UserPayment encoding", "error", err)
		Unprocessable(c, "")
		return
	}
	if _, err := h.cache.Set(ctx, init.Data.Authority, output, 30*time.Minute).Result(); err != nil {
		h.log.Error("failed to cache paymeny authority",
			"user_id", user.ID,
			"error", err)
		Unprocessable(c, "Failed to init gateway")
		return
	}

	gatewayURL := fmt.Sprintf("%s/pg/StartPay/%s", h.zpGateway.Addr, init.Data.Authority)
	c.Redirect(http.StatusMovedPermanently, gatewayURL)
}

func (h *paymentHandler) callback(c *gin.Context) {
	var input payment.ZPGatewayStatus
	if err := c.ShouldBindQuery(&input); err != nil {
		h.log.Debug("paymentHandler.callback", "error", err)
		BadRequest(c, "")
		return
	}
	if !h.zpGateway.CheckStatus(input.Status) {
		Forbidden(c, "")
		return
	}

	ctx := c.Request.Context()
	reverse := payment.ZPReverseRequest{MerchantID: h.merchandID, Authority: input.Authority}
	data, err := h.cache.Get(ctx, input.Authority).Result()
	if err != nil {
		h.reverseWithLog(ctx, "failed to get payment authority related data", err, reverse)
		BadRequest(c, "Invliad authority")
		return
	}
	var up UserPayment
	if err := json.Unmarshal([]byte(data), &up); err != nil {
		h.reverseWithLog(ctx, "failed to decode api.paymeny.UserPayment", err, reverse)
		Unprocessable(c, "")
		return
	}
	verfiedPayment, err := h.zpGateway.VerifyReq(ctx, payment.ZPVerifyRequest{
		MerchantID: h.merchandID,
		Amount:     up.Amount,
		Authority:  input.Authority,
	})
	if err != nil {
		h.reverseWithLog(ctx, "failed to verify payment", err, reverse)
		Unauthorized(c, "")
		return
	}

	status := queries.PaymentStatus{
		PaymentSummary: queries.ProductProperty{
			"code":      fmt.Sprintf("%d", verfiedPayment.Data.Code),
			"ref_id":    fmt.Sprintf("%d", verfiedPayment.Data.RefID),
			"card_pan":  verfiedPayment.Data.CardPan,
			"card_hash": verfiedPayment.Data.CardHash,
			"fee_type":  verfiedPayment.Data.FeeType,
			"fee":       fmt.Sprintf("%d", verfiedPayment.Data.Fee),
			"err":       fmt.Sprintf("%v", verfiedPayment.Errs),
		},
	}
	if slices.Contains([]int{100, 101}, verfiedPayment.Data.Code) {
		status.IsPaid = true
	}
	err = h.orderStore.SetPaymentStatus(c.Request.Context(), queries.OrderID{
		ID:     up.OrderID,
		UserID: up.UserID,
	}, status)
	if err != nil {
		h.reverseWithLog(ctx, "failed to set payment summary", err, reverse)
		NotFound(c, "Order not found")
		return
	}
	Accepted(c, "")
}

func (h *paymentHandler) reverseWithLog(ctx context.Context, reason string, err error, payload payment.ZPReverseRequest) {
	result, rerr := h.zpGateway.ReverseReq(ctx, payload)
	h.log.Error(fmt.Sprintf(`transaction reversed: %s`, reason),
		"error", err,
		"reverse_res", result,
		"reverse_error", rerr,
	)
}
