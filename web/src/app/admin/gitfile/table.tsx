import { getAdminGitfiles, ModelGitFileStatusResp, ModelPageDtoModelGitFileDto, ModelGitFileDto, postAdminGitfilesManual } from "@/service/client"
import { Button, Flex, Input, message, Pagination, Result, Select, Spin, Table } from "antd";
import { CloseCircleFilled, CheckCircleFilled, SyncOutlined } from "@ant-design/icons"
import dayjs from 'dayjs'
import Paragraph from "antd/es/typography/Paragraph"
import { useRequest } from "ahooks"
import { useState } from "react";

function GitFileOperations({ g }: {
  g: ModelGitFileDto
}) {
  const { loading, run } = useRequest(async () => {
    try {
      await postAdminGitfilesManual({
        body: {
          gitLink: g.gitLink
        },
        throwOnError: true
      });
      message.success("已经加入同步列表！")
    } catch {
      message.error("操作失败！")
    }
  }, {
    manual: true
  })
  return <div>
    <Button size="small" onClick={run} loading={loading}>立即同步</Button>
  </div>
}

export default function GitFileTable() {
  const [search, setSearch] = useState("")
  const [filter, setFilter] = useState(0)
  const DEFAULT_PAGESIZE = 20;

  const { data, error, loading, run } = useRequest<ModelPageDtoModelGitFileDto, {
    pageSize?: number;
    current?: number;
  }[]>(async (arg) => {
    const c = arg?.current || 1;
    const skip = (c - 1) * (arg?.pageSize || DEFAULT_PAGESIZE);
    return (await getAdminGitfiles({
      query: {
        link: search,
        filter: filter,
        skip: skip,
        take: arg?.pageSize || DEFAULT_PAGESIZE
      },
      throwOnError: true
    })).data
  })



  if (error) {
    return <Result
      status="500" title="加载失败" extra={
        <Button onClick={() => { run() }}>重新加载</Button>
      } />
  }

  return <div>
    <div className="flex py-4 gap-4">
      <Button shape="circle" icon={<SyncOutlined />} onClick={() => { run() }} loading={loading} />
      <Select className="w-40" value={filter} onChange={x => setFilter(x)} options={[
        { value: 0, label: "（不筛选）" },
        { value: 1, label: "成功" },
        { value: 2, label: "失败" },
        { value: 3, label: "从未成功" },
      ]}> </Select>
      <Input className="grow" value={search} onChange={(x) => setSearch(x.target.value)} placeholder="搜索仓库链接..." onPressEnter={() => { run() }} allowClear />
      <Button type="primary" onClick={() => { run() }} loading={loading}>搜索</Button>
      <Button onClick={() => {
        setSearch("")
        setTimeout(run, 0)
      }}>清空</Button>
    </div>
    <Table<ModelGitFileDto>
      sticky={{
        offsetHeader: 60,
        offsetScroll: 64
      }}
      scroll={{ x: 1800 }}
      rowKey={x => x.gitLink!!}
      columns={[
        { dataIndex: "gitLink", title: "仓库链接", width: 400 },
        {
          dataIndex: "filePath", title: "文件路径", width: 200, render: (v) => (<Paragraph className="!mb-0 max-h-6 overflow-hidden" copyable ellipsis={{ rows: 1, tooltip: { placement: "bottom" } }}>{v}</Paragraph>)
        },
        {
          dataIndex: "success", title: "上次结果", render: (_, r) => {
            if (r.success) {
              return <span><CheckCircleFilled className="!text-green-500 mr-2" />成功</span>
            } else {
              return <span><CloseCircleFilled className="!text-red-500 mr-2" />失败</span>
            }
          },
          width: 100
        },
        { dataIndex: "lastSuccess", title: "上次成功时间", render: (v) => v && dayjs(v).format("YYYY-MM-DD HH:mm:ssZ"), width: 280 },
        // { dataIndex: "takeStorage", title: "占用存储" },
        { dataIndex: "takeTimeMs", title: "上次执行时长（毫秒）", width: 120 },
        { dataIndex: "updateTime", title: "更新时间", width: 280 },
        { dataIndex: "message", title: "失败消息", render: (v) => (<Paragraph className="!mb-0 max-h-6 overflow-hidden" copyable={!!v} ellipsis={{ rows: 1, tooltip: { placement: "bottom" } }}>{v}</Paragraph>) },
        { title: "操作", render: (_, v) => <GitFileOperations g={v} /> }
      ]}
      dataSource={data?.items}
      loading={loading}
      pagination={false}
    // pagination={{
    //   current: data ? (data.start!! / data.count!!) + 1 : 1,
    //   total: data?.total,
    //   pageSize: data?.count,
    // }}
    // onChange={(x) => {
    //   run({
    //     current: x.current || 0,
    //     pageSize: x.pageSize || 100,
    //   })
    // }}
    ></Table>
    <div className="sticky bottom-0 py-4 bg-white border-t border-slate-200 flex justify-end">
      <Spin spinning={loading}>
        <Pagination
          showQuickJumper
          current={data ? (data.start!! / data.count!!) + 1 : 1}
          total={data?.total}
          pageSize={data?.count}
          defaultPageSize={DEFAULT_PAGESIZE}
          onChange={(current, pageSize) => {
            run({ current, pageSize })
          }}
        ></Pagination>
      </Spin>
    </div>
  </div>

}