package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ZPGatewayAddr = "https://payment.zarinpal.com"
	ZPSandboxAddr = "https://sandbox.zarinpal.com"
)

type ZPGatewayStatus struct {
	Authority string `form:"Authority"`
	Status    string `form:"Status"`
}

type ZPReqMetadata struct {
	Mobile  string `json:"mobile"`
	Email   string `json:"email"`
	OrderID string `json:"order_id"`
}

type ZPRequest struct {
	MerchantID  string        `json:"merchant_id"`
	Amount      int64         `json:"amount"`
	Currency    string        `json:"currency"`
	Description string        `json:"description"`
	CallbackURL string        `json:"callback_url"`
	Metadata    ZPReqMetadata `json:"metadata"`
}

type ZPVerifyRequest struct {
	MerchantID string `json:"merchant_id"`
	Amount     int64  `json:"amount"`
	Authority  string `json:"authority"`
}

type ZPReverseRequest struct {
	MerchantID string `json:"merchant_id"`
	Authority  string `json:"authority"`
}

type ZPResMetadata struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Authority string `json:"authority"`
	FeeType   string `json:"fee_type"`
	Fee       int    `json:"fee"`
}

type ZPResponse struct {
	Data ZPResMetadata `json:"data"`
	Errs []any         `json:"errors"`
}

type ZPResVerifyMetadata struct {
	Code     int    `json:"code"`
	RefID    int    `json:"ref_id"`
	CardPan  string `json:"card_pan"`
	CardHash string `json:"card_hash"`
	FeeType  string `json:"fee_type"`
	Fee      int    `json:"fee"`
}

type ZPVerifyReponse struct {
	Data ZPResVerifyMetadata `json:"data"`
	Errs []any               `json:"errors"`
}

type ZPGateway struct {
	Addr   string
	client *http.Client
}

func NewZarinPalGateway(addr string, client *http.Client) ZPGateway {
	if client == nil {
		client = http.DefaultClient
	}
	return ZPGateway{addr, client}
}

func (zg *ZPGateway) InitReq(ctx context.Context, payload *ZPRequest) (*ZPResponse, error) {
	var output ZPResponse
	err := zg.doJSON(ctx, "/pg/v4/payment/request.json", payload, &output)
	return &output, err
}

func (zg *ZPGateway) CheckStatus(status string) bool {
	return strings.EqualFold(status, "OK")
}

func (zg *ZPGateway) VerifyReq(ctx context.Context, payload *ZPVerifyRequest) (*ZPVerifyReponse, error) {
	var output ZPVerifyReponse
	err := zg.doJSON(ctx, "/pg/v4/payment/verify.json", payload, &output)
	return &output, err
}

func (zg *ZPGateway) ReverseReq(ctx context.Context, payload *ZPReverseRequest) (string, error) {
	var res ZPResponse
	err := zg.doJSON(ctx, "/pg/v4/payment/reverse.json", payload, &res)
	output, _ := json.Marshal(res)
	return string(output), err
}

func (zg *ZPGateway) doJSON(ctx context.Context, path string, payload, output any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode json, %s", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s%s", zg.Addr, path), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create new request, %s", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err := zg.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request, %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(res.Body)
		return fmt.Errorf(`bad status="%d", buf="%s"`, res.StatusCode, buf)
	}
	return json.NewDecoder(res.Body).Decode(output)
}
