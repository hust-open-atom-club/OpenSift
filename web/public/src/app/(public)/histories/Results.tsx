"use client";

import { getHistories, ModelPageDtoModelResultDto } from "@/service/client";
import { Result } from "antd";
import { Pagination } from "antd";
import { useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Loading from "../results/loading";
import ScoreCard from "@/components/ScoreCard";

export default function Results(props: {
  items?: ModelPageDtoModelResultDto
}) {
  const init = useRef(true);
  const link = useSearchParams().get('link') || '';
  const q = useSearchParams().get('q') || '';
  const router = useRouter();

  const [data, setData] = useState(props.items);
  const [loading, setLoading] = useState(false);
  const [pageSize, setPageSize] = useState(props.items?.count || 10);
  const [page, setPage] = useState(1);

  const getItems = async (link: string, page: number, pageSize: number) => {
    // set search query to the URL
    setLoading(true);
    try {
      const data = await getHistories({
        query: {
          link,
          start: (page - 1) * pageSize,
          take: pageSize
        }
      });
      setData(data.data);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (init.current) {
      init.current = false;
      return
    }

    setPage(1);
    getItems(link, 1, pageSize);
  }, [link]);

  const onPageChange = async (page: number, pageSize: number) => {
    setPage(page);
    setPageSize(pageSize);
    router.push(`histories?link=${link}&take=${pageSize}&start=${(page - 1) * pageSize}`)
    getItems(link, page, pageSize);
  }

  let content = null;
  if (loading) {
    return <Loading />
  } else if (data?.items) {
    content = data.items.map((item) => (
      <ScoreCard key={item.scoreID} item={item} hideHistory keepSearchParams={{
        link: link,
        q: q
      }} />
    ));
  } else {
    content = <Result
      status="404"
      title="No Results Found"
      subTitle="Sorry, we couldn't find any results for your search."
    />
  }

  return (
    <div className="max-w-screen-xl mx-auto my-8">
      <h2 className="mx-8 text-2xl font-bold mb-4">Histories of  <span className="text-blue-600">{link} </span> </h2>
      <div className="mx-8">
        {content}
      </div>

      <div className="flex justify-center">
        <Pagination defaultCurrent={1} total={data?.total} showSizeChanger current={page} pageSize={pageSize}
          onChange={onPageChange}
        />
      </div>
    </div>
  )
}