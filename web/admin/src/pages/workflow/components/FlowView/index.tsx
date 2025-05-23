import { Button } from "antd";
import Canvas, { CanvasAction, TaskNode } from "./Canvas";
import React from "react";


type Props = {
  data?: TaskNode[];
  onSelect?: (node?: TaskNode) => void;
}

export default function ({
  data, onSelect
}: Props) {

  const ref = React.useRef<CanvasAction>(null);

  return <div className="h-full w-full relative">
    <Canvas data={data} onSelect={onSelect} ref={ref} />
    {/* right top corner */}
    <div className="absolute top-0 right-0 p-2 opacity-35 hover:opacity-100 transition-opacity duration-200">
      <Button onClick={() => ref.current?.relayout()}>重置布局</Button>
    </div>
  </div>

}