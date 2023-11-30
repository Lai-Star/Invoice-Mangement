package secrets

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

var (
	_ PlaidSecretsProvider = &vaultPlaidSecretsProvider{}
)

type vaultPlaidSecretsProvider struct {
	client *api.Client
}

func (v *vaultPlaidSecretsProvider) UpdateAccessTokenForPlaidLinkId(ctx context.Context, accountId, plaidLinkId uint64, accessToken string) error {
	span := sentry.StartSpan(ctx, "UpdateAccessTokenForPlaidLinkId")
	defer span.Finish()
	panic("implement me")
}

func (v *vaultPlaidSecretsProvider) GetAccessTokenForPlaidLinkId(ctx context.Context, accountId, plaidLinkId uint64) (accessToken string, err error) {
	span := sentry.StartSpan(ctx, "GetAccessTokenForPlaidLinkId")
	defer span.Finish()

	result, err := v.client.Logical().Read(v.buildPath(accountId, plaidLinkId))
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve access token")
	}

	fmt.Sprint(result)

	panic("implement me")
}

func (v *vaultPlaidSecretsProvider) buildPath(accountId, plaidLinkId uint64) string {
	return fmt.Sprintf("customers/plaid/%X/%X", accountId, plaidLinkId)
}

func (v *vaultPlaidSecretsProvider) Close() error {
	panic("implement me")
}
