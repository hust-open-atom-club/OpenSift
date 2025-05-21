"use client";

import { getResults, ModelPageDtoModelResultDto } from "@/service/client";
import { Result, Skeleton } from "antd";
import { ForkOutlined, GithubOutlined } from "@ant-design/icons";
import { Col, Pagination, Row, Statistic } from "antd";
import { useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import Loading from "./loading";
import ScoreCard from "@/components/ScoreCard";

export default function Results(props: {
  items?: ModelPageDtoModelResultDto
}) {
  const init = useRef(true);
  const search = useSearchParams().get('q') || '';
  const router = useRouter();

  const [data, setData] = useState(props.items);
  const [loading, setLoading] = useState(false);
  const [pageSize, setPageSize] = useState(props.items?.count || 10);
  const [page, setPage] = useState(1);

  const getItems = async (q: string, page: number, pageSize: number) => {
    // set search query to the URL
    setLoading(true);
    try {
      const data = await getResults({
        query: {
          q,
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
    getItems(search, 1, pageSize);
  }, [search]);

  const onPageChange = async (page: number, pageSize: number) => {
    setPage(page);
    setPageSize(pageSize);
    router.push(`results?q=${search}&take=${pageSize}&start=${(page - 1) * pageSize}`)
    getItems(search, page, pageSize);
  }


  let content = null;
  if (loading) {
    return <Loading />
  } else if (data?.items) {
    content = data.items.map((item) => (
      <ScoreCard key={item.scoreID} item={item} keepSearchParams={{
        q: search
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