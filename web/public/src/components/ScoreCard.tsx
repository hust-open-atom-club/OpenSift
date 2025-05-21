import { ModelResultDto } from "@/service/client"
import { Col, Pagination, Row, Statistic } from "antd";

import { ForkOutlined, GithubOutlined } from "@ant-design/icons";
import Link from "next/link";

type Props = {
  item: ModelResultDto
  keepSearchParams?: { [key: string]: string }
  hideHistory?: boolean
};

export default function ScoreCard(
  { item, keepSearchParams, hideHistory }: Props
) {
  keepSearchParams = keepSearchParams || {};
  const detailParams = new URLSearchParams(keepSearchParams);
  const historyParams = new URLSearchParams({
    ...keepSearchParams,
    link: item.link || ""
  });


  return <div className="my-8 p-8 w-full rounded-xl bg-white border-solid border border-slate-200 shadow-sm">
    <div className="flex items-center gap-2">
      {item.link?.indexOf("github") !== -1 ? <GithubOutlined /> : <ForkOutlined />}
      <span className="font-bold">{item.link}</span>
    </div>

    {
      item.updateTime && <div className="mt-4 text-gray-600">
        Updated at {new Date(item.updateTime).toLocaleString()}
      </div>
    }

    {item.scoreID == null ? <div className="mt-4 text-red-600">
      No Score Data Found
    </div> : <>
      <Row gutter={16} className="mt-4">
        <Col span={4}>
          <Statistic title="Total Score" value={item.score} precision={4} />
        </Col>
        <Col span={4}>
          <Statistic title="Git Metadata" value={item.gitScore} precision={4} />
        </Col>
        <Col span={4}>
          <Statistic title="Lang Ecosystem" value={item.langScore} precision={4} />
        </Col>
        <Col span={4}>
          <Statistic title="Distributions" value={item.distroScore} precision={4} />
        </Col>
      </Row>

      <div className="mt-4 flex gap-4">
        <Link href={`/detail/${item.scoreID}?${detailParams}`} className="text-blue-800" passHref>Details</Link>
        {!hideHistory && <Link href={`/histories?${historyParams}`} className="text-blue-800">Histories</Link>}
      </div>
    </>
    }
  </div>


}