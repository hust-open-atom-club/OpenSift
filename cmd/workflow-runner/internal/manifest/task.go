package manifest

import (
	"time"

	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/internal/workflow"
)

var (
	taskCalcScore           workflow.WorkflowNode
	taskUpdateDistruibution workflow.WorkflowNode
	taskSyncGitMetrics      workflow.WorkflowNode
	taskEnumeratePlatforms  workflow.WorkflowNode

	srcDistributionNeedUpdate  workflow.WorkflowNode
	srcGitlinkNeedUpdate       workflow.WorkflowNode // triggered manually
	srcGitPlatformNeedUpdate   workflow.WorkflowNode
	srcAllGitMetricsNeedUpdate workflow.WorkflowNode
)

var tasks []*workflow.WorkflowNode = []*workflow.WorkflowNode{
	&taskCalcScore,
	&taskUpdateDistruibution,
	&taskSyncGitMetrics,
	&taskEnumeratePlatforms,
	&srcDistributionNeedUpdate,
	&srcGitlinkNeedUpdate,
	&srcGitPlatformNeedUpdate,
	&srcAllGitMetricsNeedUpdate,
}

func GetAllTasks() []*workflow.WorkflowNode {
	return tasks
}

func GetTargetTask() *workflow.WorkflowNode {
	// This function is used to get the target task for the current workflow.
	// In this case, we return the taskCalcScore as the target task.
	return &taskCalcScore
}

func setNodeDefaults(node *workflow.WorkflowNode) {
	node.Title = "No title provided"
	node.Description = "No description provided"
	node.RunAfter = WorkflowRunAfter
	node.RunBefore = WorkflowRunBefore
	node.Type = "regular"
}

func initTasks() {
	/** calculate score **/
	setNodeDefaults(&taskCalcScore)
	taskCalcScore.Name = "calc-score"
	taskCalcScore.Title = "计算分数"
	taskCalcScore.Description = "将所有指标汇总，计算每个开源项目的最终分数"
	taskCalcScore.Run = WorkflowRunExecWrapper([]string{"bash", "-c", "sleep 10; echo 'taskCalcScore'"})
	taskCalcScore.Dependencies = []*workflow.WorkflowNode{
		&taskUpdateDistruibution,
	}

	/** update distribution **/
	setNodeDefaults(&taskUpdateDistruibution)
	taskUpdateDistruibution.Name = "update-distribution"
	taskUpdateDistruibution.Title = "更新发行版本"
	taskUpdateDistruibution.Description = "更新发行版本中的软件信息"
	taskUpdateDistruibution.Run = WorkflowRunExecWrapper([]string{"bash", "-c", "sleep 1; echo 'taskUpdateDistruibution'"})
	taskUpdateDistruibution.Dependencies = []*workflow.WorkflowNode{
		&srcDistributionNeedUpdate,
		&taskSyncGitMetrics,
	}

	/** sync git metrics **/
	setNodeDefaults(&taskSyncGitMetrics)
	taskSyncGitMetrics.Name = "sync-git-metrics"
	taskSyncGitMetrics.Title = "同步 Git 指标"
	taskSyncGitMetrics.Description = "将所有来源的 GitLink 汇总到 all_gitlinks 表中"
	taskSyncGitMetrics.Run = WorkflowRunExecWrapper([]string{"bash", "-c", "sleep 1; echo 'taskSyncGitMetrics'"})
	taskSyncGitMetrics.Dependencies = []*workflow.WorkflowNode{
		&srcGitlinkNeedUpdate,
		&taskEnumeratePlatforms,
	}

	/** enumerate platforms **/
	setNodeDefaults(&taskEnumeratePlatforms)
	taskEnumeratePlatforms.Name = "enumerate-platforms"
	taskEnumeratePlatforms.Title = "枚举平台"
	taskEnumeratePlatforms.Description = "枚举各大 git 平台，包括 GitHub， GitLab 等，获取最新的 GitLink 信息"
	taskEnumeratePlatforms.Run = WorkflowRunExecWrapper([]string{"bash", "-c", "sleep 1; echo 'taskEnumeratePlatforms'"})
	taskEnumeratePlatforms.Dependencies = []*workflow.WorkflowNode{
		&srcGitPlatformNeedUpdate,
	}
}

func initSources() {
	setNodeDefaults(&srcDistributionNeedUpdate)
	srcDistributionNeedUpdate.Name = "src-distribution-need-update"
	srcDistributionNeedUpdate.Title = "发行版本已更新"
	srcDistributionNeedUpdate.Type = "source"
	srcDistributionNeedUpdate.Description = "事件：发行版本已更新"
	srcDistributionNeedUpdate.NeedUpdate = NeedUpdateWrapper(&srcDistributionNeedUpdate, time.Minute*1)

	setNodeDefaults(&srcGitlinkNeedUpdate)
	srcGitlinkNeedUpdate.Name = "src-gitlink-need-update"
	srcGitlinkNeedUpdate.Title = "GitLink 已更新"
	srcGitlinkNeedUpdate.Type = "source"
	srcGitlinkNeedUpdate.Description = "事件：GitLink 已手动更新，这通常指的是发行版本的 GitLink 的更新"
	srcGitlinkNeedUpdate.NeedUpdate = NeedUpdateWrapper(&srcGitlinkNeedUpdate, time.Hour*24)

	setNodeDefaults(&srcGitPlatformNeedUpdate)
	srcGitPlatformNeedUpdate.Name = "src-git-platform-need-update"
	srcGitPlatformNeedUpdate.Title = "Git 平台已更新"
	srcGitPlatformNeedUpdate.Type = "source"
	srcGitPlatformNeedUpdate.Description = "事件：Git 平台已更新"
	srcGitPlatformNeedUpdate.NeedUpdate = NeedUpdateWrapper(&srcGitPlatformNeedUpdate, time.Hour*24)

	setNodeDefaults(&srcAllGitMetricsNeedUpdate)
	srcAllGitMetricsNeedUpdate.Name = "src-all-git-metrics-need-update"
	srcAllGitMetricsNeedUpdate.Title = "所有 Git 指标已更新"
	srcAllGitMetricsNeedUpdate.Type = "source"
	srcAllGitMetricsNeedUpdate.Description = "事件：所有 Git 指标已更新"
	srcAllGitMetricsNeedUpdate.NeedUpdate = NeedUpdateWrapper(&srcAllGitMetricsNeedUpdate, time.Hour*24)

}

func InitManifests() {
	initCmds()
	initTasks()
	initSources()
}
