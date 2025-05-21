"use client";
import { Button, Input } from "antd";
import { SearchOutlined } from "@ant-design/icons";
import { useCallback, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import useMessage from "antd/es/message/useMessage";

export default function Home() {

  const [search, setSearch] = useState('');

  const [message, holder] = useMessage();

  const router = useRouter()


  const onSearch = () => {
    if (!search) {
      message.error('Please enter a search query')
      return
    }
    // set search query to the URL
    router.push(`results?q=${search}`)
  }

  return (
    <div className="mx-auto my-16 max-w-5xl">
      {holder}
      <img src="/logo.svg" alt="logo" className="h-32 mx-auto" />

      <div className="pt-24 px-8">
        <div className="flex items-center justify-center">
          <Input size="large" className="grow !text-2xl !py-4 !px-8 !rounded-l-full"
            placeholder="Search Git Repos..." value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && onSearch()}
          />
          <Button type="primary" size="large"
            className="!rounded-r-full !h-16 !text-2xl !py-4 !px-8"
            onClick={onSearch}
          >
            <SearchOutlined />
          </Button>
        </div>
      </div>

    </div>
  );
}
