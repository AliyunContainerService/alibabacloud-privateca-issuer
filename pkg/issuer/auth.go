package issuer

import (
	"context"
	"github.com/AliyunContainerService/ack-ram-tool/pkg/credentials/provider"
	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/api/v1beta"
	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/pkg/utils"
	"github.com/aliyun/credentials-go/credentials"
	"k8s.io/klog/v2"
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
	AccessKey             string
	AccessSecretKey       string
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
	if a.AccessKey != "" && a.AccessSecretKey != "" && a.RoleSessionName != "" && a.RoleArn != "" {
		ramRoleProvider := provider.NewRoleArnProvider(provider.NewAccessKeyProvider(a.AccessKey, a.AccessSecretKey), a.RoleArn, provider.RoleArnProviderOptions{
			STSEndpoint:   provider.GetSTSEndpoint(region, true),
			SessionName:   a.RoleSessionName,
			RefreshPeriod: a.RefreshPeriod,
		})
		providers = append(providers, ramRoleProvider)
	}
	if a.AccessKey != "" && a.AccessSecretKey != "" {
		akProvider := provider.NewAccessKeyProvider(a.AccessKey, a.AccessSecretKey)
		providers = append(providers, akProvider)
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

func (m *IssuerManager) createAuthConfig(ctx context.Context, key string, issuerSpec *v1beta.PCAIssuerSpec) (*AuthConfig, error) {
	var accessKey, accessKeySecret string
	var authConfig *AuthConfig
	if issuerSpec.AccessKey != nil {
		accessKeyBytes, err := utils.GetConfigFromSecret(ctx, m.KubeClient, issuerSpec.AccessKey)
		if err != nil {
			klog.Errorf("get access key from secret error %v", err)
			return nil, err
		}
		accessKey = string(accessKeyBytes)
	}
	if issuerSpec.AccessKeySecret != nil {
		accessKeySecretBytes, err := utils.GetConfigFromSecret(ctx, m.KubeClient, issuerSpec.AccessKeySecret)
		if err != nil {
			klog.Errorf("get access key secret from secret error %v", err)
			return nil, err
		}
		accessKeySecret = string(accessKeySecretBytes)
	}
	authConfig = &AuthConfig{
		ClientName:            key,
		RoleArn:               issuerSpec.RAMRoleARN,
		OidcArn:               issuerSpec.OIDCProviderARN,
		AccessKey:             accessKey,
		AccessSecretKey:       accessKeySecret,
		RoleSessionName:       issuerSpec.RAMRoleSessionName,
		RoleSessionExpiration: issuerSpec.RoleSessionExpiration,
		RemoteRoleArn:         issuerSpec.RemoteRAMRoleARN,
		RemoteRoleSessionName: issuerSpec.RemoteRAMRoleSessionName,
		RefreshPeriod:         time.Minute * 10,
	}
	return authConfig, nil
}
