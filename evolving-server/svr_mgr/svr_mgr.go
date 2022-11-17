package svr_mgr

import (
	"github.com/yuhao-jack/evolving-rpc/model"
	"github.com/yuhao-jack/go-toolx/netx"
	"sync"
)

// ServiceMgr
// @Description: 服务管理器，管理注册过来的服务
type ServiceMgr struct {
	ServiceInfoList []*model.ServiceInfo
	DataPackMap     map[string]*netx.DataPack
	lock            sync.RWMutex
}

var serviceMgrInstance = &ServiceMgr{
	ServiceInfoList: nil,
	lock:            sync.RWMutex{},
}

// GetServiceMgrInstance
//
//	@Description:  获取一个单例的服务管理器
//	@return *ServiceMgr
func GetServiceMgrInstance() *ServiceMgr {
	return serviceMgrInstance
}

// FindServiceInfosByServiceName
//
//	@Description: 通过服务的名字获取所有可用的服务的信息
//	@receiver m
//	@param serviceName 服务名
//	@return serviceList 所有可用的服务的信息
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

// AddServiceInfo
//
//	@Description: 添加服务信息
//	@receiver m
//	@param serviceInfo 服务信息
func (m *ServiceMgr) AddServiceInfo(serviceInfo *model.ServiceInfo) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.ServiceInfoList = append(m.ServiceInfoList, serviceInfo)
}

// AddDataPack
//
//	@Description: 添加一个连接包
//	@receiver m
//	@param pack 连接包
func (m *ServiceMgr) AddDataPack(pack *netx.DataPack) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.DataPackMap == nil {
		m.DataPackMap = make(map[string]*netx.DataPack)
	}
	m.DataPackMap[pack.RemoteAddr().String()] = pack
}

// DelDataPack
//
//	@Description: 删除一个连接包
//	@receiver m
//	@param pack
func (m *ServiceMgr) DelDataPack(pack *netx.DataPack) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.DataPackMap, pack.RemoteAddr().String())
}
