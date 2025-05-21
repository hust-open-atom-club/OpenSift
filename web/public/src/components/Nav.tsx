"use client";
import Search from "antd/es/input/Search"
import Link from "next/link"

import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";

export default function Nav() {

  const router = useRouter();
  const searchParams = useSearchParams();

  const [search, setSearch] = useState(searchParams.get('q') || '');
  return <Suspense>
    <div className="shadow-md bg-white sticky top-0 z-50">
      <div className="flex py-4 px-8 items-center max-w-screen-xl mx-auto z-50">
        <Link href="/">
          <img src="/logo.svg" alt="logo" className="h-8" />
        </Link>
        <Search enterButton size="large" className="pl-8 !max-w-2xl" placeholder="Search Git Repos..." value={search} onChange={e => {
          setSearch(e.target.value)
        }} onSearch={s => {
          router.push(`/results?q=${s}`)
        }} />
      </div>
    </div>
  </Suspense>
}