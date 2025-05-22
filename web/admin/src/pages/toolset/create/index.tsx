import { getAdminToolsetList, postAdminToolsetInstances } from "@/services/csapi/toolset";
import { AppstoreOutlined, PlayCircleFilled, SyncOutlined, ToolTwoTone } from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import { history } from "@umijs/max";
import { useRequest } from "ahooks";
import { App, Button, Card, Col, Drawer, Input, Row, Space, Spin, Switch, Typography } from "antd";
import { useState } from "react";

export default function () {

  const { message } = App.useApp();
  const { data, loading, run } = useRequest(getAdminToolsetList)
  const [filter, setFilter] = useState("");
  const [args, setArgs] = useState<Record<string, any>>({});
  const { loading: launchLoading, run: launchRun } = useRequest(async (tool: API.ToolDTO) => {
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

  const groups = data?.filter(x => x.name?.includes(filter))?.reduce((acc, tool) => {
    const grpName = tool.group || "其他";
    const group = acc[grpName] || [];
    group.push(tool);
    acc[grpName] = group;
    return acc;
  }, {} as Record<string, API.ToolDTO[]>) || {};

  return <PageContainer loading={loading} extra={
    <Space>
      <Input placeholder="搜索工具" allowClear onChange={(e) => {
        setFilter(e.target.value);
      }} style={{ width: 300 }} />
      <Button icon={<SyncOutlined />} onClick={run}>刷新 </Button>
    </Space>}>
    <Drawer open={!!selected} title="启动工具" destroyOnHidden onClose={() => {
      setSelected(undefined);
    }} footer={
      <div className="flex justify-end">
        <Button onClick={() => {
          setSelected(undefined);
        }} className="mr-2">取消</Button>
        <Button type="primary" icon={<PlayCircleFilled />} loading={launchLoading} onClick={() => { launchRun(selected!) }}>启动</Button>
      </div>
    }>
      <div className="p-4">
        <h3 className="text-xl mb-4"><ToolTwoTone /> {selected?.name}</h3>
        <Typography.Paragraph>{selected?.description}</Typography.Paragraph>
        <ArgInput args={selected?.args} value={args} onChange={(e) => {
          console.log(e);
          setArgs(e);
        }} />
      </div>
    </Drawer>

    {
      groups && Object.keys(groups).map((group) => {
        return <div key={group}>
          <h3 className="text-xl my-4"><AppstoreOutlined /> {group}</h3>
          <Row gutter={[16, 16]}>
            {groups[group].map((tool) => (
              <Col span={6} key={tool.id}>
                <Card hoverable onClick={() => {
                  setSelected(tool);
                }}>
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
        </div>
      })
    }


  </PageContainer>
}


function ArgInput(props: {
  args?: API.ToolArgDTO[];
  value?: Record<string, any>;
  onChange?: (args: Record<string, any>) => void;
}) {
  const { args, value, onChange } = props;

  if (!args || args.length === 0) {
    return <div className="text-gray-500">该工具不需要任何参数。</div>;
  }

  const handleChange = (name: string, val: any) => {
    onChange?.({ ...value, [name]: val });
  };

  return (
    <div>
      {args.map((arg) => (
        <div key={arg.name} className="mb-4">
          <Typography.Text strong>{arg.name}</Typography.Text>
          <div className="text-gray-500">{arg.description}</div>
          <div className="mt-2">
            {arg.type === "string" && (
              <Input
                placeholder={`请输入${arg.name}`}
                value={value?.[arg.name ?? ''] ?? arg.default}
                onChange={(e) => handleChange(arg.name ?? '', e.target.value)}
                className="w-full"
              />
            )}
            {arg.type === "int" && (
              <Input
                type="number"
                placeholder={`请输入${arg.name}`}
                value={value?.[arg.name ?? ''] ?? arg.default}
                onChange={(e) => handleChange(arg.name ?? '', parseInt(e.target.value, 10) || 0)}
                className="w-full"
              />
            )}
            {arg.type === "float" && (
              <Input
                type="number"
                step="0.01"
                placeholder={`请输入${arg.name}`}
                value={value?.[arg.name ?? ''] ?? arg.default}
                onChange={(e) => handleChange(arg.name ?? '', parseFloat(e.target.value) || 0)}
                className="w-full"
              />
            )}
            {arg.type === "bool" && (
              <div>
                <Switch
                  checked={value?.[arg.name ?? ''] ?? arg.default}
                  onChange={(checked) => handleChange(arg.name ?? '', checked)}
                />
              </div>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}