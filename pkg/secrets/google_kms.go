package secrets

import (
	"context"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type GoogleKMSConfig struct {
	Log             *logrus.Entry
	KeyName         string
	URL             *string
	APIKey          *string
	CredentialsFile *string
}

type GoogleKMS struct {
	log    *logrus.Entry
	config GoogleKMSConfig
	client *kms.KeyManagementClient
}

func NewGoogleKMS(ctx context.Context, config GoogleKMSConfig) (KeyManagement, error) {
	options := make([]option.ClientOption, 0)
	if config.URL != nil && *config.URL != "" {
		options = append(options, option.WithEndpoint(*config.URL))
	}

	if config.APIKey != nil && *config.APIKey != "" {
		options = append(options, option.WithAPIKey(*config.APIKey))
	}

	if config.CredentialsFile != nil && *config.CredentialsFile != "" {
		options = append(options, option.WithCredentialsFile(*config.CredentialsFile))
	}

	client, err := kms.NewKeyManagementClient(ctx, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create google key management client")
	}

	requiredPermissions := []string{
		"cloudkms.cryptoKeyVersions.useToDecrypt",
		"cloudkms.cryptoKeyVersions.useToEncrypt",
	}

	// Check to make sure that we have permissions to access the specified key name using our current credentials.
	permissions, err := client.ResourceIAM(config.KeyName).TestPermissions(ctx, requiredPermissions)
	if err != nil {
		defer client.Close()
		return nil, errors.Wrap(err, "could not validate permissions to use google KMS")
	}
	if len(permissions) != len(requiredPermissions) {
		defer client.Close()
		return nil, errors.Errorf("insufficient permissions to use google KMS, required: %+v have: %+v", requiredPermissions, permissions)
	}

	return &GoogleKMS{
		log:    config.Log,
		config: config,
		client: client,
	}, nil
}

func (g *GoogleKMS) Encrypt(ctx context.Context, input []byte) (keyID, version string, result []byte, _ error) {
	span := sentry.StartSpan(ctx, "Encrypt KMS")
	defer span.Finish()
	span.SetTag("kms", "google")

	span.Data = map[string]interface{}{
		"resource": g.config.KeyName,
	}

	request := &kmspb.EncryptRequest{
		Name:                        g.config.KeyName,
		Plaintext:                   input,
		AdditionalAuthenticatedData: nil,
		PlaintextCrc32C: &wrapperspb.Int64Value{
			Value: 0, // TODO Add a CRC32 hash of the input.
		},
		AdditionalAuthenticatedDataCrc32C: nil,
	}

	response, err := g.client.Encrypt(context.Background(), request)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return "", "", nil, errors.Wrap(err, "failed to encrypt data using Google KMS")
	}

	return response.Name, "", response.Ciphertext, nil
}

func (g *GoogleKMS) Decrypt(ctx context.Context, keyID, version string, input []byte) (result []byte, _ error) {
	//TODO implement me
	panic("implement me")
}
