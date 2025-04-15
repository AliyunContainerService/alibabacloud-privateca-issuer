package controller

import (
	"context"
	"fmt"
	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/api/v1beta"
	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/pkg/issuer"
	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/pkg/utils"
	cas20200630 "github.com/alibabacloud-go/cas-20200630/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type PCAIssuerReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Log           logr.Logger
	Ctx           context.Context
	IssuerManager *issuer.IssuerManager
}

func (p *PCAIssuerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := p.Log.WithValues("PCAIssuerReconciler", req.NamespacedName)
	pcaIssuer := &v1beta.PCAIssuer{}
	err := p.Get(p.Ctx, req.NamespacedName, pcaIssuer)
	if err != nil {
		log.Error(err, fmt.Sprintf("could not get PCAIssuer '%s'", req.NamespacedName))
		return ctrl.Result{}, err
	}
	key := fmt.Sprintf("%s/%s", pcaIssuer.GetNamespace(), pcaIssuer.GetName())
	if pcaIssuer.GetDeletionTimestamp() != nil {
		if utils.Contains(pcaIssuer.GetFinalizers(), pcaFinalizer) {
			p.IssuerManager.Delete(key)
			// remove secretFinalizer
			log.Info("removing finalizer", "currentFinalizers", pcaIssuer.GetFinalizers())
			pcaIssuer.SetFinalizers(utils.Remove(pcaIssuer.GetFinalizers(), pcaFinalizer))
			err = p.Update(context.TODO(), pcaIssuer)
			if err != nil {
				log.Error(err, "failed to update externalSec when clean finalizers")
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}
	authConfig, err := p.IssuerManager.CreateAuthConfig(ctx, key, &pcaIssuer.Spec)
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

func (p *PCAIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	options := controller.Options{
		Reconciler: p,
	}
	PCAIssuerController, err := controller.New("PCAIssuer-controller", mgr, options)
	if err != nil {
		return err
	}
	err = PCAIssuerController.Watch(source.Kind(mgr.GetCache(), &v1beta.PCAIssuer{}, &handler.TypedEnqueueRequestForObject[*v1beta.PCAIssuer]{}, PCAIssuerPredicate[*v1beta.PCAIssuer]{}))
	if err != nil {
		return err
	}
	return nil
}
