package controller_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/pkg/internal/mock_plaid"
	"net/http"
	"testing"
)

func TestPostTokenCallback(t *testing.T) {
	t.Run("cant retrieve accounts", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		publicToken := mock_plaid.MockExchangePublicToken(t)
		mock_plaid.MockGetAccounts(t, nil)

		response := e.POST("/api/plaid/link/token/callback").
			WithHeader("M-Token", token).
			WithJSON(map[string]interface{}{
				"publicToken":     publicToken,
				"institutionId":   "123",
				"institutionName": gofakeit.Company(),
				"accountIds": []string{
					gofakeit.UUID(),
				},
			}).
			Expect()

		response.Status(http.StatusInternalServerError)
		response.JSON().Path("$.error").String().Equal("could not retrieve details for any accounts")
	})
}
