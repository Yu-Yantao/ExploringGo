package main

// ============================================================
// 升级流程核心数据结构
// ============================================================

// UpgradeVersion 升级版本
type UpgradeVersion struct {
	ID           string   `json:"id"`            // 版本编号
	Name         string   `json:"name"`          // 版本名称
	VersionOwner string   `json:"version_owner"` // 版本负责人
	VendorOwner  string   `json:"vendor_owner"`  // 厂家负责人
	BTETester    string   `json:"bte_tester"`    // BTE测试负责人
	GrayTester   string   `json:"gray_tester"`   // 灰度测试负责人
	ProdTester   string   `json:"prod_tester"`   // 生产测试负责人
	IsUrgent     bool     `json:"is_urgent"`     // 是否紧急版本（跳过灰度）
	Status       string   `json:"status"`        // 状态
	CurrentStage string   `json:"current_stage"` // 当前阶段
	ItemIDs      []string `json:"item_ids"`      // 包含的条目ID
	CreatedAt    string   `json:"created_at"`    // 创建时间
	CompletedAt  string   `json:"completed_at"`  // 完成时间
}

// UpgradeItem 升级条目
type UpgradeItem struct {
	ID            string   `json:"id"`             // 条目ID
	Name          string   `json:"name"`           // 条目名称
	Type          string   `json:"type"`           // 类型：BUG/需求/优化
	RequirementID string   `json:"requirement_id"` // 关联需求ID
	Developer     string   `json:"developer"`      // 开发人员
	Tester        string   `json:"tester"`         // 测试人员
	ItemOwner     string   `json:"item_owner"`     // 条目负责人
	Status        string   `json:"status"`         // 状态：已登记/测试完成/审核完成/BTE已定版/灰度已定版/生产已定版/已关闭
	HasScript     bool     `json:"has_script"`     // 是否有脚本
	HasCache      bool     `json:"has_cache"`      // 是否涉及缓存
	HasRestart    bool     `json:"has_restart"`    // 是否需要重启
	BTEResult     string   `json:"bte_result"`     // BTE测试结果
	GrayResult    string   `json:"gray_result"`    // 灰度测试结果
	ProdResult    string   `json:"prod_result"`    // 生产测试结果
	CloseReason   string   `json:"close_reason"`   // 关闭原因
	TestCases     []string `json:"test_cases"`     // 测试用例
	CreatedAt     string   `json:"created_at"`     // 创建时间
}

// StageConfirmation 阶段确认信息
type StageConfirmation struct {
	Stage       string `json:"stage"`        // 阶段
	ConfirmedBy string `json:"confirmed_by"` // 确认人
	ConfirmedAt string `json:"confirmed_at"` // 确认时间
	Comment     string `json:"comment"`      // 备注
}

// TestSubmission 测试提交
type TestSubmission struct {
	ItemID      string   `json:"item_id"`      // 条目ID
	Stage       string   `json:"stage"`        // 阶段：BTE/灰度/生产
	Tester      string   `json:"tester"`       // 测试人员
	Passed      bool     `json:"passed"`       // 是否通过
	BugDesc     string   `json:"bug_desc"`     // BUG描述（如果不通过）
	Artifacts   []string `json:"artifacts"`    // 测试产物
	SubmittedAt string   `json:"submitted_at"` // 提交时间
}

// VersionPrepare 版本准备信息
type VersionPrepare struct {
	Stage       string `json:"stage"`        // 阶段
	PreparedBy  string `json:"prepared_by"`  // 准备人
	StartedAt   string `json:"started_at"`   // 开始时间
	CompletedAt string `json:"completed_at"` // 完成时间
	UpgradeLog  string `json:"upgrade_log"`  // 升级日志
}

// ApprovalAction 审批动作
type ApprovalAction struct {
	Stage      string `json:"stage"`       // 阶段
	ActionType string `json:"action_type"` // 动作类型：confirm/finalize/prepare/test/close
	Operator   string `json:"operator"`    // 操作人
	Approved   bool   `json:"approved"`    // 是否通过
	Comment    string `json:"comment"`     // 备注
	Timestamp  string `json:"timestamp"`   // 时间戳
}

// ============================================================
// Workflow 请求/响应结构
// ============================================================

// UpgradeWorkflowRequest 升级流程请求
type UpgradeWorkflowRequest struct {
	Version      UpgradeVersion `json:"version"`
	Items        []UpgradeItem  `json:"items"`
	FlowConfigID uint           `json:"flow_config_id"`
}

// UpgradeWorkflowResult 升级流程结果
type UpgradeWorkflowResult struct {
	VersionID    string `json:"version_id"`
	Status       string `json:"status"` // completed/failed/cancelled
	CurrentStage string `json:"current_stage"`
	Message      string `json:"message"`
}

// StageResult 阶段结果
type StageResult struct {
	Stage       string   `json:"stage"`
	Passed      bool     `json:"passed"`
	PassedItems []string `json:"passed_items"` // 通过的条目
	FailedItems []string `json:"failed_items"` // 失败的条目
	Message     string   `json:"message"`
}

// ============================================================
// HTTP API 请求/响应结构
// ============================================================

// CreateVersionRequest 创建版本请求
type CreateVersionRequest struct {
	Name         string   `json:"name"`
	VersionOwner string   `json:"version_owner"`
	VendorOwner  string   `json:"vendor_owner"`
	BTETester    string   `json:"bte_tester"`
	GrayTester   string   `json:"gray_tester"`
	ProdTester   string   `json:"prod_tester"`
	IsUrgent     bool     `json:"is_urgent"`
	ItemIDs      []string `json:"item_ids"`
}

// CreateVersionResponse 创建版本响应
type CreateVersionResponse struct {
	WorkflowID string `json:"workflow_id"`
	VersionID  string `json:"version_id"`
	RunID      string `json:"run_id"`
}

// SubmitActionRequest 提交动作请求
type SubmitActionRequest struct {
	WorkflowID string         `json:"workflow_id"`
	Action     ApprovalAction `json:"action"`
}

// SubmitTestRequest 提交测试结果请求
type SubmitTestRequest struct {
	WorkflowID string         `json:"workflow_id"`
	Test       TestSubmission `json:"test"`
}

// VersionStatusResponse 版本状态响应
type VersionStatusResponse struct {
	VersionID    string          `json:"version_id"`
	VersionName  string          `json:"version_name"`
	Status       string          `json:"status"`
	CurrentStage string          `json:"current_stage"`
	Items        []UpgradeItem   `json:"items"`
	Timeline     []StageTimeline `json:"timeline"`
}

// StageTimeline 阶段时间线
type StageTimeline struct {
	Stage       string `json:"stage"`
	Status      string `json:"status"` // pending/in_progress/completed/skipped
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
	Operator    string `json:"operator"`
}

// ItemListResponse 条目列表响应
type ItemListResponse struct {
	Items []UpgradeItem `json:"items"`
}

// CreateItemRequest 创建条目请求
type CreateItemRequest struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	RequirementID string   `json:"requirement_id"`
	Developer     string   `json:"developer"`
	Tester        string   `json:"tester"`
	ItemOwner     string   `json:"item_owner"`
	HasScript     bool     `json:"has_script"`
	HasCache      bool     `json:"has_cache"`
	HasRestart    bool     `json:"has_restart"`
	TestCases     []string `json:"test_cases"`
}

// ============================================================
// 常量定义
// ============================================================

// 流程阶段
const (
	StageBTEConfirm   = "bte_confirm"   // BTE条目确认
	StageBTEFinalize  = "bte_finalize"  // BTE定版
	StageBTEPrepare   = "bte_prepare"   // BTE版本准备
	StageBTETest      = "bte_test"      // BTE测试
	StageGrayConfirm  = "gray_confirm"  // 灰度条目确认
	StageGrayFinalize = "gray_finalize" // 灰度定版
	StageGrayPrepare  = "gray_prepare"  // 灰度版本准备
	StageGrayTest     = "gray_test"     // 灰度测试
	StageProdFinalize = "prod_finalize" // 生产定版
	StageProdPrepare  = "prod_prepare"  // 生产版本准备
	StageProdTest     = "prod_test"     // 生产测试
	StageCloseConfirm = "close_confirm" // 关闭确认
	StageEndConfirm   = "end_confirm"   // 结束确认
	StageCompleted    = "completed"     // 已完成
)

// 条目状态
const (
	ItemStatusRegistered    = "已登记"
	ItemStatusTestComplete  = "测试完成"
	ItemStatusAuditComplete = "审核完成"
	ItemStatusBTEFinalized  = "BTE已定版"
	ItemStatusGrayFinalized = "灰度已定版"
	ItemStatusProdFinalized = "生产已定版"
	ItemStatusClosed        = "已关闭"
	ItemStatusSuspended     = "挂起"
)

// 测试结果
const (
	TestResultPending = "待测试"
	TestResultPassed  = "通过"
	TestResultFailed  = "不通过"
	TestResultSkipped = "跳过"
)
