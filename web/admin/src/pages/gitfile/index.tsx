"use client";
import { PageContainer } from "@ant-design/pro-layout";
import GitFileSummary from "./components/Summary";
import GitFileTable from "./components/Table";
import { Card } from "antd";

export default function () {
  return <PageContainer>
    <GitFileSummary />
    <Card variant="borderless">
      <GitFileTable />
    </Card>
  </PageContainer>
}