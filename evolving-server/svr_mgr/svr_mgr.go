package svr_mgr

import (
	"evolving-rpc/model"
	"gitee.com/yuhao-jack/go-toolx/netx"
	"sync"
)

type ServiceMgr struct {
	ServiceInfoList []*model.ServiceInfo
	DataPackMap     map[string]*netx.DataPack
	lock            sync.RWMutex
}

var serviceMgrInstance = &ServiceMgr{
	ServiceInfoList: nil,
	lock:            sync.RWMutex{},
}

func GetServiceMgrInstance() *ServiceMgr {
	return serviceMgrInstance
}

func (m *ServiceMgr) FindServiceInfosByServiceName(serviceName string) (serviceList []*model.ServiceInfo) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, info := range m.ServiceInfoList {
		if info.ServiceName == serviceName {
			serviceList = append(serviceList, info)
		}
	}
	return serviceList
}

func (m *ServiceMgr) AddServiceInfo(serviceInfo *model.ServiceInfo) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.ServiceInfoList = append(m.ServiceInfoList, serviceInfo)
}

func (m *ServiceMgr) AddDataPack(pack *netx.DataPack) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.DataPackMap == nil {
		m.DataPackMap = make(map[string]*netx.DataPack)
	}
	m.DataPackMap[pack.RemoteAddr().String()] = pack
}

func (m *ServiceMgr) DelDataPack(pack *netx.DataPack) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.DataPackMap, pack.RemoteAddr().String())
}
