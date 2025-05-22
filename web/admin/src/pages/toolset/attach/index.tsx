import { getAdminToolsetInstances, getAdminToolsetInstancesId, getAdminToolsetInstancesIdLog, postAdminToolsetInstances, postAdminToolsetInstancesIdKill } from "@/services/csapi/toolset";
import { EllipsisOutlined, FileAddOutlined, InfoCircleOutlined, SyncOutlined } from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import { useRequest } from "ahooks";
import { App, Badge, Button, Dropdown, DropDownProps, Pagination, Popover, Segmented, Space, Spin, theme, } from "antd";
import dayjs from "dayjs";
import { useEffect, useState } from "react";
import Term from "./components/Term";
import { useSearchParams } from "@umijs/max";
import { history } from "@umijs/max";
import { getToken } from "@/bearer";

const fmtTime = (time: string | undefined) => {
  if (!time) return "";
  // if today, return time
  // else return date
  const today = dayjs().format("YYYY-MM-DD");
  const d = dayjs(time).format("YYYY-MM-DD");
  if (today === d) {
    return dayjs(time).format("HH:mm:ss");
  }
  return dayjs(time).format("YYYY-MM-DD");
}

const fmtTimeLong = (time: string | undefined) => {
  return time ? dayjs(time).format("YYYY-MM-DD HH:mm:ss") : "";
}


export default function () {
  const [filterRunning, setFilterRunning] = useState("running");

  const pageSize = 10;
  const [page, setPage] = useState(1);

  const { data, loading, run } = useRequest(() => {
    return getAdminToolsetInstances({
      all: filterRunning === "all",
      skip: (page - 1) * pageSize,
      take: pageSize,
    })
  }, {
    refreshDeps: [filterRunning, page],
  })
  const { useToken } = theme;
  const { token } = useToken();


  const [search, setSearch] = useSearchParams();
  const searchID = search.get("id");
  const [selectedID, setSelectedID] = useState<string | undefined>(searchID || undefined);

  const {
    data: selectedItem,
    loading: selectedItemLoading,
    run: selectedItemRun
  } = useRequest(async () => {
    if (!selectedID) return undefined;
    return await getAdminToolsetInstancesId({
      id: selectedID
    });
  }, {
    refreshDeps: [selectedID],
  });

  const { message, modal } = App.useApp();

  useEffect(() => {
    setSearch({
      id: selectedID || ""
    }, {
      replace: true
    })
  }, [selectedID]);

  const killMenu: DropDownProps["menu"] = {
    items: selectedItem?.tool?.allowedSignals?.map((item, k) => ({
      key: item.value || ("nov-" + k),
      type: "item",
      label: <div className="flex gap-2">
        <div className="text-3xl w-12 bg-black text-white rounded-md text-center">{item.value}</div>
        <div>
          <div>{item.name}</div>
          <div className="text-xs text-gray-600">{item.description}</div>
        </div>
      </div>

    })) || [],
    onClick: async (e) => {
      const k = e.key;
      const sig = selectedItem?.tool?.allowedSignals?.find((item) => item.value?.toString() === k);
      if (!sig || sig.value === undefined) {
        message.error("信号不存在");
        return;
      }
      const sigValue = sig.value;
      const result = await modal.confirm({
        title: "发送信号",
        content: <div className="flex flex-col gap-2">
          <p>你确定要发送信号 <span className="text-red-500">{sig.name} ({sig.value})</span> 吗？</p>
          <p>信号描述：{sig.description}</p>
        </div>,
        okText: "确定",
        cancelText: "取消",
      })
      if (!result) return;
      try {
        await postAdminToolsetInstancesIdKill({
          id: selectedID || "",
        }, {
          signal: sigValue
        });
        message.success("信号发送成功");
        run();
      } catch {
        message.error("信号发送失败");
      }
    }
  };

  const downloadMenu: DropDownProps["menu"] = {
    items: [
      {
        key: "shell",
        label: "使用 less 查看（推荐）"
      },
      {
        key: "download",
        label: "直接下载"
      }
    ],
    onClick: async (e) => {

      if (!selectedID) {
        message.error("请先选择一个进程");
        return;
      }

      const url = `${window.location.protocol}//${window.location.host}/api/v1/admin/toolset/instances/${selectedID}/log?all=true&auth_token=${encodeURIComponent(`Bearer ${getToken()}`)}`;
      if (e.key === "shell") {
        const toCopy = `curl -sL '${url}' | less -r`;
        try {
          await navigator.clipboard.writeText(toCopy);
          message.success("命令已复制到剪贴板，请在终端中粘贴运行");
        } catch (err) {
          console.error("Failed to copy: ", err);
          modal.error({
            title: "复制失败",
            content: <div className="flex flex-col gap-2">
              <p>请手动复制以下命令：</p>
              <code className="break-all select-all">{toCopy}</code>
            </div>
          });
        }
      } else if (e.key === "download") {
        // open new tab to download
        const a = document.createElement("a");
        a.href = url;
        a.download = `${selectedID}.log`;
        a.target = "_blank";
        a.click();
        a.remove();
      }
    }
  }


  return <PageContainer extra={<Space>
    {selectedItemLoading && <Spin />}
    {(!selectedItemLoading && selectedItem) && <>
      <Popover placement="bottom" title={<InstanceDetail item={selectedItem} />}>
        <Button icon={<InfoCircleOutlined />} shape="circle" /> </Popover>
      {selectedItem.isRunning && (selectedItem.tool?.allowedSignals?.length || 0) > 0 && <Dropdown menu={killMenu} trigger={["click"]}>
        <Button danger>
          <Space>
            发送信号
            <EllipsisOutlined />
          </Space>
        </Button>
      </Dropdown>}
      <Dropdown menu={downloadMenu} trigger={["click"]}>
        <Button>
          <Space>
            下载日志
            <EllipsisOutlined />
          </Space>
        </Button>
      </Dropdown>
    </>}

    <Button icon={<SyncOutlined />} onClick={run}>刷新列表</Button>
  </Space>}>
    <div className="flex gap-2" style={{
      height: "calc(100vh - 210px)",
    }}>
      <div className="bg-white rounded-lg h-full p-2 shadow-md flex flex-col gap-2">
        <div>
          <Button icon={<FileAddOutlined />} className="w-full mb-2" onClick={() => {
            history.push("/toolset/create");
          }}>创建新进程</Button>
          <Segmented options={[{
            label: "运行中",
            value: "running"
          }, {
            label: "全部",
            value: "all"
          }]} defaultValue="running" block value={filterRunning} onChange={(e) => {
            setFilterRunning(e);
          }
          } />
        </div>
        <div className="overflow-auto pb-2 flex flex-col gap-2 max-w-50 grow">
          {loading ? <div className="flex justify-center items-center h-full"> <Spin /> </div> :
            data?.items?.map((tool) => {
              return <Popover placement="right" key={tool.id} title={<InstanceDetail item={tool} />}>
                <div className="bg-white rounded-lg border border-transparent hover:border-slate-300 py-2 px-4 cursor-pointer transition-all"
                  style={selectedID === tool.id ? {
                    background: token.colorPrimary,
                    color: token.colorTextLightSolid

                  } : {}} onClick={() => setSelectedID(tool.id)}>
                  <div>
                    <div className="text-md">
                      <Badge className="mr-1" status={tool.isRunning ? "success" : "error"} />
                      {tool.toolName}
                    </div>
                    <div className="text-sm">{fmtTime(tool.startTime)}</div>
                  </div>
                </div>
              </Popover>
            })

          }
        </div>
        <div>
          <Pagination simple current={page} pageSize={pageSize} total={data?.total || 1} onChange={(e) => {
            setPage(e);
          }} />
        </div>
      </div>
      <div className="grow rounded-md overflow-hidden p-2 bg-[#101420]">
        {!!selectedID && <Term id={selectedID} onRefresh={() => {
          run(); selectedItemRun();
        }} />}
      </div>
    </div>
  </PageContainer>
}

function InstanceDetail({ item: tool }: { item: API.ToolInstanceHistoryDTO }) {
  return <div className="max-w-50">
    <p className="text-md font-bold"> {tool.toolName} </p>
    <p className="text-xs text-slate-500 mb-2"> {tool.id} </p>
    <p className="text-xs text-slate-500"> 状态： {tool.isRunning ? <span className="text-green-600">运行中</span> : <span className="text-red-600">已停止</span>} </p>
    <p className="text-xs text-slate-500"> 开始时间: {fmtTimeLong(tool.startTime)} </p>
    {tool.endTime && <p className="text-xs text-slate-500"> 结束时间: {fmtTimeLong(tool.endTime)} </p>}
    {tool.launchUserName && <p className="text-xs text-slate-500"> 启动用户: {tool.launchUserName} </p>}
    {(!tool.isRunning) && <p className="text-xs text-slate-500"> 退出码: {tool.ret === null ? "异常退出" : tool.ret} </p>}
    {tool.err && <p className="text-xs text-slate-500"> 错误信息: {tool.err} </p>}
  </div>

}