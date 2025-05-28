import { putAdminLabelDistributionsGitlink } from '@/services/csapi/label';
import { ProForm, ProFormItem, ProFormSlider, ProFormText, ProFormTextArea } from '@ant-design/pro-components';
import { App, Button, Form, Result, Slider, Space } from 'antd';
import React, { useEffect } from 'react';

type Props = {
  distribution?: string;
  data?: API.DistributionPackageDTO;
  onCancel?: () => void;
  onRefresh?: () => void;
  onNext?: () => void;
  onPrev?: () => void;
}

export default function GitLinkForm({ distribution, data, onCancel, onRefresh, onNext, onPrev }: Props) {
  const [form] = ProForm.useForm<API.DistributionPackageDTO>();
  useEffect(() => {
    if (data) {
      form.setFieldsValue(data);
    }
  }, [data, form]);

  if (!data) {
    return <Result status="404" title="没有数据" subTitle="请先从表格中选择一个包。" />;
  }

  const { message, modal } = App.useApp();

  const onSubmit = async (values: API.DistributionPackageDTO) => {
    if (!distribution || !values.package) return;
    console.log("提交数据", values);
    await putAdminLabelDistributionsGitlink({
      confidence: values.linkConfidence,
      distribution: distribution,
      link: values.gitLink,
      packageName: values.package,
    });
    message.success("更新成功");
    onRefresh?.();
  };

  const confirmWrapper = (f?: () => void) => async () => {
    const d1 = form.getFieldsValue();
    const d2 = data;
    if (d1.gitLink !== d2?.gitLink ||
      d1.linkConfidence !== d2?.linkConfidence
    ) {
      const confirm = await modal.confirm({
        title: '确认',
        content: '您有未保存的更改，是否继续切换？',
        okText: '放弃更改并继续',
        cancelText: '取消',
      });
      if (confirm) {
        f?.();
      }
    } else {
      f?.();
    }
  };

  return (<ProForm<API.DistributionPackageDTO>
    form={form}
    // initialValues={data}
    onFinish={onSubmit}
    submitter={{
      render: (props, doms) => [
        <Button key="prev" onClick={confirmWrapper(onPrev)}>
          上一项
        </Button>,
        ...doms,
        <Button key="submitNext" color='green' variant='solid' onClick={async () => {
          await onSubmit(form.getFieldsValue());
          onNext?.();
        }}>
          提交并下一项
        </Button>,
        <Button key="next" onClick={confirmWrapper(onNext)}>
          下一项
        </Button>,
        <Button key="cancel" onClick={onCancel} >
          取消
        </Button>
      ],
    }}
    layout="vertical"
    grid={true}
  >
    {/* <ProFormText name="distribution" label="发行版本" width="md" readonly /> */}
    <ProFormText name="package" label="包名" width="md" readonly />
    <ProFormTextArea name="description" label="描述" width="md" readonly />
    <ProFormText name="homePage" label="主页" width="md" readonly />

    <ProFormText name="gitLink" label="Git Link" width="md" fieldProps={{
      addonAfter: <Button size='small' type='link' onClick={(e) => {
        e.preventDefault();
        form.setFieldsValue({ gitLink: "NA" });
      }}>设为NA</Button>
    }} />

    <ProForm.Item label="置信度" tooltip="置信度是指对该 Git Link 的可信程度，1 表示完全确信，0 表示完全不确信。" shouldUpdate>
      {
        () => (
          <Space>
            <Form.Item name="linkConfidence" noStyle>
              <Slider min={0} max={1} step={0.01} style={{ width: 200 }} />
            </Form.Item>
            <Button type='primary' size='small' onClick={() => {
              form.setFieldsValue({ linkConfidence: 1 });
            }}>设为确信</Button>
            <Button size='small' onClick={() => {
              // @ts-ignore
              form.setFieldsValue({ linkConfidence: null });
            }}>设为NULL</Button>
          </Space>
        )
      }
    </ProForm.Item>


  </ProForm>)
}