import { Button, Result } from "antd";
import { ReactElement, ReactNode, useEffect, useState } from "react";
import { formatTime } from "@/utils/format";
import { JsonEditor } from "json-edit-react";

type TaskNode = API.TaskDTO

type Props = {
  node?: TaskNode;
}

function Desc({ title, children }: {
  title: string,
  children: ReactNode
}) {
  return <div className="mt-4">
    <h4 className="text-md font-semibold mb-2">{title}</h4>
    <div>
      {children}
    </div>
  </div>
}


export default function ({
  node
}: Props) {
  const [args, setArgs] = useState<string>(node?.args || "");
  useEffect(() => {
    if (node) {
      setArgs(node.args || "");
    }
  }, [node]);


  return <div className="border shadow-lg rounded-lg ml-4 p-4 h-full overflow-auto">
    <h2 className="text-lg font-semibold">操作面板</h2>
    {!node && (
      <div className="mt-4">
        <Result
          status="warning"
          title="没有选中节点"
          subTitle="请在图中选择一个节点进行配置"
        />
      </div>
    )}
    {node && (
      <div className="mt-4">
        <h3 className="text-lg font-semibold">{node.title}</h3>
        <p className="text-sm text-gray-500">{node.description}</p>
        <Desc title="节点 ID">{node.name}</Desc>

        {node.startTime && <Desc title="开始时间">{formatTime(node.startTime)}</Desc>}
        {node.endTime && <Desc title="结束时间">{formatTime(node.endTime)}</Desc>}

        <Desc title="当前状态">
          <p>{node.status === 'pending' ? '待执行' : node.status === 'running' ? '执行中' : node.status === 'success' ? '成功' : node.status === 'failed' ? '失败' : ''}</p>
        </Desc>

        <Desc title="节点类型"> {node.type} </Desc>

        <Desc title="依赖节点"> {(node.dependencies?.length || 0) > 0 ? <ul className="list-disc pl-4">
          {node.dependencies?.map((dep) => <li key={dep} className="text-sm text-gray-500">{dep}</li>)}
        </ul> : '无'} </Desc>

        <Desc title="参数配置">
          <>
            <JsonEditor data={args} setData={(e) => { setArgs(e as string); }} translations={{
              ITEM_SINGLE: '{{count}} 项',
              ITEMS_MULTIPLE: '{{count}} 项',
              KEY_NEW: '输入新键',
              KEY_SELECT: '选择键',
              NO_KEY_OPTIONS: '没有键选项',
              ERROR_KEY_EXISTS: '键已存在',
              ERROR_INVALID_JSON: '无效的 JSON',
              ERROR_UPDATE: '更新失败',
              ERROR_DELETE: '删除失败',
              ERROR_ADD: '添加节点失败',
              DEFAULT_STRING: '新数据！',
              DEFAULT_NEW_KEY: '键',
              SHOW_LESS: '(显示更少)'
            }} />
            <Button block className="mt-2">保存配置</Button>
          </>
        </Desc>
        <Desc title="日志输出位置"> stdout </Desc>
        <Desc title="操作">
          <Button type="primary" block className="mb-2">查看输出</Button>
          <Button danger block>停止执行</Button>
        </Desc>

      </div>)
    }
  </div >
}