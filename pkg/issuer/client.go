package issuer

import (
	"context"
	"fmt"
	"github.com/AliyunContainerService/ack-ram-tool/pkg/credentials/provider"
	cas20200630 "github.com/alibabacloud-go/cas-20200630/client"
	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"time"
)

type IssuerManager struct {
	Region             string
	MaxConcurrentCount int
	ClientMap          sync.Map
	RamLock            *sync.Mutex
	RamProvider        map[string]provider.Stopper
	KubeClient         client.Client
	log                logr.Logger
	signingLimiter     *rate.Limiter
}

func NewIssuerManager(kubeClient client.Client, region string, maxConcurrentCount int) *IssuerManager {
	return &IssuerManager{
		signingLimiter:     rate.NewLimiter(rate.Limit(maxConcurrentCount), 1),
		MaxConcurrentCount: maxConcurrentCount,
		Region:             region,
		RamLock:            &sync.Mutex{},
		RamProvider:        make(map[string]provider.Stopper),
		ClientMap:          sync.Map{},
		KubeClient:         kubeClient,
		log:                ctrl.Log.WithName("issuerManager"),
	}

}

func (m *IssuerManager) RegisterRamProvider(clientName string, stopper provider.Stopper) {
	if m == nil || m.RamLock == nil {
		klog.Errorf("Manager init error")
		return
	}
	m.RamLock.Lock()
	defer m.RamLock.Unlock()
	providerIns, ok := m.RamProvider[clientName]
	if ok {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		// cancel is earlier than m.RamLock.Unlock
		defer cancel()
		providerIns.Stop(timeoutCtx)
	}
	m.RamProvider[clientName] = stopper
	klog.Infof("register provider %v success", clientName)
}

func (m *IssuerManager) StopProvider(clientName string) {
	if m == nil || m.RamLock == nil {
		klog.Errorf("Manager init error")
		return
	}
	m.RamLock.Lock()
	defer m.RamLock.Unlock()
	providerIns, ok := m.RamProvider[clientName]
	if !ok || providerIns == nil {
		return
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	// cancel is earlier than m.RamLock.Unlock
	defer cancel()
	providerIns.Stop(timeoutCtx)
	delete(m.RamProvider, clientName)
	klog.Infof("stop provider %v success", clientName)
}

func (m *IssuerManager) Register(clientName string, client interface{}) {
	casClient, ok := client.(*cas20200630.Client)
	if casClient == nil {
		klog.Errorf("client %s is nil", clientName)
		return
	}
	if !ok {
		klog.Errorf("client %s type error", clientName)
		return
	}
	m.ClientMap.Store(clientName, client)
	klog.Infof("register or update client, clientName %v", clientName)
}

func (m *IssuerManager) Delete(clientName string) {
	// delete the client map, and stop the ram provider refresh go routine
	m.ClientMap.Delete(clientName)
	m.StopProvider(clientName)
	klog.Infof("delete client, clientName %v", clientName)
}

func (m *IssuerManager) GetClient(clientName string) (*cas20200630.Client, error) {
	client, ok := m.ClientMap.Load(clientName)
	if ok {
		casClient, ok := client.(*cas20200630.Client)
		if !ok {
			return nil, fmt.Errorf("client type error, clientName %v", clientName)
		}
		return casClient, nil
	}
	return nil, nil
}
