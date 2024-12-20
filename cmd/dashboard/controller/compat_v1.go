// 为 v1 版本提供兼容接口
package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/naiba/nezha/model"
	"github.com/naiba/nezha/pkg/mygin"
	"github.com/naiba/nezha/pkg/utils"
	"github.com/naiba/nezha/service/singleton"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/sync/singleflight"
)

type compatV1 struct {
	r            gin.IRouter
	requestGroup singleflight.Group
}

func (cv *compatV1) serve() {
	r := cv.r.Group("")
	r.GET("/ws/server", cv.serverStream)
	r.GET("/server-group", cv.listServerGroup)

	r.GET("/service", cv.showService)
	r.GET("/service/:id", cv.listServiceHistory)
	r.GET("/service/server", cv.listServerWithServices)

	r.GET("/setting", cv.listConfig)
}

func (cv *compatV1) serverStream(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		mygin.ShowErrorPage(c, mygin.ErrInfo{
			Code: http.StatusInternalServerError,
			Title: singleton.Localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "NetworkError",
			}),
			Msg:  "Websocket协议切换失败",
			Link: "/",
			Btn:  "返回首页",
		}, true)
		return
	}
	defer conn.Close()
	count := 0
	for {
		stat, err := cv.getServerStat(c, count == 0)
		if err != nil {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, stat); err != nil {
			break
		}
		count += 1
		if count%4 == 0 {
			err = conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				break
			}
		}
		time.Sleep(time.Second * 2)
	}
}

func (cv *compatV1) getServerStat(c *gin.Context, withPublicNote bool) ([]byte, error) {
	_, isMember := c.Get(model.CtxKeyAuthorizedUser)
	_, isViewPasswordVerfied := c.Get(model.CtxKeyViewPasswordVerified)
	authorized := isMember || isViewPasswordVerfied
	v, err, _ := cv.requestGroup.Do(fmt.Sprintf("serverStats::%t", authorized), func() (interface{}, error) {
		singleton.SortedServerLock.RLock()
		defer singleton.SortedServerLock.RUnlock()

		var serverList []*model.Server
		if authorized {
			serverList = singleton.SortedServerList
		} else {
			serverList = singleton.SortedServerListForGuest
		}

		servers := make([]model.V1StreamServer, 0, len(serverList))
		for _, server := range serverList {
			servers = append(servers, model.V1StreamServer{
				ID:   server.ID,
				Name: server.Name,
				PublicNote: func() string {
					if withPublicNote {
						return server.PublicNote
					}
					return ""
				}(),
				DisplayIndex: server.DisplayIndex,
				Host: &model.V1Host{
					Platform:        server.Host.Platform,
					PlatformVersion: server.Host.PlatformVersion,
					CPU:             server.Host.CPU,
					MemTotal:        server.Host.MemTotal,
					DiskTotal:       server.Host.DiskTotal,
					SwapTotal:       server.Host.SwapTotal,
					Arch:            server.Host.Arch,
					Virtualization:  server.Host.Virtualization,
					BootTime:        server.Host.BootTime,
					Version:         server.Host.Version,
					GPU:             server.Host.GPU,
				},
				State: &model.V1HostState{
					CPU:            server.State.CPU,
					MemUsed:        server.State.MemUsed,
					SwapUsed:       server.State.SwapUsed,
					DiskUsed:       server.State.DiskUsed,
					NetInTransfer:  server.State.NetInTransfer,
					NetOutTransfer: server.State.NetOutTransfer,
					NetInSpeed:     server.State.NetInSpeed,
					NetOutSpeed:    server.State.NetOutSpeed,
					Uptime:         server.State.Uptime,
					Load1:          server.State.Load1,
					Load5:          server.State.Load5,
					Load15:         server.State.Load15,
					TcpConnCount:   server.State.TcpConnCount,
					UdpConnCount:   server.State.UdpConnCount,
					ProcessCount:   server.State.ProcessCount,
					Temperatures:   server.State.Temperatures,
					GPU:            []float64{server.State.GPU},
				},
				CountryCode: server.Host.CountryCode,
				LastActive:  server.LastActive,
			})
		}

		return utils.Json.Marshal(model.V1StreamServerData{
			Now:     time.Now().Unix() * 1000,
			Servers: servers,
		})
	})
	return v.([]byte), err
}

func (cv *compatV1) listServerGroup(c *gin.Context) {
	var sgRes []model.V1ServerGroupResponseItem

	tagID := uint64(0)
	for tag, ids := range singleton.ServerTagToIDList {
		sgRes = append(sgRes, model.V1ServerGroupResponseItem{
			Group: model.V1ServerGroup{
				V1Common: model.V1Common{
					ID:        tagID,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name: tag,
			},
			Servers: ids,
		})
		tagID++ // 虽然无法保证 tagID 的唯一性，但至少在绝大部分情况下不会出问题
	}

	c.JSON(200, sgRes)
}

func (cv *compatV1) showService(c *gin.Context) {
	res, err, _ := cv.requestGroup.Do("list-service", func() (interface{}, error) {
		singleton.AlertsLock.RLock()
		defer singleton.AlertsLock.RUnlock()

		sri := make(map[uint64]model.V1ServiceResponseItem)
		for k, service := range singleton.ServiceSentinelShared.LoadStats() {
			if !service.Monitor.EnableShowInService {
				continue
			}

			sri[k] = model.V1ServiceResponseItem{
				ServiceName: service.Monitor.Name,
				CurrentUp:   service.CurrentUp,
				CurrentDown: service.CurrentDown,
				TotalUp:     service.TotalUp,
				TotalDown:   service.TotalDown,
				Delay:       service.Delay,
				Up:          service.Up,
				Down:        service.Down,
			}
		}
		cycleTransferStats := make(map[uint64]model.V1CycleTransferStats)
		for k, v := range singleton.AlertsCycleTransferStatsStore {
			cycleTransferStats[k] = model.V1CycleTransferStats{
				Name:       v.Name,
				From:       v.From,
				To:         v.To,
				Max:        v.Max,
				Min:        v.Min,
				ServerName: v.ServerName,
				Transfer:   v.Transfer,
				NextUpdate: v.NextUpdate,
			}
		}
		return []interface {
		}{
			sri, cycleTransferStats,
		}, nil
	})
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "获取服务状态失败"})
		return
	}

	response := model.V1ServiceResponse{
		Services:           res.([]interface{})[0].(map[uint64]model.V1ServiceResponseItem),
		CycleTransferStats: res.([]interface{})[1].(map[uint64]model.V1CycleTransferStats),
	}
	c.JSON(200, response)
}

func (cv *compatV1) listServiceHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	singleton.ServerLock.RLock()
	server, ok := singleton.ServerList[id]
	if !ok {
		c.JSON(404, gin.H{"code": 404, "message": "server not found"})
		return
	}

	_, isMember := c.Get(model.CtxKeyAuthorizedUser)
	_, isViewPasswordVerfied := c.Get(model.CtxKeyViewPasswordVerified)
	authorized := isMember || isViewPasswordVerfied

	if server.HideForGuest && !authorized {
		c.JSON(403, gin.H{"code": 403, "message": "unauthorized"})
		return
	}
	singleton.ServerLock.RUnlock()

	query := map[string]any{"server_id": id}
	monitorHistories := singleton.MonitorAPI.GetMonitorHistories(query)
	if monitorHistories.Code != 0 {
		c.JSON(500, gin.H{"code": 500, "message": monitorHistories.Message})
		return
	}

	ret := make([]*model.V1ServiceInfos, 0, len(monitorHistories.Result))
	for _, history := range monitorHistories.Result {
		ret = append(ret, &model.V1ServiceInfos{
			ServiceID:   history.MonitorID,
			ServerID:    history.ServerID,
			ServiceName: singleton.ServiceSentinelShared.Monitors()[history.MonitorID].Name,
			ServerName:  singleton.ServerList[history.ServerID].Name,
			CreatedAt:   history.CreatedAt,
			AvgDelay:    history.AvgDelay,
		})
	}

	c.JSON(200, ret)
}

func (cv *compatV1) listServerWithServices(c *gin.Context) {
	var serverIdsWithService []uint64
	if err := singleton.DB.Model(&model.MonitorHistory{}).
		Select("distinct(server_id)").
		Where("server_id != 0").
		Find(&serverIdsWithService).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	_, isMember := c.Get(model.CtxKeyAuthorizedUser)
	_, isViewPasswordVerfied := c.Get(model.CtxKeyViewPasswordVerified)
	authorized := isMember || isViewPasswordVerfied

	var ret []uint64
	for _, id := range serverIdsWithService {
		singleton.ServerLock.RLock()
		server, ok := singleton.ServerList[id]
		if !ok {
			singleton.ServerLock.RUnlock()
			c.JSON(404, gin.H{"code": 404, "message": "server not found"})
			return
		}

		if !server.HideForGuest || authorized {
			ret = append(ret, id)
		}
		singleton.ServerLock.RUnlock()
	}

	c.JSON(200, ret)
}

func (cv *compatV1) listConfig(c *gin.Context) {
	_, isMember := c.Get(model.CtxKeyAuthorizedUser)
	_, isViewPasswordVerfied := c.Get(model.CtxKeyViewPasswordVerified)
	authorized := isMember || isViewPasswordVerfied

	conf := model.V1SettingResponse{
		V1Config: model.V1Config{
			SiteName:            singleton.Conf.Site.Brand,
			Language:            strings.Replace(singleton.Conf.Language, "_", "-", -1),
			CustomCode:          singleton.Conf.Site.CustomCode,
			CustomCodeDashboard: singleton.Conf.Site.CustomCodeDashboard,
		},
		Version: func() string {
			if authorized {
				return singleton.Version
			}
			return ""
		}(),
	}

	c.JSON(200, conf)
}
