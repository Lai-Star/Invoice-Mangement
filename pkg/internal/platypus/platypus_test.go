package platypus

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/rest-api/pkg/config"
	"github.com/monetr/rest-api/pkg/internal/mock_plaid"
	"github.com/monetr/rest-api/pkg/internal/testutils"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlaid_CreateLinkToken(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)
		mock_plaid.MockCreateLinkToken(t)

		platypus := NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
			OAuthDomain:  "localhost",
		})

		linkToken, err := platypus.CreateLinkToken(context.Background(), LinkTokenOptions{
			ClientUserID:             "1234",
			LegalName:                gofakeit.Name(),
			PhoneNumber:              nil,
			PhoneNumberVerifiedTime:  nil,
			EmailAddress:             gofakeit.Email(),
			EmailAddressVerifiedTime: nil,
			RedirectURI:              "",
			UpdateMode:               false,
		})
		assert.NoError(t, err, "should not return an error creating a link token")
		assert.NotEmpty(t, linkToken.Token(), "must not be empty")
	})
}
