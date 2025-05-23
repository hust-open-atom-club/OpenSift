import { PageContainer } from "@ant-design/pro-components";
import { Button, Select, Space } from "antd";
import FlowView from "./components/FlowView";
import NodeConfig from "./components/NodeConfig";
import { useState } from "react";
import { TaskNode } from "./components/FlowView/Canvas";

const d: TaskNode[] = [
  { name: "dist_updated_0", title: "发行版已更新", description: "指示上游发行版已经更新的事件", args: "", status: "success", type: "event", dependencies: [] },
  { name: "update_manual", title: "手动更新", description: "手动更新", args: "", status: "pending", type: "event", dependencies: [] },
  { name: "debian", title: "更新 Debian 软件数据", description: "Debian 发行版", args: "", status: "success", type: "dist", dependencies: ["dist_updated_0"] },
  { name: "ubuntu", title: "更新 Ubuntu 软件数据", description: "Ubuntu 发行版", args: "", status: "success", type: "dist", dependencies: ["dist_updated_0"] },
  { name: "github", title: "枚举 GitHub 链接", description: "GitHub 代码仓库", args: "", status: "success", type: "repo", dependencies: ["dist_updated_0"] },
  { name: "gitlab", title: "枚举 GitLab 链接", description: "GitLab 代码仓库", args: "", status: "success", type: "repo", dependencies: ["dist_updated_0"] },
  { name: "analyze_dep_deps", title: "分析依赖关系", description: "", args: "", status: "success", type: "", dependencies: ["debian", "ubuntu", "github", "gitlab"] },
  { name: "analyze_dist_deps", title: "分析发行版依赖关系", description: "", args: "", status: "failed", type: "", dependencies: ["union"] },
  { name: "calc_score", title: "计算分数", description: "", args: "", status: "pending", type: "", dependencies: ["analyze_dist_deps"] },
  { name: 'union', title: "合并结果", description: "合并结果", args: "", status: "success", type: "", dependencies: ["analyze_dep_deps", "debian", "ubuntu", "github", "gitlab", "update_manual"] }
]


export default function () {
  const [selected, setSelected] = useState<TaskNode | undefined>(undefined);


  return <PageContainer extra={
    <Space>
      <Select options={[{
        label: "第 1 轮",
        value: 1
      }]} style={{ width: 150 }} placeholder="选择轮次" />
      <Button>到最新轮次</Button>
      <Button type="primary">刷新</Button>
      <Button danger>停止运行</Button>

    </Space>

  }>
    <div style={{
      height: "calc(100vh - 160px)"
    }} className="w-full bg-white rounded-md shadow-sm p-4 flex">
      {/* flexbox grow will cause bugs of resizeObserver, so use width instead */}
      <div className="h-full" style={{
        width: "calc(100% - 320px)",
      }}>
        <FlowView data={d} onSelect={setSelected} />
      </div>
      <div className="w-[320px] h-full">
        <NodeConfig node={selected} />
      </div>
    </div>

  </PageContainer >
}