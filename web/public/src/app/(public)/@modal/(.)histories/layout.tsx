"use client";
import { Drawer } from "antd";
import { useRouter } from "next/navigation";

export default function ({ children }: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  return <Drawer width={1200} open onClose={() => {
    router.back()
  }} footer={null}>
    <div className="overflow-auto">
      {children}
    </div>
  </Drawer>


}