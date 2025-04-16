package issuer

import (
	"context"
	"github.com/AliyunContainerService/ack-ram-tool/pkg/credentials/provider"
	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/api/v1beta"
	"github.com/aliyun/credentials-go/credentials"
	"time"
)

const (
	oidcRoleSessionName = "alibabacloud-privateca-issuer"
	oidcTokenFilePath   = "/var/run/secrets/tokens/alibabacloud-privateca-issuer"
)

type AuthConfig struct {
	ClientName            string
	RoleArn               string
	OidcArn               string
	RoleSessionName       string
	RoleSessionExpiration string
	RemoteRoleArn         string
	RemoteRoleSessionName string
	RefreshPeriod         time.Duration
}

func (m *IssuerManager) GetAuthCred(region string, maxConcurrentCount int, a *AuthConfig) (credentials.Credential, error) {
	providers := make([]provider.CredentialsProvider, 0)
	var semaphoreProvider *provider.SemaphoreProvider
	if a.OidcArn != "" && a.RoleArn != "" {
		oidcProvider := provider.NewOIDCProvider(provider.OIDCProviderOptions{
			STSEndpoint:     provider.GetSTSEndpoint(region, true),
			SessionName:     oidcRoleSessionName,
			OIDCTokenFile:   oidcTokenFilePath,
			RoleArn:         a.RoleArn,
			OIDCProviderArn: a.OidcArn,
			RefreshPeriod:   a.RefreshPeriod,
		})
		providers = append(providers, oidcProvider)
	}
	providers = append(providers, provider.NewECSMetadataProvider(provider.ECSMetadataProviderOptions{
		RefreshPeriod: a.RefreshPeriod,
	}))

	chainProvider := provider.NewChainProvider(providers...)
	var remoteRoleProvider *provider.RoleArnProvider
	var cred *provider.CredentialForV2SDK
	if a.RemoteRoleArn != "" && a.RemoteRoleSessionName != "" {
		remoteRoleProvider = provider.NewRoleArnProvider(chainProvider, a.RemoteRoleArn, provider.RoleArnProviderOptions{
			STSEndpoint:   provider.GetSTSEndpoint(region, true),
			SessionName:   a.RemoteRoleSessionName,
			RefreshPeriod: a.RefreshPeriod,
		})
		semaphoreProvider = provider.NewSemaphoreProvider(remoteRoleProvider, provider.SemaphoreProviderOptions{
			MaxWeight: int64(maxConcurrentCount),
		})
	} else {
		semaphoreProvider = provider.NewSemaphoreProvider(chainProvider, provider.SemaphoreProviderOptions{
			MaxWeight: int64(maxConcurrentCount),
		})
	}
	m.RegisterRamProvider(a.ClientName, semaphoreProvider)
	cred = provider.NewCredentialForV2SDK(semaphoreProvider, provider.CredentialForV2SDKOptions{
		CredentialRetrievalTimeout: 10 * time.Minute,
	})
	return cred, nil
}

func (m *IssuerManager) CreateAuthConfig(ctx context.Context, key string, issuerSpec *v1beta.PCAIssuerSpec) (*AuthConfig, error) {
	authConfig := &AuthConfig{
		ClientName:            key,
		RoleArn:               issuerSpec.RAMRoleARN,
		OidcArn:               issuerSpec.OIDCProviderARN,
		RoleSessionName:       issuerSpec.RAMRoleSessionName,
		RoleSessionExpiration: issuerSpec.RoleSessionExpiration,
		RemoteRoleArn:         issuerSpec.RemoteRAMRoleARN,
		RemoteRoleSessionName: issuerSpec.RemoteRAMRoleSessionName,
		RefreshPeriod:         time.Minute * 10,
	}
	return authConfig, nil
}
