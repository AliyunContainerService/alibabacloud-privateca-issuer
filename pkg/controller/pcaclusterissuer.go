package controller

import (
	"context"
	"fmt"
	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/api/v1beta"
	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/pkg/issuer"
	cas20200630 "github.com/alibabacloud-go/cas-20200630/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	casEndpoint  = "cas.aliyuncs.com"
	pcaFinalizer = "finalizer.pcaissuer.alibabacloud.com"
)

type PCAClusterIssuerReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Log           logr.Logger
	Ctx           context.Context
	IssuerManager *issuer.IssuerManager
}

func (p *PCAClusterIssuerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := p.Log.WithValues("PCAClusterIssuerReconciler", req.NamespacedName)
	pcaClusterIssuer := &v1beta.PCAClusterIssuer{}
	err := p.Get(p.Ctx, req.NamespacedName, pcaClusterIssuer)
	if err != nil {
		log.Error(err, fmt.Sprintf("could not get PCAClusterIssuer '%s'", req.NamespacedName))
		return ctrl.Result{}, err
	}
	key := fmt.Sprintf("%s/%s", "cluster-issuer", pcaClusterIssuer.GetName())
	if pcaClusterIssuer.GetDeletionTimestamp() != nil {
		if utils.Contains(pcaClusterIssuer.GetFinalizers(), pcaFinalizer) {
			p.IssuerManager.Delete(key)
			// remove secretFinalizer
			log.Info("removing finalizer", "currentFinalizers", pcaClusterIssuer.GetFinalizers())
			pcaClusterIssuer.SetFinalizers(utils.Remove(pcaClusterIssuer.GetFinalizers(), pcaFinalizer))
			err = p.Update(context.TODO(), pcaClusterIssuer)
			if err != nil {
				log.Error(err, "failed to update externalSec when clean finalizers")
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}
	authConfig, err := p.IssuerManager.CreateAuthConfig(ctx, key, &pcaClusterIssuer.Spec)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("create auth config %s error %v", key, err)
	}
	cred, err := p.IssuerManager.GetAuthCred(p.IssuerManager.Region, p.IssuerManager.MaxConcurrentCount, authConfig)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("get auth cred %s error %v", key, err)
	}
	pcaClient, err := cas20200630.NewClient(&openapi.Config{
		Endpoint:   tea.String(casEndpoint),
		Credential: cred,
	})
	p.IssuerManager.Register(key, pcaClient)
	return ctrl.Result{}, nil
}

func (p *PCAClusterIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	options := controller.Options{
		Reconciler: p,
	}
	PCAClusterIssuerController, err := controller.New("PCAClusterIssuer-controller", mgr, options)
	if err != nil {
		return err
	}
	err = PCAClusterIssuerController.Watch(source.Kind(mgr.GetCache(), &v1beta.PCAClusterIssuer{}, &handler.TypedEnqueueRequestForObject[*v1beta.PCAClusterIssuer]{}, PCAClusterIssuerPredicate[*v1beta.PCAClusterIssuer]{}))
	if err != nil {
		return err
	}
	return nil
}
