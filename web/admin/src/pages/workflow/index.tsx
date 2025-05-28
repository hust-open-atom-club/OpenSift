import { PageContainer } from "@ant-design/pro-components";
import { Alert, Button, Popconfirm, Select, Space } from "antd";
import FlowView from "./components/FlowView";
import NodeConfig from "./components/NodeConfig";
import { useState } from "react";
import { useRequest } from "ahooks";
import { getAdminWorkflowsMaxRounds, getAdminWorkflowsRoundsId, postAdminWorkflowsStatus } from "@/services/csapi/workflow";
import { SyncOutlined } from "@ant-design/icons";

type TaskNode = API.TaskDTO;

export default function () {
  const [selected, setSelected] = useState<TaskNode | undefined>(undefined);

  const [currentRound, setCurrentRound] = useState<number | string>("");

  const { data: maxRound, run: runMaxRound, error } = useRequest(async () => {
    const res = await getAdminWorkflowsMaxRounds();
    setCurrentRound(res.currentRound || -1);
    return res.currentRound || -1;
  });

  const { data: d, run } = useRequest(async () => {
    let c = currentRound;
    if (c === -1) return undefined;
    if (typeof c === 'string') {
      if (c === "next") {
        return //TODO:
      }
      else { return }
    }

    const res = await getAdminWorkflowsRoundsId({ id: c });
    return res;
  }, {
    refreshDeps: [currentRound],
  })



  return <PageContainer extra={
    <Space>
      {
        error && <Alert message="获取轮次失败，可能是执行器离线" className="!py-1" type="error" showIcon />
      }
      <Button icon={<SyncOutlined />} onClick={runMaxRound} />
      <Select virtual options={maxRound !== undefined ? new Array(maxRound).fill(0).map((_, k) => ({
        label: `第 ${maxRound - k} 轮`,
        value: maxRound - k,
      })) : []} value={currentRound} onChange={setCurrentRound} style={{ width: 150 }} placeholder="选择轮次" />
      <Button onClick={() => {
        if (maxRound === undefined) return;
        setCurrentRound(maxRound);
      }}>到最新轮次</Button>
      <Button type="primary" onClick={run}>刷新</Button>
      <Button onClick={async () => {
        await postAdminWorkflowsStatus({
          running: true,
        });
      }}>开始运行</Button>

      <Popconfirm title="确定要停止当前轮次的所有任务吗？" placement="bottomLeft" onConfirm={async () => {
        await postAdminWorkflowsStatus({
          running: false,
        });
      }}>
        <Button danger>停止运行</Button>
      </Popconfirm>

    </Space>

  }>
    <div style={{
      height: "calc(100vh - 160px)"
    }} className="w-full bg-white rounded-md shadow-sm p-4 flex">
      {/* flexbox grow will cause bugs of resizeObserver, so use width instead */}
      <div className="h-full" style={{
        width: "calc(100% - 320px)",
      }}>
        <FlowView data={d?.tasks} onSelect={setSelected} />
      </div>
      <div className="w-[320px] h-full">
        <NodeConfig node={selected} round={currentRound} />
      </div>
    </div>

  </PageContainer >
}