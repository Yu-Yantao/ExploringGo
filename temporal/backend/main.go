package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

var (
	temporalClient client.Client
	logger         *zap.Logger
)

func main() {
	// 初始化 Zap 日志
	var err error
	logger, err = zap.NewDevelopment()
	logger = logger.WithOptions(zap.AddCaller())
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// 初始化数据库
	if err := initDatabase(); err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}

	// 初始化演示数据
	InitDemoData()

	// 连接 Temporal Server
	temporalClient, err = client.Dial(client.Options{
		HostPort: "127.0.0.1:7233",
	})
	if err != nil {
		logger.Fatal("无法连接 Temporal", zap.Error(err))
	}
	defer temporalClient.Close()

	// 启动 Worker
	go StartWorker(temporalClient)

	// 设置 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	// 条目管理 API
	r.GET("/api/items", listItems)
	r.POST("/api/items", createItem)
	r.GET("/api/items/:itemId", getItem)

	// 版本管理 API
	r.GET("/api/versions", listVersions)
	r.POST("/api/versions", createVersion)
	r.GET("/api/versions/:versionId/status", getVersionStatus)

	// 流程配置 API
	r.GET("/api/flow-configs", listFlowConfigs)
	r.POST("/api/flow-configs", createFlowConfigHandler)
	r.GET("/api/flow-configs/:id", getFlowConfigHandler)
	r.PUT("/api/flow-configs/:id", updateFlowConfigHandler)
	r.DELETE("/api/flow-configs/:id", deleteFlowConfigHandler)

	// 流程操作 API
	r.POST("/api/workflow/:stage/approve", submitApproval)
	r.POST("/api/workflow/:stage/test", submitTestResult)

	// 静态文件
	r.NoRoute(func(c *gin.Context) {
		c.File("../frontend/index.html")
	})

	logger.Info("========================================")
	logger.Info("BSS3.0 升级流程管理系统")
	logger.Info("========================================")
	logger.Info("后端服务启动", zap.String("addr", "http://localhost:8082"))

	if err := r.Run(":8082"); err != nil {
		logger.Fatal("服务启动失败", zap.Error(err))
	}
}

// ============================================================
// 条目管理
// ============================================================

func listItems(c *gin.Context) {
	items, err := GetAllItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func createItem(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := ItemModel{
		ID:            GenerateItemID(),
		Name:          req.Name,
		Type:          req.Type,
		RequirementID: req.RequirementID,
		Developer:     req.Developer,
		Tester:        req.Tester,
		ItemOwner:     req.ItemOwner,
		Status:        ItemStatusRegistered,
		HasScript:     req.HasScript,
		HasCache:      req.HasCache,
		HasRestart:    req.HasRestart,
		BTEResult:     TestResultPending,
		GrayResult:    TestResultPending,
		ProdResult:    TestResultPending,
	}

	if err := CreateItem(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("条目已创建", zap.String("id", item.ID), zap.String("name", req.Name))
	c.JSON(http.StatusOK, item)
}

func getItem(c *gin.Context) {
	itemID := c.Param("itemId")
	item, err := GetItemByID(itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "条目不存在"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// ============================================================
// 版本管理
// ============================================================

func listVersions(c *gin.Context) {
	versions, err := GetAllVersions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, versions)
}

func createVersion(c *gin.Context) {
	var req struct {
		Name         string   `json:"name"`
		VersionOwner string   `json:"version_owner"`
		VendorOwner  string   `json:"vendor_owner"`
		BTETester    string   `json:"bte_tester"`
		GrayTester   string   `json:"gray_tester"`
		ProdTester   string   `json:"prod_tester"`
		IsUrgent     bool     `json:"is_urgent"`
		ItemIDs      []string `json:"item_ids"`
		FlowConfigID uint     `json:"flow_config_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取流程配置
	var flowConfig *FlowConfig
	var err error
	if req.FlowConfigID > 0 {
		flowConfig, err = GetFlowConfig(req.FlowConfigID)
	} else {
		// 使用默认配置
		var configs []FlowConfig
		db.Where("is_default = ?", true).First(&configs)
		if len(configs) > 0 {
			flowConfig = &configs[0]
		}
	}
	if err != nil || flowConfig == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "流程配置不存在"})
		return
	}

	// 获取流程阶段
	stages, _ := GetFlowStages(flowConfig)
	var firstStage string
	for _, s := range stages {
		if s.Enabled {
			firstStage = s.Key
			break
		}
	}

	versionID := GenerateVersionID()
	itemIDsJSON, _ := json.Marshal(req.ItemIDs)

	version := VersionModel{
		ID:           versionID,
		Name:         req.Name,
		VersionOwner: req.VersionOwner,
		VendorOwner:  req.VendorOwner,
		BTETester:    req.BTETester,
		GrayTester:   req.GrayTester,
		ProdTester:   req.ProdTester,
		IsUrgent:     req.IsUrgent,
		Status:       "running",
		CurrentStage: firstStage,
		ItemIDs:      string(itemIDsJSON),
		FlowConfigID: flowConfig.ID,
	}

	if err := CreateVersion(&version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 收集条目信息
	var itemList []UpgradeItem
	for _, itemID := range req.ItemIDs {
		item, _ := GetItemByID(itemID)
		if item != nil {
			itemList = append(itemList, UpgradeItem{
				ID:        item.ID,
				Name:      item.Name,
				Type:      item.Type,
				Developer: item.Developer,
				Tester:    item.Tester,
				ItemOwner: item.ItemOwner,
				Status:    item.Status,
			})
		}
	}

	// 启动 Temporal Workflow
	workflowID := "upgrade-" + versionID
	we, err := temporalClient.ExecuteWorkflow(
		c.Request.Context(),
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: TaskQueue,
		},
		UpgradeWorkflow,
		UpgradeWorkflowRequest{
			Version: UpgradeVersion{
				ID:           versionID,
				Name:         req.Name,
				VersionOwner: req.VersionOwner,
				IsUrgent:     req.IsUrgent,
				CurrentStage: firstStage,
				ItemIDs:      req.ItemIDs,
			},
			Items:        itemList,
			FlowConfigID: flowConfig.ID,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新 workflowID
	version.WorkflowID = workflowID
	UpdateVersion(&version)

	logger.Info("升级版本已创建",
		zap.String("versionId", versionID),
		zap.String("workflowId", workflowID),
		zap.Uint("flowConfigId", flowConfig.ID))

	c.JSON(http.StatusOK, gin.H{
		"workflow_id": we.GetID(),
		"version_id":  versionID,
		"run_id":      we.GetRunID(),
	})
}

func getVersionStatus(c *gin.Context) {
	versionID := c.Param("versionId")

	version, err := GetVersionByID(versionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "版本不存在"})
		return
	}

	// 解析条目 IDs
	var itemIDs []string
	json.Unmarshal([]byte(version.ItemIDs), &itemIDs)

	// 收集条目信息
	var itemList []ItemModel
	for _, itemID := range itemIDs {
		item, _ := GetItemByID(itemID)
		if item != nil {
			itemList = append(itemList, *item)
		}
	}

	// 获取流程配置
	flowConfig, _ := GetFlowConfig(version.FlowConfigID)
	var stages []StageConfig
	if flowConfig != nil {
		stages, _ = GetFlowStages(flowConfig)
	}

	// 查询 Temporal 工作流状态
	workflowID := "upgrade-" + versionID
	desc, err := temporalClient.DescribeWorkflowExecution(c.Request.Context(), workflowID, "")
	status := version.Status
	if err == nil {
		status = desc.WorkflowExecutionInfo.Status.String()
	}

	c.JSON(http.StatusOK, gin.H{
		"version_id":    versionID,
		"version_name":  version.Name,
		"status":        status,
		"current_stage": version.CurrentStage,
		"items":         itemList,
		"timeline":      buildTimelineFromConfig(stages, version.CurrentStage),
	})
}

func buildTimelineFromConfig(stages []StageConfig, currentStage string) []gin.H {
	var timeline []gin.H
	currentFound := false

	for _, stage := range stages {
		if !stage.Enabled {
			continue
		}

		status := "pending"
		if stage.Key == currentStage {
			status = "in_progress"
			currentFound = true
		} else if !currentFound {
			status = "completed"
		}

		timeline = append(timeline, gin.H{
			"stage":  stage.Name,
			"key":    stage.Key,
			"status": status,
		})
	}

	return timeline
}

// ============================================================
// 流程配置 API
// ============================================================

func listFlowConfigs(c *gin.Context) {
	configs, err := GetFlowConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 解析每个配置的阶段
	var result []gin.H
	for _, config := range configs {
		stages, _ := GetFlowStages(&config)
		result = append(result, gin.H{
			"id":          config.ID,
			"name":        config.Name,
			"description": config.Description,
			"stages":      stages,
			"is_default":  config.IsDefault,
			"created_at":  config.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

func getFlowConfigHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	config, err := GetFlowConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
		return
	}

	stages, _ := GetFlowStages(config)
	c.JSON(http.StatusOK, gin.H{
		"id":          config.ID,
		"name":        config.Name,
		"description": config.Description,
		"stages":      stages,
		"is_default":  config.IsDefault,
	})
}

func createFlowConfigHandler(c *gin.Context) {
	var req struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Stages      []StageConfig `json:"stages"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stagesJSON, _ := json.Marshal(req.Stages)
	config := FlowConfig{
		Name:        req.Name,
		Description: req.Description,
		Stages:      string(stagesJSON),
	}

	if err := CreateFlowConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("流程配置已创建", zap.Uint("id", config.ID), zap.String("name", config.Name))
	c.JSON(http.StatusOK, config)
}

func updateFlowConfigHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	config, err := GetFlowConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
		return
	}

	var req struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Stages      []StageConfig `json:"stages"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stagesJSON, _ := json.Marshal(req.Stages)
	config.Name = req.Name
	config.Description = req.Description
	config.Stages = string(stagesJSON)

	if err := UpdateFlowConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("流程配置已更新", zap.Uint("id", config.ID))
	c.JSON(http.StatusOK, config)
}

func deleteFlowConfigHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	config, err := GetFlowConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
		return
	}

	if config.IsDefault {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除默认配置"})
		return
	}

	if err := DeleteFlowConfig(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("流程配置已删除", zap.Uint("id", uint(id)))
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================================
// 流程操作
// ============================================================

func submitApproval(c *gin.Context) {
	stage := c.Param("stage")

	var req struct {
		WorkflowID string `json:"workflow_id"`
		Action     struct {
			Operator string `json:"operator"`
			Approved bool   `json:"approved"`
			Comment  string `json:"comment"`
		} `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	action := ApprovalAction{
		Stage:     stage,
		Operator:  req.Action.Operator,
		Approved:  req.Action.Approved,
		Comment:   req.Action.Comment,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err := temporalClient.SignalWorkflow(c.Request.Context(), req.WorkflowID, "", stage+"-approval", action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 只有通过时才更新版本当前阶段
	if req.Action.Approved {
		// 从 workflowID 解析 versionID
		versionID := req.WorkflowID[8:] // 去掉 "upgrade-" 前缀
		version, _ := GetVersionByID(versionID)
		if version != nil {
			// 获取流程配置
			flowConfig, _ := GetFlowConfig(version.FlowConfigID)
			if flowConfig != nil {
				stages, _ := GetFlowStages(flowConfig)
				version.CurrentStage = getNextStageFromConfig(stages, stage)
				UpdateVersion(version)
			}
		}
	}

	logger.Info("阶段审批已提交",
		zap.String("stage", stage),
		zap.String("operator", req.Action.Operator),
		zap.Bool("approved", req.Action.Approved))

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func getNextStageFromConfig(stages []StageConfig, currentStage string) string {
	found := false
	for _, stage := range stages {
		if found && stage.Enabled {
			return stage.Key
		}
		if stage.Key == currentStage {
			found = true
		}
	}
	return StageCompleted
}

func submitTestResult(c *gin.Context) {
	stage := c.Param("stage")

	var req struct {
		WorkflowID string `json:"workflow_id"`
		AllPassed  bool   `json:"all_passed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := temporalClient.SignalWorkflow(c.Request.Context(), req.WorkflowID, "", stage+"-test-complete", req.AllPassed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新版本当前阶段
	versionID := req.WorkflowID[8:]
	version, _ := GetVersionByID(versionID)
	if version != nil {
		flowConfig, _ := GetFlowConfig(version.FlowConfigID)
		if flowConfig != nil {
			stages, _ := GetFlowStages(flowConfig)
			version.CurrentStage = getNextStageFromConfig(stages, stage)
			UpdateVersion(version)
		}
	}

	logger.Info("测试结果已提交", zap.String("stage", stage), zap.Bool("allPassed", req.AllPassed))
	c.JSON(http.StatusOK, gin.H{"success": true})
}
