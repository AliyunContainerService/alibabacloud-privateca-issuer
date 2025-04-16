package issuer

import (
	"context"
	"fmt"
	"time"

	cas20200630 "github.com/alibabacloud-go/cas-20200630/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/cert-manager/cert-manager/pkg/util/pki"
	issuerapi "github.com/cert-manager/issuer-lib/api/v1alpha1"
	"github.com/cert-manager/issuer-lib/controllers"
	"github.com/cert-manager/issuer-lib/controllers/signer"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/api/v1beta"
)

func (i *IssuerManager) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return (&controllers.CombinedController{
		IssuerTypes:        []issuerapi.Issuer{&v1beta.PCAIssuer{}},
		ClusterIssuerTypes: []issuerapi.Issuer{&v1beta.PCAClusterIssuer{}},

		FieldOwner:       "pcaissuer.alibabacloud.com",
		MaxRetryDuration: 1 * time.Minute,
		Sign:             i.Sign,
		Check:            i.Check,
		EventRecorder:    mgr.GetEventRecorderFor("pcaissuer.alibabacloud.com"),
	}).SetupWithManager(ctx, mgr)
}

func (i *IssuerManager) Sign(ctx context.Context, cr signer.CertificateRequestObject, issuerObject issuerapi.Issuer) (signer.PEMBundle, error) {
	waitTimeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	err := i.signingLimiter.Wait(waitTimeoutCtx)
	if err != nil {
		return signer.PEMBundle{}, err
	}
	key := ""
	issuerSpec, key, err := i.getIssuerDetails(issuerObject)
	if err != nil {
		return signer.PEMBundle{}, err
	}

	pcaClient, err := i.GetClient(key)
	if err != nil || pcaClient == nil {
		return signer.PEMBundle{}, fmt.Errorf("get pca client %s error %v", key, err)
	}

	createCustomCertificateRequest, err := i.CreateCustomCertificateReq(issuerSpec, cr)
	if err != nil {
		return signer.PEMBundle{}, err
	}
	resp, err := pcaClient.CreateCustomCertificate(createCustomCertificateRequest)
	if err != nil {
		return signer.PEMBundle{}, err
	}
	if resp != nil && resp.Body != nil {
		pemBundle, err := pki.ParseSingleCertificateChainPEM([]byte(tea.StringValue(resp.Body.Certificate)))
		if err != nil {
			return signer.PEMBundle{}, fmt.Errorf("parse %s single certificate chain pem error %v", key, err)
		}
		return signer.PEMBundle(pemBundle), nil
	}
	return signer.PEMBundle{}, fmt.Errorf("CreateCustomCertificate resp is invalid, certificate obj %s, issuer obj %s",
		fmt.Sprintf("%s/%s", cr.GetName(), cr.GetNamespace()), fmt.Sprintf("%s/%s", issuerObject.GetName(), issuerObject.GetNamespace()))
}

func (i *IssuerManager) Check(ctx context.Context, issuerObject issuerapi.Issuer) error {
	issuerSpec, key, err := i.getIssuerDetails(issuerObject)
	if err != nil {
		return err
	}
	pcaClient, err := i.GetClient(key)
	if err != nil || pcaClient == nil {
		return fmt.Errorf("get pca client %s error %v", key, err)
	}
	caResp, err := pcaClient.DescribeCACertificate(&cas20200630.DescribeCACertificateRequest{
		Identifier: tea.String(issuerSpec.ParentIdentifier),
	})
	if err != nil {
		return fmt.Errorf("describe ca certificate %s state error %v", issuerSpec.ParentIdentifier, err)
	}
	if caResp != nil && caResp.Body != nil && caResp.Body.Certificate != nil {
		if tea.Int64Value(caResp.Body.Certificate.CertRemainingCount) <= 0 {
			return fmt.Errorf("the quota for ca certificate %s has exceeded the limit", issuerSpec.ParentIdentifier)
		}
		beforeTime := time.UnixMilli(tea.Int64Value(caResp.Body.Certificate.BeforeDate)).UTC()
		timeNow := time.Now().UTC()
		afterTime := time.UnixMilli(tea.Int64Value(caResp.Body.Certificate.AfterDate)).UTC()
		if timeNow.Before(beforeTime) || timeNow.After(afterTime) {
			return fmt.Errorf("the current time is not within the validity period of the ca certificate %s", issuerSpec.ParentIdentifier)
		}
		return nil
	}
	return fmt.Errorf("check ca certificate state failed")
}

func (i *IssuerManager) getIssuerDetails(issuerObject issuerapi.Issuer) (*v1beta.PCAIssuerSpec, string, error) {
	switch t := issuerObject.(type) {
	case *v1beta.PCAIssuer:
		return &t.Spec, fmt.Sprintf("%s/%s", issuerObject.GetNamespace(), issuerObject.GetName()), nil
	case *v1beta.PCAClusterIssuer:
		return &t.Spec, fmt.Sprintf("%s/%s", "cluster-issuer", issuerObject.GetName()), nil
	default:
		return nil, "", fmt.Errorf("issuerObject type error, issuerObject is %s/%s", issuerObject.GetNamespace(), issuerObject.GetName())
	}
}

func (i *IssuerManager) CreateCustomCertificateReq(issuerSpec *v1beta.PCAIssuerSpec, cr signer.CertificateRequestObject) (*cas20200630.CreateCustomCertificateRequest, error) {
	createCustomCertificateReq := &cas20200630.CreateCustomCertificateRequest{}
	_, duration, csr, err := cr.GetRequest()
	if err != nil {
		return nil, err
	}
	// issuer certification immediately
	createCustomCertificateReq.Immediately = tea.Int32(2)

	timeNow := time.Now().UTC()
	expireTime := timeNow.Add(duration).UTC()
	notBefore := timeNow.Format("2006-01-02T15:04:05Z")
	notAfter := expireTime.Format("2006-01-02T15:04:05Z")
	createCustomCertificateReq.Validity = tea.String(fmt.Sprintf("%s/%s", notBefore, notAfter))
	createCustomCertificateReq.ParentIdentifier = tea.String(issuerSpec.ParentIdentifier)
	createCustomCertificateReq.Csr = tea.String(string(csr))
	return createCustomCertificateReq, nil
}
