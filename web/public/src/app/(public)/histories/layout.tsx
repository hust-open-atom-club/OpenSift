import Nav from "@/components/Nav";
import React from "react";

type Props = {
  children: React.ReactNode;
}

export default function ({
  children
}: Props) {
  return <div>
    <Nav />
    {children}
  </div>
}