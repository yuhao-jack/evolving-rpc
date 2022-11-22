package svr_mgr

import (
	"github.com/yuhao-jack/evolving-rpc/contents"
	"github.com/yuhao-jack/evolving-rpc/model"
	"github.com/yuhao-jack/go-toolx/fun"
	"github.com/yuhao-jack/go-toolx/netx"
	"sync"
	"time"
)

// ServiceMgr
// @Description: 服务管理器，管理注册过来的服务
type ServiceMgr struct {
	ServiceInfoList []*model.ServiceInfo
	DataPackMap     map[string]*netx.DataPack
	lock            sync.RWMutex
	keepDuration    time.Duration
}

var once sync.Once
var serviceMgrInstance = &ServiceMgr{
	ServiceInfoList: nil,
	lock:            sync.RWMutex{},
}

// GetServiceMgrInstance
//
//	@Description:  获取一个单例的服务管理器
//	@return *ServiceMgr
func GetServiceMgrInstance() *ServiceMgr {
	once.Do(func() {
		go func() {
			time.AfterFunc(time.Second, func() {
				for i, info := range serviceMgrInstance.ServiceInfoList {
					lostTime, ok := info.AdditionalMeta[contents.LostTime.String()]
					if ok {
						serviceMgrInstance.lock.Lock()
						if time.Since(lostTime.(time.Time)) > serviceMgrInstance.GetKeepDuration() {
							serviceMgrInstance.ServiceInfoList = append(serviceMgrInstance.ServiceInfoList[:i], serviceMgrInstance.ServiceInfoList[i+1:]...)
						} else {
							info.AdditionalMeta[contents.Status.String()] = contents.Down
						}
						serviceMgrInstance.lock.Unlock()
					}
				}
			})

		}()
	})
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

// SetKeepDuration
//
//	@Description: 设置注册中心保存服务信息的时间
//	@receiver m
//	@param duration　时间
func (m *ServiceMgr) SetKeepDuration(duration time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.keepDuration = duration
}

// GetKeepDuration
//
//	@Description:　获取注册中心保存服务信息的时间
//	@receiver m
//	@return time.Duration　时间
func (m *ServiceMgr) GetKeepDuration() time.Duration {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return fun.IfOr(serviceMgrInstance.keepDuration == 0, time.Second*30, serviceMgrInstance.keepDuration)
}

// AddServiceInfo
//
//	@Description: 添加服务信息
//	@receiver m
//	@param serviceInfo 服务信息
func (m *ServiceMgr) AddServiceInfo(serviceInfo *model.ServiceInfo) {
	m.lock.Lock()
	defer m.lock.Unlock()
	serviceInfo.AdditionalMeta[contents.Status.String()] = contents.Up
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
