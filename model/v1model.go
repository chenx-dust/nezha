package model

import "time"

type V1Host struct {
	Platform        string   `json:"platform,omitempty"`
	PlatformVersion string   `json:"platform_version,omitempty"`
	CPU             []string `json:"cpu,omitempty"`
	MemTotal        uint64   `json:"mem_total,omitempty"`
	DiskTotal       uint64   `json:"disk_total,omitempty"`
	SwapTotal       uint64   `json:"swap_total,omitempty"`
	Arch            string   `json:"arch,omitempty"`
	Virtualization  string   `json:"virtualization,omitempty"`
	BootTime        uint64   `json:"boot_time,omitempty"`
	Version         string   `json:"version,omitempty"`
	GPU             []string `json:"gpu,omitempty"`
}

type V1HostState struct {
	CPU            float64             `json:"cpu,omitempty"`
	MemUsed        uint64              `json:"mem_used,omitempty"`
	SwapUsed       uint64              `json:"swap_used,omitempty"`
	DiskUsed       uint64              `json:"disk_used,omitempty"`
	NetInTransfer  uint64              `json:"net_in_transfer,omitempty"`
	NetOutTransfer uint64              `json:"net_out_transfer,omitempty"`
	NetInSpeed     uint64              `json:"net_in_speed,omitempty"`
	NetOutSpeed    uint64              `json:"net_out_speed,omitempty"`
	Uptime         uint64              `json:"uptime,omitempty"`
	Load1          float64             `json:"load_1,omitempty"`
	Load5          float64             `json:"load_5,omitempty"`
	Load15         float64             `json:"load_15,omitempty"`
	TcpConnCount   uint64              `json:"tcp_conn_count,omitempty"`
	UdpConnCount   uint64              `json:"udp_conn_count,omitempty"`
	ProcessCount   uint64              `json:"process_count,omitempty"`
	Temperatures   []SensorTemperature `json:"temperatures,omitempty"`
	GPU            []float64           `json:"gpu,omitempty"`
}

type V1StreamServer struct {
	ID           uint64 `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	PublicNote   string `json:"public_note,omitempty"`   // 公开备注，只第一个数据包有值
	DisplayIndex int    `json:"display_index,omitempty"` // 展示排序，越大越靠前

	Host        *V1Host      `json:"host,omitempty"`
	State       *V1HostState `json:"state,omitempty"`
	CountryCode string       `json:"country_code,omitempty"`
	LastActive  time.Time    `json:"last_active,omitempty"`
}

type V1StreamServerData struct {
	Now     int64            `json:"now,omitempty"`
	Online  uint64           `json:"online,omitempty"`
	Servers []V1StreamServer `json:"servers,omitempty"`
}

type V1Common struct {
	ID        uint64    `gorm:"primaryKey" json:"id,omitempty"`
	CreatedAt time.Time `gorm:"index;<-:create" json:"created_at,omitempty"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at,omitempty"`
	// Do not use soft deletion
	// DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type V1ServerGroup struct {
	V1Common

	Name string `json:"name"`
}

type V1ServerGroupResponseItem struct {
	Group   V1ServerGroup `json:"group"`
	Servers []uint64      `json:"servers"`
}

type V1ServiceResponseItem struct {
	ServiceName string       `json:"service_name,omitempty"`
	CurrentUp   uint64       `json:"current_up"`
	CurrentDown uint64       `json:"current_down"`
	TotalUp     uint64       `json:"total_up"`
	TotalDown   uint64       `json:"total_down"`
	Delay       *[30]float32 `json:"delay,omitempty"`
	Up          *[30]int     `json:"up,omitempty"`
	Down        *[30]int     `json:"down,omitempty"`
}

type V1CycleTransferStats struct {
	Name       string               `json:"name"`
	From       time.Time            `json:"from"`
	To         time.Time            `json:"to"`
	Max        uint64               `json:"max"`
	Min        uint64               `json:"min"`
	ServerName map[uint64]string    `json:"server_name,omitempty"`
	Transfer   map[uint64]uint64    `json:"transfer,omitempty"`
	NextUpdate map[uint64]time.Time `json:"next_update,omitempty"`
}

type V1ServiceResponse struct {
	Services           map[uint64]V1ServiceResponseItem `json:"services,omitempty"`
	CycleTransferStats map[uint64]V1CycleTransferStats  `json:"cycle_transfer_stats,omitempty"`
}

type V1ServiceInfos struct {
	ServiceID   uint64    `json:"monitor_id"`
	ServerID    uint64    `json:"server_id"`
	ServiceName string    `json:"monitor_name"`
	ServerName  string    `json:"server_name"`
	CreatedAt   []int64   `json:"created_at"`
	AvgDelay    []float32 `json:"avg_delay"`
}

type V1Config struct {
	SiteName            string `mapstructure:"site_name" json:"site_name"`
	Language            string `mapstructure:"language" json:"language"`
	CustomCode          string `mapstructure:"custom_code" json:"custom_code,omitempty"`
	CustomCodeDashboard string `mapstructure:"custom_code_dashboard" json:"custom_code_dashboard,omitempty"`
}

type V1SettingResponse[T any] struct {
	Config T `json:"config,omitempty"`

	Version string `json:"version,omitempty"`
}

type V1User struct {
	V1Common
	Username string `json:"username,omitempty" gorm:"uniqueIndex"`
	Password string `json:"password,omitempty" gorm:"type:char(72)"`
}

type V1Profile struct {
	V1User
	LoginIP string `json:"login_ip,omitempty"`
}

type V1LoginRequest struct {
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
}

type V1LoginResponse struct {
	Expire string `json:"expire,omitempty"`
	Token  string `json:"token,omitempty"`
}
