import { getHistories } from "@/service/client";
import { Result } from "antd";
import Results from "./Results";

export default async function Page({
  searchParams
}: {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}) {
  const link = (await searchParams)['link'];
  const start = (await searchParams)['start'];
  const take = (await searchParams)['take'];
  let data;
  if (typeof link !== "string" ||
    typeof start === "object" ||
    typeof take === "object") {
    data = undefined;
  } else {
    data = await getHistories({
      query: {
        link: link,
        start: start ? parseInt(start) : 0,
        take: take ? parseInt(take) : 10
      }
    })
  }

  if (!data?.data) {
    return <Result status="500" title="No Results Found"
      subTitle="Maybe server error occurs"
      extra={<a href="/">Back Home</a>}
    />
  }

  return <Results items={data.data} />

}


