package svr_mgr

import (
	"github.com/yuhao-jack/evolving-rpc/contents"
	"github.com/yuhao-jack/evolving-rpc/model"
	"github.com/yuhao-jack/go-toolx/containerx"
	"github.com/yuhao-jack/go-toolx/fun"
	"github.com/yuhao-jack/go-toolx/netx"
	"sync"
	"time"
)

// ServiceMgr
// @Description: 服务管理器，管理注册过来的服务
type ServiceMgr struct {
	ServiceInfoList containerx.ISet[*model.ServiceInfo]
	DataPackMap     *containerx.ConcurrentMap[string, *netx.DataPack]
	lock            sync.RWMutex
	keepDuration    time.Duration
}

var once sync.Once
var serviceMgrInstance = &ServiceMgr{
	ServiceInfoList: containerx.NewConcurrentSet[*model.ServiceInfo](),
	DataPackMap:     containerx.NewConcurrentMap[string, *netx.DataPack](),
	lock:            sync.RWMutex{},
}

// GetServiceMgrInstance
//
//	@Description:  获取一个单例的服务管理器
//	@return *ServiceMgr
func GetServiceMgrInstance() *ServiceMgr {
	once.Do(func() {
		go func() {
			ticker := time.NewTicker(time.Second)
			for range ticker.C {
				serviceMgrInstance.ServiceInfoList.ForEach(func(info *model.ServiceInfo) {
					lostTime, ok := info.AdditionalMeta[contents.LostTime.String()]
					if !ok {
						return
					}
					if time.Since(lostTime.(time.Time)) > serviceMgrInstance.GetKeepDuration() {
						serviceMgrInstance.ServiceInfoList.Remove(info)
					} else {
						info.AdditionalMeta[contents.Status.String()] = contents.Down
					}
				})
			}

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
	m.ServiceInfoList.ForEach(func(info *model.ServiceInfo) {
		if info.ServiceName == serviceName {
			serviceList = append(serviceList, info)
		}
	})
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
	m.ServiceInfoList.Add(serviceInfo)
}

// AddDataPack
//
//	@Description: 添加一个连接包
//	@receiver m
//	@param pack 连接包
func (m *ServiceMgr) AddDataPack(pack *netx.DataPack) {
	m.DataPackMap.Set(pack.RemoteAddr().String(), pack)
}

// DelDataPack
//
//	@Description: 删除一个连接包
//	@receiver m
//	@param pack
func (m *ServiceMgr) DelDataPack(pack *netx.DataPack) {
	m.DataPackMap.Remove(pack.RemoteAddr().String())
}

// PushConfig
//
//	@Description: 主动推送配置给客户
//	@author yuhao<154826195@qq.com>
//	@Data 2023-06-27 20:41:13
//	@receiver m
func (m *ServiceMgr) PushConfig() {

}
