package main

import (
	"encoding/json"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

// ============================================================
// 数据库模型
// ============================================================

// FlowConfig 流程配置
type FlowConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Stages      string    `gorm:"type:text" json:"stages"` // JSON 数组
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StageConfig 阶段配置（用于 JSON 序列化）
type StageConfig struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Type     string `json:"type"` // approval/test/prepare
	Enabled  bool   `json:"enabled"`
	Timeout  int    `json:"timeout"`   // 超时时间（小时）
	AutoPass bool   `json:"auto_pass"` // 超时是否自动通过
	Order    int    `json:"order"`
}

// ItemModel 条目模型
type ItemModel struct {
	ID            string    `gorm:"primaryKey;size:50" json:"id"`
	Name          string    `gorm:"size:200;not null" json:"name"`
	Type          string    `gorm:"size:50" json:"type"`
	RequirementID string    `gorm:"size:100" json:"requirement_id"`
	Developer     string    `gorm:"size:100" json:"developer"`
	Tester        string    `gorm:"size:100" json:"tester"`
	ItemOwner     string    `gorm:"size:100" json:"item_owner"`
	Status        string    `gorm:"size:50" json:"status"`
	HasScript     bool      `json:"has_script"`
	HasCache      bool      `json:"has_cache"`
	HasRestart    bool      `json:"has_restart"`
	BTEResult     string    `gorm:"size:50" json:"bte_result"`
	GrayResult    string    `gorm:"size:50" json:"gray_result"`
	ProdResult    string    `gorm:"size:50" json:"prod_result"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (ItemModel) TableName() string { return "upgrade_items" }

// VersionModel 版本模型
type VersionModel struct {
	ID           string    `gorm:"primaryKey;size:50" json:"id"`
	Name         string    `gorm:"size:200;not null" json:"name"`
	VersionOwner string    `gorm:"size:100" json:"version_owner"`
	VendorOwner  string    `gorm:"size:100" json:"vendor_owner"`
	BTETester    string    `gorm:"size:100" json:"bte_tester"`
	GrayTester   string    `gorm:"size:100" json:"gray_tester"`
	ProdTester   string    `gorm:"size:100" json:"prod_tester"`
	IsUrgent     bool      `json:"is_urgent"`
	Status       string    `gorm:"size:50" json:"status"`
	CurrentStage string    `gorm:"size:50" json:"current_stage"`
	ItemIDs      string    `gorm:"type:text" json:"item_ids"` // JSON 数组
	FlowConfigID uint      `json:"flow_config_id"`
	WorkflowID   string    `gorm:"size:100" json:"workflow_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (VersionModel) TableName() string { return "upgrade_versions" }

// ============================================================
// 数据库初始化
// ============================================================

func initDatabase() error {
	dsn := "root:Nuttertools1103..@tcp(127.0.0.1:3306)/upgrade_workflow?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移
	err = db.AutoMigrate(&FlowConfig{}, &ItemModel{}, &VersionModel{})
	if err != nil {
		return err
	}

	// 初始化默认流程配置
	initDefaultFlowConfig()

	logger.Info("数据库连接成功")
	return nil
}

// initDefaultFlowConfig 初始化默认流程配置
func initDefaultFlowConfig() {
	var count int64
	db.Model(&FlowConfig{}).Count(&count)
	if count > 0 {
		return
	}

	// 默认完整流程
	defaultStages := []StageConfig{
		{Key: StageBTEConfirm, Name: "BTE条目确认", Type: "approval", Enabled: true, Timeout: 72, Order: 1},
		{Key: StageBTEFinalize, Name: "BTE定版", Type: "approval", Enabled: true, Timeout: 48, Order: 2},
		{Key: StageBTEPrepare, Name: "BTE版本准备", Type: "prepare", Enabled: true, Timeout: 24, Order: 3},
		{Key: StageBTETest, Name: "BTE测试", Type: "test", Enabled: true, Timeout: 96, Order: 4},
		{Key: StageGrayConfirm, Name: "灰度条目确认", Type: "approval", Enabled: true, Timeout: 48, Order: 5},
		{Key: StageGrayFinalize, Name: "灰度定版", Type: "approval", Enabled: true, Timeout: 24, Order: 6},
		{Key: StageGrayPrepare, Name: "灰度版本准备", Type: "prepare", Enabled: true, Timeout: 24, Order: 7},
		{Key: StageGrayTest, Name: "灰度测试", Type: "test", Enabled: true, Timeout: 96, Order: 8},
		{Key: StageProdFinalize, Name: "生产定版", Type: "approval", Enabled: true, Timeout: 48, Order: 9},
		{Key: StageProdPrepare, Name: "生产版本准备", Type: "prepare", Enabled: true, Timeout: 24, Order: 10},
		{Key: StageProdTest, Name: "生产测试", Type: "test", Enabled: true, Timeout: 96, Order: 11},
		{Key: StageCloseConfirm, Name: "关闭确认", Type: "approval", Enabled: true, Timeout: 72, AutoPass: true, Order: 12},
		{Key: StageEndConfirm, Name: "结束确认", Type: "approval", Enabled: true, Timeout: 48, AutoPass: true, Order: 13},
	}

	stagesJSON, _ := json.Marshal(defaultStages)
	defaultConfig := FlowConfig{
		Name:        "默认升级流程",
		Description: "包含完整的 BTE → 灰度 → 生产 测试流程",
		Stages:      string(stagesJSON),
		IsDefault:   true,
	}
	db.Create(&defaultConfig)

	// 简化流程（跳过灰度）
	simpleStages := []StageConfig{
		{Key: StageBTEConfirm, Name: "BTE条目确认", Type: "approval", Enabled: true, Timeout: 72, Order: 1},
		{Key: StageBTEFinalize, Name: "BTE定版", Type: "approval", Enabled: true, Timeout: 48, Order: 2},
		{Key: StageBTEPrepare, Name: "BTE版本准备", Type: "prepare", Enabled: true, Timeout: 24, Order: 3},
		{Key: StageBTETest, Name: "BTE测试", Type: "test", Enabled: true, Timeout: 96, Order: 4},
		{Key: StageProdFinalize, Name: "生产定版", Type: "approval", Enabled: true, Timeout: 48, Order: 5},
		{Key: StageProdPrepare, Name: "生产版本准备", Type: "prepare", Enabled: true, Timeout: 24, Order: 6},
		{Key: StageProdTest, Name: "生产测试", Type: "test", Enabled: true, Timeout: 96, Order: 7},
		{Key: StageCloseConfirm, Name: "关闭确认", Type: "approval", Enabled: true, Timeout: 72, AutoPass: true, Order: 8},
		{Key: StageEndConfirm, Name: "结束确认", Type: "approval", Enabled: true, Timeout: 48, AutoPass: true, Order: 9},
	}
	simpleJSON, _ := json.Marshal(simpleStages)
	simpleConfig := FlowConfig{
		Name:        "紧急升级流程",
		Description: "跳过灰度测试，直接进入生产",
		Stages:      string(simpleJSON),
		IsDefault:   false,
	}
	db.Create(&simpleConfig)

	logger.Info("默认流程配置已初始化")
}

// ============================================================
// 数据库操作
// ============================================================

// GetFlowConfigs 获取所有流程配置
func GetFlowConfigs() ([]FlowConfig, error) {
	var configs []FlowConfig
	err := db.Find(&configs).Error
	return configs, err
}

// GetFlowConfig 获取单个流程配置
func GetFlowConfig(id uint) (*FlowConfig, error) {
	var config FlowConfig
	err := db.First(&config, id).Error
	return &config, err
}

// CreateFlowConfig 创建流程配置
func CreateFlowConfig(config *FlowConfig) error {
	return db.Create(config).Error
}

// UpdateFlowConfig 更新流程配置
func UpdateFlowConfig(config *FlowConfig) error {
	return db.Save(config).Error
}

// DeleteFlowConfig 删除流程配置
func DeleteFlowConfig(id uint) error {
	return db.Delete(&FlowConfig{}, id).Error
}

// GetFlowStages 解析流程阶段配置
func GetFlowStages(config *FlowConfig) ([]StageConfig, error) {
	var stages []StageConfig
	err := json.Unmarshal([]byte(config.Stages), &stages)
	return stages, err
}

// ============================================================
// 条目数据库操作
// ============================================================

func GetAllItems() ([]ItemModel, error) {
	var items []ItemModel
	err := db.Find(&items).Error
	return items, err
}

func GetItemByID(id string) (*ItemModel, error) {
	var item ItemModel
	err := db.First(&item, "id = ?", id).Error
	return &item, err
}

func CreateItem(item *ItemModel) error {
	return db.Create(item).Error
}

func UpdateItem(item *ItemModel) error {
	return db.Save(item).Error
}

// ============================================================
// 版本数据库操作
// ============================================================

func GetAllVersions() ([]VersionModel, error) {
	var versions []VersionModel
	err := db.Order("created_at desc").Find(&versions).Error
	return versions, err
}

func GetVersionByID(id string) (*VersionModel, error) {
	var version VersionModel
	err := db.First(&version, "id = ?", id).Error
	return &version, err
}

func CreateVersion(version *VersionModel) error {
	return db.Create(version).Error
}

func UpdateVersion(version *VersionModel) error {
	return db.Save(version).Error
}

// 生成条目ID
func GenerateItemID() string {
	var count int64
	db.Model(&ItemModel{}).Count(&count)
	return "ITEM-" + time.Now().Format("20060102") + "-" + padNumber(int(count)+1, 4)
}

// 生成版本ID
func GenerateVersionID() string {
	var count int64
	db.Model(&VersionModel{}).Count(&count)
	return "V" + time.Now().Format("200601") + "-" + padNumber(int(count)+1, 3)
}

func padNumber(n, width int) string {
	s := ""
	for i := 0; i < width; i++ {
		s = string('0'+byte(n%10)) + s
		n /= 10
	}
	return s
}

// InitDemoData 初始化演示数据
func InitDemoData() {
	var count int64
	db.Model(&ItemModel{}).Count(&count)
	if count > 0 {
		return
	}

	demoItems := []ItemModel{
		{
			ID:        "ITEM-1001",
			Name:      "CPC同步offer表状态接口",
			Type:      "需求",
			Developer: "刘奕玄",
			Tester:    "潘燕燕",
			ItemOwner: "秦雪",
			Status:    ItemStatusAuditComplete,
			HasScript: true,
			BTEResult: TestResultPending,
		},
		{
			ID:        "ITEM-1002",
			Name:      "客户标签查询优化",
			Type:      "优化",
			Developer: "郑维忠",
			Tester:    "潘燕燕",
			ItemOwner: "秦雪",
			Status:    ItemStatusAuditComplete,
			HasCache:  true,
			BTEResult: TestResultPending,
		},
		{
			ID:        "ITEM-1003",
			Name:      "受理单打印功能修复",
			Type:      "BUG",
			Developer: "刘奕玄",
			Tester:    "潘燕燕",
			ItemOwner: "秦雪",
			Status:    ItemStatusAuditComplete,
			BTEResult: TestResultPending,
		},
	}

	for _, item := range demoItems {
		db.Create(&item)
	}

	logger.Info("演示数据初始化完成")
}
