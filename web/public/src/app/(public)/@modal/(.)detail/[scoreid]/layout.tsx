"use client";
import { Modal } from "antd";
import { useRouter } from "next/navigation";

export default function ({ children }: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  return <Modal width={1200} open onCancel={() => {
    router.back()
  }} footer={null}>
    <div className="overflow-auto">
      {children}
    </div>
  </Modal>


}