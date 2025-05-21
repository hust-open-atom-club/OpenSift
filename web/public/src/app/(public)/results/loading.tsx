import { Skeleton } from "antd";

export default function Loading() {
  return (
    <div className="max-w-screen-xl mx-auto my-8">
      {
        Array.from({ length: 3 }).map((_, index) => (
          <div key={index} className="my-8 p-8 w-full rounded-xl bg-white border-solid border border-slate-200 shadow-sm">
            <Skeleton active />
          </div>
        ))
      }
    </div>
  );
}