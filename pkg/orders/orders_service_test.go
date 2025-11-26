package orders

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	keycloakmocks "gitlab.bbdev.team/vh/vh-srv-profile/internal/mocks/pkg/keycloak"
)

func TestOrdersAPI_GetAccountByID_Success(t *testing.T) {
	expectedAccount := Account{
		ID:      1,
		Email:   "test@example.com",
		UserKey: "user123",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/account/1", r.URL.Path)
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))

		response := AccountRes{
			MessageAndSuccess: MessageAndSuccess{Success: true},
			Data:              expectedAccount,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil)

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	account, err := api.GetAccountByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, expectedAccount.ID, account.ID)
	assert.Equal(t, expectedAccount.Email, account.Email)
}

func TestOrdersAPI_GetAccountByID_401RetrySuccess(t *testing.T) {
	expectedAccount := Account{
		ID:      1,
		Email:   "test@example.com",
		UserKey: "user123",
	}

	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(APIError{Error: "unauthorized"})
			return
		}

		assert.Equal(t, "Bearer token456", r.Header.Get("Authorization"))

		response := AccountRes{
			MessageAndSuccess: MessageAndSuccess{Success: true},
			Data:              expectedAccount,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil).Once()
	mockTokenSource.EXPECT().Invalidate().Once()
	mockTokenSource.EXPECT().Token().Return("token456", nil).Once()

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	account, err := api.GetAccountByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, expectedAccount.ID, account.ID)
	assert.Equal(t, 2, attemptCount)
}

func TestOrdersAPI_GetAccountByID_401RetryFailure(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIError{Error: "unauthorized"})
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil).Once()
	mockTokenSource.EXPECT().Invalidate().Once()
	mockTokenSource.EXPECT().Token().Return("token456", nil).Once()

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	account, err := api.GetAccountByID(ctx, 1)

	require.Error(t, err)
	assert.Nil(t, account)
	assert.Contains(t, err.Error(), "unauthorized")
	assert.Equal(t, 2, attemptCount)
}

func TestOrdersAPI_GetAccountByID_OtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIError{Error: "internal server error"})
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil).Once()
	// Invalidate should NOT be called for non-401 errors
	mockTokenSource.AssertNotCalled(t, "Invalidate")

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	account, err := api.GetAccountByID(ctx, 1)

	require.Error(t, err)
	assert.Nil(t, account)
	assert.Contains(t, err.Error(), "internal server error")
}

func TestOrdersAPI_GetAccountByEmailIfExist_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil)

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	account, err := api.GetAccountByEmailIfExist(ctx, "test@example.com")

	require.NoError(t, err)
	assert.Nil(t, account)
}

func TestOrdersAPI_GetOrders_Success(t *testing.T) {
	expectedOrders := []Order{
		{ID: 1, AccountID: 1, Status: "active"},
		{ID: 2, AccountID: 1, Status: "completed"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/orders", r.URL.Path)
		assert.Equal(t, "test@example.com", r.URL.Query().Get("email"))

		response := OrdersRes{
			MessageAndSuccess: MessageAndSuccess{Success: true},
			Data:              expectedOrders,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil)

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	orders, err := api.GetOrders(ctx, "test@example.com", "", false, "", 0, 0)

	require.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, expectedOrders[0].ID, orders[0].ID)
}

func TestOrdersAPI_GetOrderByID_401RetrySuccess(t *testing.T) {
	expectedOrder := Order{
		ID:        1,
		AccountID: 1,
		Status:    "active",
	}

	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		response := OrderRes{
			MessageAndSuccess: MessageAndSuccess{Success: true},
			Data:              expectedOrder,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil).Once()
	mockTokenSource.EXPECT().Invalidate().Once()
	mockTokenSource.EXPECT().Token().Return("token456", nil).Once()

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	order, err := api.GetOrderByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, expectedOrder.ID, order.ID)
}

func TestOrdersAPI_GetOrderPayments_Success(t *testing.T) {
	expectedPayments := []Payment{
		{ID: 1, OrderID: 1, Amount: 100.0},
		{ID: 2, OrderID: 1, Amount: 200.0},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := MultiplePaymentRes{
			MessageAndSuccess: MessageAndSuccess{Success: true},
			Data:              expectedPayments,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil)

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	payments, err := api.GetOrderPayments(ctx, 1, "", 0, 0)

	require.NoError(t, err)
	assert.Len(t, payments, 2)
}

func TestOrdersAPI_GetPaymentByID_401RetrySuccess(t *testing.T) {
	expectedPayment := Payment{
		ID:      1,
		OrderID: 1,
		Amount:  100.0,
	}

	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		response := PaymentRes{
			MessageAndSuccess: MessageAndSuccess{Success: true},
			Data:              expectedPayment,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil).Once()
	mockTokenSource.EXPECT().Invalidate().Once()
	mockTokenSource.EXPECT().Token().Return("token456", nil).Once()

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	payment, err := api.GetPaymentByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, expectedPayment.ID, payment.ID)
}

func TestOrdersAPI_GetSpecials_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil)

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	specials, err := api.GetSpecials(ctx, "test@example.com")

	require.NoError(t, err)
	assert.Nil(t, specials)
}

func TestOrdersAPI_DeleteSpecial_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v2/special/1", r.URL.Path)

		response := MessageAndSuccess{Success: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil)

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	err := api.DeleteSpecial(ctx, 1)

	require.NoError(t, err)
}

func TestOrdersAPI_UpdateSpecialSetKeycloakIdByEmail_401RetrySuccess(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/special/update", r.URL.Path)

		response := MessageAndSuccess{Success: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil).Once()
	mockTokenSource.EXPECT().Invalidate().Once()
	mockTokenSource.EXPECT().Token().Return("token456", nil).Once()

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	payload := map[string]interface{}{
		"email":       "test@example.com",
		"keycloak_id": "kc123",
	}
	err := api.UpdateSpecialSetKeycloakIdByEmail(ctx, payload)

	require.NoError(t, err)
}

func TestOrdersAPI_DeleteSpecialIfExist_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil)

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	err := api.DeleteSpecialIfExist(ctx, "kc123")

	require.NoError(t, err)
}

func TestOrdersAPI_StatusByEmail_Success(t *testing.T) {
	expectedStatus := Status{
		Membership:  true,
		Ticket:      false,
		StatusName:  "active",
		StatusColor: "green",
		IsSpecial:   false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedStatus)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil)

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	status, err := api.StatusByEmail(ctx, "test@example.com")

	require.NoError(t, err)
	assert.Equal(t, expectedStatus.Membership, status.Membership)
}

func TestOrdersAPI_StatusByEmail_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil)

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	status, err := api.StatusByEmail(ctx, "test@example.com")

	require.NoError(t, err)
	assert.Nil(t, status)
}

func TestOrdersAPI_SetOrderStartingDate_401RetrySuccess(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v2/order/1", r.URL.Path)

		response := MessageAndSuccess{Success: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil).Once()
	mockTokenSource.EXPECT().Invalidate().Once()
	mockTokenSource.EXPECT().Token().Return("token456", nil).Once()

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	startingDate := time.Now()
	err := api.SetOrderStartingDate(ctx, 1, startingDate)

	require.NoError(t, err)
}

func TestOrdersAPI_CancelOrder_401RetrySuccess(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v2/order/1", r.URL.Path)

		response := MessageAndSuccess{Success: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockTokenSource := keycloakmocks.NewMockTokenSource(t)
	mockTokenSource.EXPECT().Token().Return("token123", nil).Once()
	mockTokenSource.EXPECT().Invalidate().Once()
	mockTokenSource.EXPECT().Token().Return("token456", nil).Once()

	api := newTestOrdersAPI(server.URL)
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, mockTokenSource)

	err := api.CancelOrder(ctx, 1)

	require.NoError(t, err)
}

// Helper function to create a test OrdersAPI with a test server URL
func newTestOrdersAPI(serverURL string) *OrdersAPI {
	client := resty.New()
	client.SetBaseURL(serverURL)
	client.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "test",
	})
	client.SetError(APIError{})

	return &OrdersAPI{client: client}
}
