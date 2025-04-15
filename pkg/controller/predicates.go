package controller

import (
	"github.com/AliyunContainerService/alibabacloud-privateca-issuer/api/v1beta"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type PCAClusterIssuerPredicate[object any] struct{}

func (p PCAClusterIssuerPredicate[object]) Create(e event.TypedCreateEvent[object]) bool {
	return true
}

func (p PCAClusterIssuerPredicate[object]) Delete(e event.TypedDeleteEvent[object]) bool {
	return true
}

func (p PCAClusterIssuerPredicate[object]) Update(e event.TypedUpdateEvent[object]) bool {
	var oldObjInterface interface{} = e.ObjectOld
	var newObjInterface interface{} = e.ObjectNew
	oldObj, ok := oldObjInterface.(*v1beta.PCAClusterIssuer)
	if !ok {
		return false
	}
	newObj, ok := newObjInterface.(*v1beta.PCAClusterIssuer)
	if !ok {
		return false
	}
	if !reflect.DeepEqual(oldObj.Spec, newObj.Spec) || !reflect.DeepEqual(oldObj.Status, newObj.Status) ||
		oldObj.GetDeletionTimestamp() != newObj.GetDeletionTimestamp() ||
		oldObj.GetGeneration() != newObj.GetGeneration() {
		return true
	}
	return false
}

func (p PCAClusterIssuerPredicate[object]) Generic(e event.TypedGenericEvent[object]) bool {
	return true
}

type PCAIssuerPredicate[object any] struct{}

func (p PCAIssuerPredicate[object]) Create(e event.TypedCreateEvent[object]) bool {
	return true
}

func (p PCAIssuerPredicate[object]) Delete(e event.TypedDeleteEvent[object]) bool {
	return true
}

func (p PCAIssuerPredicate[object]) Update(e event.TypedUpdateEvent[object]) bool {
	var oldObjInterface interface{} = e.ObjectOld
	var newObjInterface interface{} = e.ObjectNew
	oldObj, ok := oldObjInterface.(*v1beta.PCAIssuer)
	if !ok {
		return false
	}
	newObj, ok := newObjInterface.(*v1beta.PCAIssuer)
	if !ok {
		return false
	}
	if !reflect.DeepEqual(oldObj.Spec, newObj.Spec) || !reflect.DeepEqual(oldObj.Status, newObj.Status) ||
		oldObj.GetDeletionTimestamp() != newObj.GetDeletionTimestamp() ||
		oldObj.GetGeneration() != newObj.GetGeneration() {
		return true
	}
	return false
}

func (p PCAIssuerPredicate[object]) Generic(e event.TypedGenericEvent[object]) bool {
	return true
}
