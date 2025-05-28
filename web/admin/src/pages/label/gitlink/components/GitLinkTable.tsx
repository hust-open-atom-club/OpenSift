import { getAdminLabelDistributions, getAdminLabelDistributionsAll } from "@/services/csapi/label";
import { ProTable, ProColumns, ProFormInstance, ActionType } from "@ant-design/pro-components";
import { useRequest } from "ahooks";
import { Button } from "antd";
import React, { useEffect, useImperativeHandle, useRef } from "react";

type Props = {
  distribution?: string;
  selected?: API.DistributionPackageDTO;
  onSelected?: (data?: API.DistributionPackageDTO) => void;
}

type Action = {
  onNext?: () => void | Promise<void>;
  onPrev?: () => void | Promise<void>;
  refresh?: () => void | Promise<void>;
}

export type GitLinkTableAction = Action;

export default React.forwardRef<Action, Props>(({ distribution, selected, onSelected }: Props, ref) => {
  const form = useRef<ProFormInstance>();
  const { data, params, runAsync, loading, refreshAsync } = useRequest<{
    success: boolean;
    total?: number;
    data?: API.DistributionPackageDTO[];
  }, {
    confidence?: number;
    search?: string;
    pageSize?: number;
    current?: number;
    keyword?: string;
  }[]>(async (params) => {
    if (!distribution) {
      return {
        data: [],
        success: true,
      }
    }
    const d = await getAdminLabelDistributions({
      distribution: distribution,
      confidence: params?.confidence,
      search: params?.search,
      skip: ((params?.current || 1) - 1) * (params?.pageSize || 20),
      take: params?.pageSize || 20,
    })
    return {
      data: d.items,
      success: true,
      total: d.total,
    }
  }, {
    manual: true,
    defaultParams: [{
      confidence: 0,
      search: '',
      pageSize: 20,
      current: 1,
    }],
    refreshDeps: [distribution],
  });

  useImperativeHandle(ref, () => ({
    refresh: async () => { await refreshAsync() },
    onNext: async () => {
      // select next item
      if (!selected) return;
      const index = data?.data?.findIndex(item => item.package === selected.package);
      if (index === undefined || index === -1) return;
      if (index + 1 < (data?.data?.length || 0)) {
        onSelected?.(data?.data?.[index + 1]);
      } else {
        try {
          const d = await runAsync({
            ...params?.[params.length - 1],
            current: (params?.[params.length - 1]?.current || 1) + 1,
          });
          onSelected?.(d?.data?.[0]);
        } catch (e) {
          console.error("Reload failed", e);
        }
      }
    },
    onPrev: async () => {
      // select previous item
      if (!selected) return;
      const index = data?.data?.findIndex(item => item.package === selected.package);
      if (index === undefined || index === -1) return;
      if (index - 1 >= 0) {
        onSelected?.(data?.data?.[index - 1]);
      } else {
        if (params?.[params.length - 1]?.current === 1) return;
        try {
          const d = await runAsync({
            ...params?.[params.length - 1],
            current: Math.max((params?.[params.length - 1]?.current || 1) - 1, 1),
          });
          onSelected?.(d?.data?.[d?.data?.length - 1]);
        } catch (e) {
          console.error("Reload failed", e);
        }
      }
    },
  }));

  const columns: ProColumns<API.DistributionPackageDTO>[] = [
    { title: '置信度', hideInTable: true, key: 'confidence', valueType: 'select', valueEnum: { 0: "全部", 1: "不为1", 2: "为1" } },
    { title: '包名', hideInTable: true, dataIndex: 'search', valueType: 'text' },

    { title: '包名', hideInSearch: true, dataIndex: 'package', key: 'package' },
    { title: '描述', hideInSearch: true, dataIndex: 'description', key: 'description', width: '20%' },
    { title: '主页', hideInSearch: true, dataIndex: 'homePage', key: 'homepage' },
    { title: 'Git Link', hideInSearch: true, dataIndex: 'gitLink', key: 'git_link', width: '25%' },
    { title: '置信度', hideInSearch: true, dataIndex: 'linkConfidence', key: 'linkConfidence', width: '10%' },
    {
      title: '操作',
      hideInSearch: true,
      key: 'action',
      render: (_, record) => (
        <Button type="link">编辑</Button>
      ),
    },

  ]

  return <ProTable<API.DistributionPackageDTO, {
    distribution?: string;
    confidence?: number;
    search?: string;
  }>
    columns={columns}
    formRef={form}
    scroll={{ x: 1000 }}
    rowKey="package"
    // actionRef={table}
    rowSelection={{
      onChange: (_, rows) => {
        onSelected?.(rows[0]);
      },
      type: 'radio',
      selectedRowKeys: selected?.package ? [selected.package] : [],
    }}
    onRow={(record) => ({
      onClick: () => {
        onSelected?.(record);
      },
    })}
    dataSource={data?.data || []}
    pagination={{
      showSizeChanger: true,
      showQuickJumper: true,
      pageSizeOptions: ['20', '50', '100'],
      pageSize: params?.[params.length - 1]?.pageSize || 20,
      current: params?.[params.length - 1]?.current || 1,
      total: data?.total || 0,
      onChange: (page, pageSize) => {
        runAsync({
          ...params?.[params.length - 1],
          current: page,
          pageSize,
        });
      },
    }}
    onSubmit={(values) => {
      runAsync({
        ...values,
        confidence: values.confidence,
        search: values.search,
        current: 1,
      });
    }}
    loading={loading}

  >
  </ ProTable>
})