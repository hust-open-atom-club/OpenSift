import { Alert, Button, Checkbox, Form, Input, Progress, Result, Segmented, Select, Spin, Steps, Table, Upload } from "antd";
import { useRef, useState } from "react";
import { FileExcelOutlined, InboxOutlined } from "@ant-design/icons"
import { parse, ParseResult } from "papaparse";

export default function Importer({
  service
}: {
  service?: (link: string) => Promise<void>;
}) {
  const [step, setStep] = useState(0);
  const [type, setType] = useState("upload");

  const [data, setData] = useState("");
  const [file, setFile] = useState<File>();
  const [parseResult, setParseResult] = useState<ParseResult<unknown>>();
  const [list, setList] = useState<string[]>();

  const [withHeader, setWithHeader] = useState(true);

  const [running, setRunning] = useState(false);
  const [success, setSuccess] = useState(0);
  const [fail, setFail] = useState(0);
  const needStop = useRef<boolean>(false);

  const stop = () => {
    needStop.current = true;
  }

  const run = async () => {
    if (!list || !service) return;
    needStop.current = false;
    setSuccess(0);
    setFail(0);
    setRunning(true);
    for (const l of list) {
      if (needStop.current) break;
      try {
        await service(l);
        setSuccess(x => x + 1);
      } catch {
        setFail(x => x + 1);
      }
    }
    setRunning(false);
  }

  const canPrevious = (() => {
    if (step === 0) return false;
    else if (step === 1) return true;
    else if (step === 2) return !running;
    return false;
  })();
  const previous = () => {
    if (step === 1) { setStep(0); }
    if (step === 2) { setStep(1); }
  };
  const canNext = (() => {
    if (step === 0) {
      if (type === 'upload') return !!file;
      else return !!data;
    }
    else if (step === 1) return !!list && list.length > 0;
    return false;
  })();
  const next = () => {
    if (step === 0) {
      let v: File | string | undefined = undefined;
      if (type === 'upload' && file) v = file;
      else if (type === 'input' && data) v = data;

      if (v) parse(v, {
        header: withHeader,
        complete(results, file) {
          setParseResult(results)
          setList(undefined);
          setStep(1);
        }
      })
    } else if (step === 1) {
      if (!!list && list.length > 0) setStep(2);
    }
  };

  return <div className="flex flex-col h-full">
    <Steps items={[
      { title: "导入文件" },
      { title: "确认数据" },
      { title: "送入队列" },
    ]} current={step} />

    <div className="grow my-4 overflow-auto">
      {
        step === 0 ? <div>
          <div className="mb-2">
            <Segmented options={[
              { label: "上传文件", value: "upload" },
              { label: "手动输入", value: "input" },
            ]} value={type} onChange={(v) => {
              setData("");
              setType(v);
            }} />
          </div>
          {
            type === "upload" ? <Upload.Dragger accept=".csv" showUploadList={false} maxCount={1} onChange={async (info) => {
              setFile(info.file.originFileObj);
            }}>
              {!!file ? <>
                <p className="ant-upload-drag-icon">
                  <FileExcelOutlined />
                </p>
                <p className="ant-upload-text">{file.name}</p>
                <p className="ant-upload-hint">重新拖拽或点击可传入其他文件</p>
              </> : <>
                <p className="ant-upload-drag-icon">
                  <InboxOutlined />
                </p>
                <p className="ant-upload-text">拖拽文件以上传</p>
                <p className="ant-upload-hint">仅支持 .csv 文件格式</p>
              </>}
            </Upload.Dragger> : <Input.TextArea rows={6} placeholder="输入 csv" value={data} onChange={x => setData(x.target.value)}></Input.TextArea>
          }
          <div className="mt-2">
            <Checkbox checked={withHeader} onChange={(v) => setWithHeader(v.target.checked)} >
              文件包含表头
            </Checkbox>
          </div>
        </div> : step === 1 ? <div>
          {
            parseResult?.errors && parseResult.errors.length > 0 && <Alert showIcon type="error" closable message={`解析时出现 ${parseResult.errors.length} 个错误`} description={<div>
              {parseResult.errors.map((x, k) => <div key={k}>
                {x.row && `${x.row} 行：`}{x.message}
              </div>)}
            </div>} />

          }
          <div className="my-2 font-bold">选择列</div>
          <Select<string | number> className="w-full" options={withHeader ? parseResult?.meta.fields?.map(x => ({
            label: x,
            value: x
          })) : Array.from({ length: (parseResult?.data?.[0] as any[]).length }, (_, i) => ({
            label: `第 ${i + 1} 列`,
            value: i
          }))} onChange={v => {
            setList(parseResult?.data.map(x => (x as any)[v]))
          }} />
          <div className="my-2 font-bold">数据预览</div>
          <Table rowKey={x => x} columns={[{ title: "link", render: (v, r) => r }]} dataSource={list} />
        </div> : step === 2 ? <div>
          <div className="font-bold">一共需要处理 {list?.length} 条数据。</div>
          <div>
            <Progress percent={(success + fail) / (list?.length || 1) * 100} showInfo={false} />
            <p>总共 {list?.length}，已完成 {success + fail}，成功 {success}，失败 {fail}</p>
          </div>
          <div className="mt-4">
            {
              running ? <Button onClick={stop}>停止</Button> : <Button onClick={run}>开始处理</Button>
            }
          </div>
        </div> : step === 101 ? <div>
          正在解析文件...
        </div> : <div>
        </div>
      }
    </div>
    <div className="flex justify-between">
      <Button disabled={!canPrevious} onClick={previous}>上一步</Button>
      <Button disabled={!canNext} onClick={next}>下一步</Button>
    </div>


  </div>

}