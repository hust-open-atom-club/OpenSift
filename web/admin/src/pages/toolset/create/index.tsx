import { getAdminToolsetList, postAdminToolsetInstances } from "@/services/csapi/toolset";
import { AppstoreOutlined, PlayCircleFilled, SyncOutlined, ToolTwoTone } from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import { history } from "@umijs/max";
import { useRequest } from "ahooks";
import { App, Button, Card, Col, Drawer, Row, Spin, Typography } from "antd";
import { useState } from "react";

export default function () {

  const { message } = App.useApp();
  const { data, loading, run } = useRequest(getAdminToolsetList)
  const [args, setArgs] = useState<Record<string, any>>({});
  const { loading: launchLoading, run: launchRun } = useRequest(async (tool: API.ToolDTO) => {
    let args: Record<string, any> = {}
    tool.args?.forEach((p) => {
      if (p.name) args[p.name] = args[p.name] || p.default;
    })
    message.loading({
      content: "正在启动工具...",
      key: "launch"
    });
    try {
      const ret = await postAdminToolsetInstances({
        toolId: tool.id,
        args
      });
      message.success({
        content: "工具启动成功，正在附加到会话中...",
        key: "launch"
      });
      history.push(`/toolset/attach?id=${ret.id}`);

    } catch (e) {
      message.error({
        content: "工具启动失败，请检查启动参数是否正确。",
        key: "launch"
      });
    }

  }, {
    manual: true
  });

  const [selected, setSelected] = useState<API.ToolDTO>();



  return <PageContainer extra={<Button icon={<SyncOutlined />} onClick={run}>刷新 </Button>}>
    <Drawer open={!!selected} title="启动工具" destroyOnHidden onClose={() => {
      setSelected(undefined);
    }} footer={
      <div className="flex justify-end">
        <Button onClick={() => {
          setSelected(undefined);
        }} className="mr-2">取消</Button>
        <Button type="primary" icon={<PlayCircleFilled />} loading={loading} onClick={() => { launchRun(selected!) }}>启动</Button>
      </div>
    }>
      <div className="p-4">
        <h3 className="text-xl mb-4"><ToolTwoTone /> {selected?.name}</h3>
        <Typography.Paragraph>{selected?.description}</Typography.Paragraph>
        {(!selected?.args || selected?.args?.length === 0) && <div className="text-gray-500">该工具不需要任何参数。</div>}
        {
          selected?.args?.map((p) => (
            <div key={p.name} className="mb-4">
              <Typography.Text strong>{p.name}</Typography.Text>
              <div className="text-gray-500">{p.description}</div>
              <div className="mt-2">
                {p.type === "string" && <input type="text" className="border rounded-md p-2 w-full" />}
                {p.type === "number" && <input type="number" className="border rounded-md p-2 w-full" />}
                {p.type === "boolean" && <input type="checkbox" />}
              </div>
            </div>
          ))
        }
      </div>
    </Drawer>


    <h3 className="text-xl mb-4"><AppstoreOutlined /> 调试工具</h3>
    <Spin spinning={loading}>
      <Row gutter={[16, 16]}>
        {data?.map?.((tool) => (
          <Col span={6} key={tool.id}>
            <Card hoverable>
              <div className="flex flex-col h-40" >
                <div className="font-bold text-lg mb-4">
                  <ToolTwoTone /> <span>{tool.name}</span>
                </div>
                <div className="grow">
                  <Typography.Paragraph ellipsis={{
                    rows: 3,
                    tooltip: { placement: "bottom" }
                  }}>{tool.description}</Typography.Paragraph>
                </div>

                <div className="flex justify-end">
                  <Button type="primary" icon={<PlayCircleFilled />} onClick={() => {
                    setSelected(tool);
                  }}>启动</Button>
                </div>
              </div>
            </Card>
          </Col>
        ))}

      </Row>
    </Spin>

  </PageContainer>
}