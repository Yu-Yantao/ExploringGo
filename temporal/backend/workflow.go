package main

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const TaskQueue = "upgrade-workflow-queue"

// ============================================================
// 升级流程 Workflow（动态配置版本）
// 根据流程配置动态执行各个阶段
// ============================================================

func UpgradeWorkflow(ctx workflow.Context, req UpgradeWorkflowRequest) (UpgradeWorkflowResult, error) {
	result := UpgradeWorkflowResult{
		VersionID: req.Version.ID,
		Status:    "running",
	}

	// Activity 配置
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger.Info("升级流程开始", zap.String("version", req.Version.Name), zap.Uint("flowConfigId", req.FlowConfigID))

	// 获取流程配置
	var stages []StageConfig
	if err := workflow.ExecuteActivity(ctx, GetFlowConfigActivity, req.FlowConfigID).Get(ctx, &stages); err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("获取流程配置失败: %v", err)
		return result, err
	}

	// 动态执行每个阶段
	for _, stage := range stages {
		if !stage.Enabled {
			logger.Info("阶段已跳过", zap.String("stage", stage.Name))
			continue
		}

		result.CurrentStage = stage.Key
		logger.Info("开始执行阶段", zap.String("stage", stage.Name), zap.String("type", stage.Type))

		// 发送通知
		workflow.ExecuteActivity(ctx, NotifyActivity,
			fmt.Sprintf("版本 %s 进入【%s】阶段", req.Version.Name, stage.Name))

		// 根据阶段类型执行
		var err error
		switch stage.Type {
		case "approval":
			if stage.AutoPass {
				err = waitForStageApprovalWithAutoPass(ctx, stage.Key, time.Duration(stage.Timeout)*time.Hour)
			} else {
				err = waitForStageApproval(ctx, stage.Key, time.Duration(stage.Timeout)*time.Hour)
			}
		case "prepare":
			err = waitForStageApproval(ctx, stage.Key, time.Duration(stage.Timeout)*time.Hour)
		case "test":
			var testResult StageResult
			testResult, err = executeTestStage(ctx, req, stage.Key)
			if err == nil && !testResult.Passed {
				result.Status = "failed"
				result.Message = fmt.Sprintf("%s 未通过", stage.Name)
				return result, nil
			}
		}

		if err != nil {
			result.Status = "failed"
			result.Message = fmt.Sprintf("%s 失败: %v", stage.Name, err)
			return result, err
		}

		logger.Info("阶段完成", zap.String("stage", stage.Name))
	}

	// 流程完成
	result.CurrentStage = StageCompleted
	result.Status = "completed"
	result.Message = "升级流程完成"

	// 执行知识沉淀
	workflow.ExecuteActivity(ctx, ArchiveKnowledgeActivity, req.Version.ID).Get(ctx, nil)

	logger.Info("升级流程完成", zap.String("version", req.Version.Name))
	return result, nil
}

// ============================================================
// 阶段执行函数
// ============================================================

// executeTestStage 测试阶段
func executeTestStage(ctx workflow.Context, req UpgradeWorkflowRequest, stage string) (StageResult, error) {
	result := StageResult{Stage: stage, Passed: true}

	// 等待测试完成的 Signal
	testChan := workflow.GetSignalChannel(ctx, stage+"-test-complete")

	selector := workflow.NewSelector(ctx)
	timeoutTimer := workflow.NewTimer(ctx, 4*24*time.Hour)

	var testPassed bool
	var received bool
	selector.AddReceive(testChan, func(c workflow.ReceiveChannel, more bool) {
		if more {
			c.Receive(ctx, &testPassed)
			received = true
		}
	})
	selector.AddFuture(timeoutTimer, func(f workflow.Future) {})

	selector.Select(ctx)

	if !received {
		return result, fmt.Errorf("测试超时")
	}

	result.Passed = testPassed
	if !testPassed {
		result.Message = "测试存在不通过条目"
	}

	return result, nil
}

// ============================================================
// 等待审批辅助函数
// ============================================================

// waitForStageApproval 等待阶段审批
// 驳回后会继续等待重新审批，直到通过或超时
func waitForStageApproval(ctx workflow.Context, stage string, timeout time.Duration) error {
	approvalChan := workflow.GetSignalChannel(ctx, stage+"-approval")

	// 创建超时计时器
	timeoutCtx, cancelTimeout := workflow.WithCancel(ctx)
	timeoutFuture := workflow.NewTimer(timeoutCtx, timeout)

	for {
		selector := workflow.NewSelector(ctx)
		var action ApprovalAction
		var received bool
		var timedOut bool

		// 监听审批 Signal
		selector.AddReceive(approvalChan, func(c workflow.ReceiveChannel, more bool) {
			if more {
				c.Receive(ctx, &action)
				received = true
			}
		})

		// 监听超时
		selector.AddFuture(timeoutFuture, func(f workflow.Future) {
			timedOut = true
		})

		selector.Select(ctx)

		if timedOut {
			return fmt.Errorf("阶段 %s 审批超时", stage)
		}

		if received {
			if action.Approved {
				// 审批通过，取消超时计时器，继续流程
				cancelTimeout()
				logger.Info("审批通过", zap.String("stage", stage), zap.String("operator", action.Operator))
				return nil
			} else {
				// 驳回，记录日志，继续等待重新审批
				logger.Info("审批驳回，等待重新提交",
					zap.String("stage", stage),
					zap.String("operator", action.Operator),
					zap.String("comment", action.Comment))
				// 继续循环等待下一次审批
			}
		}
	}
}

// waitForStageApprovalWithAutoPass 等待阶段审批（超时自动通过）
func waitForStageApprovalWithAutoPass(ctx workflow.Context, stage string, timeout time.Duration) error {
	approvalChan := workflow.GetSignalChannel(ctx, stage+"-approval")

	selector := workflow.NewSelector(ctx)
	timeoutTimer := workflow.NewTimer(ctx, timeout)

	var action ApprovalAction
	var received bool
	selector.AddReceive(approvalChan, func(c workflow.ReceiveChannel, more bool) {
		if more {
			c.Receive(ctx, &action)
			received = true
		}
	})
	selector.AddFuture(timeoutTimer, func(f workflow.Future) {})

	selector.Select(ctx)

	if timeoutTimer.IsReady() && !received {
		logger.Info("超时自动通过", zap.String("stage", stage))
		return nil
	}

	if received && !action.Approved {
		return fmt.Errorf("阶段 %s 审批未通过: %s", stage, action.Comment)
	}

	return nil
}

// ============================================================
// Activities
// ============================================================

// GetFlowConfigActivity 获取流程配置 Activity
func GetFlowConfigActivity(ctx context.Context, flowConfigID uint) ([]StageConfig, error) {
	config, err := GetFlowConfig(flowConfigID)
	if err != nil {
		// 返回默认配置
		return []StageConfig{
			{Key: StageBTEConfirm, Name: "BTE条目确认", Type: "approval", Enabled: true, Timeout: 72, Order: 1},
			{Key: StageBTEFinalize, Name: "BTE定版", Type: "approval", Enabled: true, Timeout: 48, Order: 2},
			{Key: StageBTEPrepare, Name: "BTE版本准备", Type: "prepare", Enabled: true, Timeout: 24, Order: 3},
			{Key: StageBTETest, Name: "BTE测试", Type: "test", Enabled: true, Timeout: 96, Order: 4},
			{Key: StageProdFinalize, Name: "生产定版", Type: "approval", Enabled: true, Timeout: 48, Order: 5},
			{Key: StageProdPrepare, Name: "生产版本准备", Type: "prepare", Enabled: true, Timeout: 24, Order: 6},
			{Key: StageProdTest, Name: "生产测试", Type: "test", Enabled: true, Timeout: 96, Order: 7},
			{Key: StageCloseConfirm, Name: "关闭确认", Type: "approval", Enabled: true, Timeout: 72, AutoPass: true, Order: 8},
			{Key: StageEndConfirm, Name: "结束确认", Type: "approval", Enabled: true, Timeout: 48, AutoPass: true, Order: 9},
		}, nil
	}
	return GetFlowStages(config)
}

// NotifyActivity 通知 Activity
func NotifyActivity(ctx context.Context, message string) error {
	logger.Info("发送通知", zap.String("message", message))
	// 实际项目中：发送钉钉、邮件、短信等
	return nil
}

// ArchiveKnowledgeActivity 知识沉淀 Activity
func ArchiveKnowledgeActivity(ctx context.Context, versionID string) error {
	logger.Info("知识沉淀完成", zap.String("versionId", versionID))
	// 实际项目中：将需求、设计、测试用例等归档
	return nil
}

// ============================================================
// Worker 启动
// ============================================================

func StartWorker(c client.Client) {
	w := worker.New(c, TaskQueue, worker.Options{})

	// 注册 Workflow
	w.RegisterWorkflow(UpgradeWorkflow)

	// 注册 Activities
	w.RegisterActivity(GetFlowConfigActivity)
	w.RegisterActivity(NotifyActivity)
	w.RegisterActivity(ArchiveKnowledgeActivity)

	logger.Info("Worker 启动中...")
	err := w.Run(worker.InterruptCh())
	if err != nil {
		logger.Fatal("Worker 启动失败", zap.Error(err))
	}
}
