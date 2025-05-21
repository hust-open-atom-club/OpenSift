import { getAdminToolsetInstances } from "@/services/csapi/toolset";
import { FileAddOutlined, SyncOutlined } from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import { useRequest } from "ahooks";
import { App, Button, Card, Spin, theme } from "antd";
import dayjs from "dayjs";
import { useEffect, useState } from "react";
import Term from "./components/Term";
import { useSearchParams } from "@umijs/max";
import { history } from "@umijs/max";

export default function () {

  const { data, loading, run } = useRequest(getAdminToolsetInstances)
  const { useToken } = theme;
  const { token } = useToken();



  const [search, setSearch] = useSearchParams();
  const searchID = search.get("id");
  const [selectedID, setSelectedID] = useState<string | undefined>(searchID || undefined);

  useEffect(() => {
    setSearch({
      id: selectedID || ""
    }, {
      replace: true
    })
  }, [selectedID])


  return <PageContainer extra={<Button icon={<SyncOutlined />} onClick={run}>刷新 </Button>}>
    <div className="flex gap-2" style={{
      height: "calc(100vh - 200px)",
    }}>
      <Card className="h-full" styles={{
        body: {
          padding: 8
        }
      }}>
        <Spin spinning={loading}>
          <div className="overflow-auto pb-2 flex flex-col gap-2 w-40">
            <Button icon={<FileAddOutlined />} className="w-full" onClick={() => {
              history.push("/toolset/create");
            }}>创建新进程</Button>
            {
              data?.map((tool) => {
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


                return <div key={tool.id} className="bg-white rounded-lg border border-transparent hover:border-slate-300 py-2 px-4 cursor-pointer"
                  style={selectedID === tool.id ? {
                    background: token.colorPrimary,
                    color: token.colorTextLightSolid

                  } : {}} onClick={() => setSelectedID(tool.id)}>
                  <div>
                    <div className="text-md">{tool.tool?.name}</div>
                    <div className="text-sm">{fmtTime(tool.startTime)}</div>
                  </div>
                </div>
              })

            }

          </div>
        </Spin>
      </Card>
      <div className="grow rounded-md overflow-hidden p-2 bg-[#101420]">
        {!!selectedID && <Term id={selectedID} />}
      </div>
    </div>
  </PageContainer>
}