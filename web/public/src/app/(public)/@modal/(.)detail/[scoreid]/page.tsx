import { getResultsByScoreid } from "@/service/client";
import { Result } from "antd";
import { RepoDetail } from "@/app/(public)/detail/[scoreid]/RepoDetail";

export default async function Page({
  params
}: {
  params: Promise<{ scoreid: string }>
}) {

  const scoreidStr = (await params).scoreid;
  const scoreid = parseInt(scoreidStr);

  if (isNaN(scoreid)) {
    return <Result status="404" title="Invalid Score ID"
      subTitle="Score ID must be an integer"
      extra={<a href="/">Back Home</a>}
    />
  }

  const data = await getResultsByScoreid({
    path: {
      scoreid
    }
  })

  if (!data.data) {
    return <Result status="500" title="No Results Found"
      subTitle="Maybe server error occurs"
      extra={<a href="/">Back Home</a>}
    />
  }

  return <RepoDetail data={data.data} />

}

