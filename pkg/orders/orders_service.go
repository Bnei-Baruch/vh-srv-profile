package orders

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/keycloak"
)

type OrdersService interface {
	GetAccountByID(ctx context.Context, accountID int) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetOrders(ctx context.Context,
		email string,
		productType string,
		evaluateMembership bool,
		paymentDateOrder string,
		limit int, offset int) ([]Order, error)
	GetOrderByID(ctx context.Context, orderID int) (*Order, error)
	SetOrderStartingDate(ctx context.Context, orderID int, startingDate time.Time) error
	CancelOrder(ctx context.Context, orderID int) error
	GetOrderPayments(ctx context.Context,
		orderID int,
		createdAtOrder string,
		limit int, offset int) ([]Payment, error)
	GetPaymentByID(ctx context.Context, paymentID int) (*Payment, error)
	GetSpecials(ctx context.Context, email string) ([]Special, error)
	UpdateSpecialSetKeycloakIdByEmail(ctx context.Context, payload map[string]interface{}) error
	DeleteSpecial(ctx context.Context, id int) error
	DeleteSpecialIfExist(ctx context.Context, email string) error
	StatusByEmail(ctx context.Context, email string) (*Status, error)
}

type OrdersServiceFactory func() OrdersService

func OrdersAPIFactory() OrdersService {
	return NewOrdersAPI()
}

type OrdersAPI struct {
	client *resty.Client
}

func NewOrdersAPI() *OrdersAPI {
	client := resty.New()
	client.SetBaseURL(common.Config.OrdersServiceUrl)
	client.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   common.ServiceName, // important. This value is used in orders events stream as "actor"
	})
	client.SetError(APIError{})
	//client.EnableTrace()

	return &OrdersAPI{client: client}
}

func (api *OrdersAPI) GetAccountByID(ctx context.Context, accountID int) (*Account, error) {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	resp, err := req.
		SetPathParam("accountID", strconv.Itoa(accountID)).
		SetResult(AccountRes{}).
		Get("/v2/account/{accountID}")
	if err != nil {
		return nil, fmt.Errorf("req.Get: %w", err)
	}

	if err = respError(resp); err != nil {
		return nil, err
	}

	account := resp.Result().(*AccountRes).Data
	return &account, nil
}

func (api *OrdersAPI) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	resp, err := req.
		SetPathParam("email", email).
		SetResult(AccountRes{}).
		Get("/v2/account/email/{email}")
	if err != nil {
		return nil, fmt.Errorf("req.Get: %w", err)
	}

	if err = respError(resp); err != nil {
		return nil, err
	}

	account := resp.Result().(*AccountRes).Data
	return &account, nil
}

func (api *OrdersAPI) GetOrders(ctx context.Context,
	email string,
	productType string,
	evaluateMembership bool,
	paymentDateOrder string,
	limit int, offset int) ([]Order, error) {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	params := make(map[string]string)
	if email != "" {
		params["email"] = email
	}
	if productType != "" {
		params["product-type"] = productType
	}
	if evaluateMembership {
		params["evaluate-membership"] = "true"
	}
	if paymentDateOrder != "" {
		params["o-payment-date"] = paymentDateOrder
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	if offset > 0 {
		params["offset"] = strconv.Itoa(offset)
	}

	resp, err := req.
		SetQueryParams(params).
		SetResult(OrdersRes{}).
		Get("/v2/orders")
	if err != nil {
		return nil, fmt.Errorf("req.Get: %w", err)
	}

	if err = respError(resp); err != nil {
		return nil, err
	}

	return resp.Result().(*OrdersRes).Data, nil
}

func (api *OrdersAPI) GetOrderByID(ctx context.Context, orderID int) (*Order, error) {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	resp, err := req.
		SetPathParam("orderID", strconv.Itoa(orderID)).
		SetResult(OrderRes{}).
		Get("/v2/order/{orderID}")
	if err != nil {
		return nil, fmt.Errorf("req.Get: %w", err)
	}

	if err = respError(resp); err != nil {
		return nil, err
	}

	order := resp.Result().(*OrderRes).Data
	return &order, nil
}

func (api *OrdersAPI) SetOrderStartingDate(ctx context.Context, orderID int, startingDate time.Time) error {
	return api.patchOrder(ctx, orderID, map[string]interface{}{"StartingDate": startingDate})
}

func (api *OrdersAPI) CancelOrder(ctx context.Context, orderID int) error {
	return api.patchOrder(ctx, orderID, map[string]interface{}{"Status": "cancelled"})
}

func (api *OrdersAPI) GetOrderPayments(ctx context.Context,
	orderID int,
	createdAtOrder string,
	limit int, offset int) ([]Payment, error) {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	params := make(map[string]string)
	if orderID != 0 {
		params["order-id"] = strconv.Itoa(orderID)
	}
	if createdAtOrder != "" {
		params["o-created-at"] = createdAtOrder
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	if offset > 0 {
		params["offset"] = strconv.Itoa(offset)
	}

	resp, err := req.
		SetQueryParams(params).
		SetResult(MultiplePaymentRes{}).
		Get("/v2/payments")
	if err != nil {
		return nil, fmt.Errorf("req.Get: %w", err)
	}

	if err = respError(resp); err != nil {
		return nil, err
	}

	return resp.Result().(*MultiplePaymentRes).Data, nil
}

func (api *OrdersAPI) GetPaymentByID(ctx context.Context, paymentID int) (*Payment, error) {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	resp, err := req.
		SetPathParam("paymentID", strconv.Itoa(paymentID)).
		SetResult(PaymentRes{}).
		Get("/v2/payment/{paymentID}")
	if err != nil {
		return nil, fmt.Errorf("req.Get: %w", err)
	}

	if err = respError(resp); err != nil {
		return nil, err
	}

	payment := resp.Result().(*PaymentRes).Data
	return &payment, nil
}

func (api *OrdersAPI) GetSpecials(ctx context.Context, email string) ([]Special, error) {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	resp, err := req.
		SetPathParam("email", email).
		SetResult(SpecialRes{}).
		Get("/v2/special/email/{email}")
	if err != nil {
		return nil, fmt.Errorf("req.Get: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == http.StatusNotFound {
			return nil, nil
		}
		return nil, respError(resp)
	}

	specials := resp.Result().(*SpecialRes).Data
	return specials, nil
}

func (api *OrdersAPI) DeleteSpecial(ctx context.Context, id int) error {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return fmt.Errorf("baseRequest: %w", err)
	}

	resp, err := req.
		SetPathParam("id", strconv.Itoa(id)).
		Delete("/v2/special/{id}")
	if err != nil {
		return fmt.Errorf("req.Delete: %w", err)
	}

	if err = respError(resp); err != nil {
		return err
	}

	return nil
}

// UpdateSpecialSetKeycloakIdByEmail Consider establishing event sending with NATS for 'update' requests from 'orders' to 'profile'
func (api *OrdersAPI) UpdateSpecialSetKeycloakIdByEmail(ctx context.Context, payload map[string]interface{}) error {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return fmt.Errorf("baseRequest: %w", err)
	}

	resp, err := req.
		SetBody(payload).
		Post("/v2/special/update")
	if err != nil {
		return fmt.Errorf("req.UpdateSpecialSetKeycloakIdByEmail: %w", err)
	}

	if err = respError(resp); err != nil {
		return err
	}

	return nil
}

// DeleteSpecialIfExist  method retrieves all specials, then checks which ones are no longer actual or are actual at the moment, and removes them.
// Those that will be actual in the future remain as they are.
func (api *OrdersAPI) DeleteSpecialIfExist(ctx context.Context, keycloakId string) error {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return fmt.Errorf("baseRequest: %w", err)
	}
	resp, err := req.
		SetPathParam("keycloak_id", keycloakId).
		Delete("/v2/special/delete/{keycloak_id}")
	if err != nil {
		fmt.Println("req.Delete: %w", err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		fmt.Println("req.Delete: %w", err)
	}
	if err = respError(resp); err != nil {
		fmt.Println(err)
	}

	return nil
}

func (api *OrdersAPI) StatusByEmail(ctx context.Context, email string) (*Status, error) {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("baseRequest: %w", err)
	}

	resp, err := req.
		SetPathParam("email", email).
		SetResult(Status{}).
		Get("/status/{email}")
	if err != nil {
		return nil, fmt.Errorf("req.Get: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == http.StatusNotFound {
			return nil, nil
		}
		return nil, respError(resp)
	}

	return resp.Result().(*Status), nil
}

func (api *OrdersAPI) patchOrder(ctx context.Context, orderID int, payload map[string]interface{}) error {
	req, err := api.baseRequest(ctx)
	if err != nil {
		return fmt.Errorf("baseRequest: %w", err)
	}

	resp, err := req.
		SetPathParam("orderID", strconv.Itoa(orderID)).
		SetBody(payload).
		Patch("/v2/order/{orderID}")
	if err != nil {
		return fmt.Errorf("req.Patch: %w", err)
	}

	if err = respError(resp); err != nil {
		return err
	}

	return nil
}

func (api *OrdersAPI) baseRequest(ctx context.Context) (*resty.Request, error) {
	r := api.client.NewRequest()
	r.SetContext(ctx)

	tokenSource := ctx.Value(common.CtxTokenSource).(keycloak.TokenSource)
	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("tokenSource.Token(): %w", err)
	}
	r.SetAuthToken(token)

	return r, nil
}

func respError(resp *resty.Response) error {
	if resp.IsError() {
		if apiErr, ok := resp.Error().(*APIError); ok {
			return errors.New(apiErr.Error)
		} else {
			return fmt.Errorf("unexpected response: [%s] %s", resp.Status(), resp.String())
		}
	}
	return nil
}
