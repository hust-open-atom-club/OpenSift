import { PageContainer } from "@ant-design/pro-components";
import { Button, Card, Segmented, Select, Splitter } from "antd";
import GitLinkForm from "./components/GitLinkForm";
import GitLinkTable, { GitLinkTableAction } from "./components/GitLinkTable";
import { useEffect, useRef, useState } from "react";
import { useRequest } from "ahooks";
import { getAdminLabelDistributionsAll } from "@/services/csapi/label";
import { SyncOutlined } from "@ant-design/icons";

export default function () {
  const [layout, setLayout] = useState<"horizontal" | "vertical">("horizontal");

  const divRef = useRef<HTMLDivElement>(null);
  const tableRef = useRef<GitLinkTableAction>(null);

  const [selected, setSelected] = useState<API.DistributionPackageDTO | undefined>(undefined);

  const { data: distributions, loading, run } = useRequest(getAdminLabelDistributionsAll)
  const distributionOptions = distributions?.map((d) => ({
    label: d,
    value: d,
  })) || [];
  const [distribution, setDistribution] = useState<string | undefined>();

  useEffect(() => {
    if (distributions?.length) {
      setDistribution(distributions[0]);
    }
  }, [distributions]);

  useEffect(() => {
    tableRef.current?.refresh?.();
  }, [distribution, tableRef]);

  const scrollIntoSelected = () => {
    if (divRef.current && selected) {
      setTimeout(() => {
        // ant-table-row-selected
        divRef.current?.querySelector(`.ant-table-row-selected`)?.scrollIntoView({
          behavior: "smooth",
          block: "nearest",
          inline: "nearest",
        });
      }, 100);
    }
  }

  return <PageContainer breadcrumb={undefined} extra={[
    <Button key="refresh" icon={<SyncOutlined />} onClick={() => {
      run();
    }} />,
    <Select key="distribution" options={distributionOptions} value={distribution} onChange={setDistribution} style={{ width: 200 }} placeholder="选择分发包" loading={loading} />,
    <Segmented key="layout" options={[
      { label: "水平", value: "horizontal" },
      { label: "垂直", value: "vertical" },
    ]} value={layout} onChange={setLayout} />
  ]}>
    <Splitter layout={layout} style={{
      height: "calc(100vh - 168px)",
      gap: "12px",
    }}>
      <Splitter.Panel collapsible>
        <div className="overflow-auto" ref={divRef} >
          <GitLinkTable distribution={distribution} selected={selected} onSelected={setSelected} ref={tableRef} />
        </div>
      </Splitter.Panel>
      <Splitter.Panel collapsible>
        <div className="overflow-auto" >
          <Card>
            <GitLinkForm distribution={distribution} data={selected} onNext={async () => {
              await tableRef.current?.onNext?.();
              scrollIntoSelected();
            }} onPrev={async () => {
              await tableRef.current?.onPrev?.();
              scrollIntoSelected();
            }} onCancel={() => {
              setSelected(undefined);
            }} onRefresh={() => {
              tableRef.current?.refresh?.();
            }} />
          </Card>
        </div>
      </Splitter.Panel>
    </Splitter>



  </PageContainer>

}