import { Button, Col, message, Popconfirm, Popover, Row, Statistic, Tooltip } from "antd";
import { SyncOutlined, InfoCircleFilled, MoreOutlined, LoadingOutlined, WarningOutlined, MessageOutlined } from "@ant-design/icons"
import { useInterval, useRequest } from "ahooks";
import { getAdminGitfilesStatus, ModelGitFileStatusResp, postAdminGitfilesStart, postAdminGitfilesStop, RpcRunningTaskDto } from "@/service/client";
import { useState } from "react";
import dayjs from "dayjs";

function RunningTask({ t }: { t: RpcRunningTaskDto }) {
  const [take, setTake] = useState("");

  useInterval(() => {
    if (t.start) {
      const d = dayjs().diff(dayjs(t.start), "second")
      if (d > 3600) {
        setTake(`${Math.floor(d / 3600)} 小时 ${Math.floor(d / 60) % 60} 分`)
      } else if (d > 60) {
        setTake(`${Math.floor(d / 60)} 分 ${d % 60} 秒`)
      } else if (d > 10) {
        setTake(`${d} 秒`);
      }
    }
  }, 1000);

  return <div>
    <span className="bg-blue-600 text-white px-1 rounded-md mr-2"><LoadingOutlined /> {take} </span> {t.link} <Tooltip title={<div className="max-h-40 overflow-auto whitespace-pre-wrap">{t.progress}</div>}>
      <MessageOutlined />
    </Tooltip>
  </div>
}


export default function GitFileSummary() {

  const { data, error, loading, run, runAsync } = useRequest<ModelGitFileStatusResp, never[]>(async () => {
    const d = (await getAdminGitfilesStatus({
      throwOnError: true
    })).data;
    // sort tasks
    d.collector?.currentTasks?.sort((a, b) => {
      return dayjs(a.start).diff(dayjs(b.start))
    })
    d.collector?.pendingTasks?.sort();
    return d;
  })

  const { run: startCollector, loading: startLoading } = useRequest(async () => {
    try {
      await postAdminGitfilesStart({
        throwOnError: true
      });
      message.success("执行成功！");
    } catch {
      message.error("执行失败！");
    }
  }, {
    manual: true
  })

  const { run: stopCollector, loading: stopLoading } = useRequest(async () => {
    try {
      await postAdminGitfilesStop({
        throwOnError: true
      });
      message.success("执行成功！");
    } catch {
      message.error("执行失败！");
    }
  }, {
    manual: true
  })


  const REFRESH_INTERVAL = 10;

  const [remaining, setRemaining] = useState(REFRESH_INTERVAL);


  useInterval(async () => {
    if (remaining == 0) {
      setRemaining(-1);
      try {
        await runAsync();
      } catch { }
      setTimeout(() => {
        setRemaining(REFRESH_INTERVAL);
      }, 0);
    } else if (remaining > 0) {
      setRemaining(x => x - 1);
    }
  }, 1000);

  return <div>
    <div className="flex gap-4 items-center pb-4">
      <Button icon={<SyncOutlined />} loading={loading} onClick={run}>
        刷新数据
      </Button>
      <Button color="orange" variant="outlined" loading={startLoading} onClick={startCollector}>开始同步</Button>
      {
        <Popconfirm title="确定要停止吗？" description="停止不会影响正在运行的任务。" onConfirm={stopCollector}>
          <Button color="danger" variant="outlined" loading={stopLoading} >停止同步</Button>
        </Popconfirm>
      }
      <div className="grow"></div>
      <div>
        {!data?.collector ? <span className="mr-2"><WarningOutlined />已离线</span> : (
          !data.collector.isRunning && <span className="mr-2"><WarningOutlined />已暂停</span>
        )}

        {remaining === -1 ? <span>正在刷新数据...</span> :
          <span> <InfoCircleFilled /> 统计数据将在 {remaining} 秒后自动刷新</span>}
        {error && <span className="text-red-500">获取数据失败！</span>}
      </div>
    </div>
    <Row gutter={16} className="my-4">
      <Col span={8} md={4}>
        <Statistic title="已完成仓库" value={data?.gitFile?.total} />
      </Col>
      <Col span={8} md={4} >
        <Statistic title="同步成功" value={data?.gitFile?.success} />
      </Col>

      <Col span={8} md={4} >
        <Statistic title="同步失败" value={data?.gitFile?.fail} />
      </Col>

      <Col span={8} md={4} >
        <Statistic title="从未同步成功" value={data?.gitFile?.neverSuccess} />
      </Col>

      <Col span={8} md={4} >
        <Statistic title="正在同步" value={data?.collector?.currentTasks?.length}
          suffix={
            <Popover placement="bottom" content={<div className="max-h-96 overflow-auto">
              {!data?.collector?.currentTasks?.length && <div>暂时没有正在同步的任务</div>}
              {data?.collector?.currentTasks?.map(x => (<RunningTask t={x} key={x.link} />))}
            </div>}>
              <MoreOutlined />
            </Popover>
          }
        />
      </Col>

      <Col span={8} md={4} >
        <Statistic title="队列中任务" value={data?.collector?.pendingTasks?.length}
          suffix={
            <Popover placement="bottom" content={<div className="max-h-96 overflow-auto">
              {!data?.collector?.pendingTasks?.length && <div>暂时没有队列中的任务</div>}
              {data?.collector?.pendingTasks?.map(x => (<div key={x}>{x}</div>))}
            </div>}>
              <MoreOutlined />
            </Popover>
          }

        />
      </Col>
    </Row>
  </div>
}