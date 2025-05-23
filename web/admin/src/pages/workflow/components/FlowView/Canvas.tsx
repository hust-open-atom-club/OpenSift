import React, { useEffect, useRef } from "react";
import { Graph, GraphData } from "@antv/g6";
// import { useWhyDidYouUpdate } from "ahooks";

export type TaskNode = {
  name: string;
  title: string;
  description: string;
  args: string;
  status: 'pending' | 'running' | 'success' | 'failed';
  type: string;
  dependencies: string[];
  startTime?: string;
  endTime?: string;
}

export type CanvasAction = {
  relayout: () => void;
}

type Props = {
  data?: TaskNode[];
  onSelect?: (node?: TaskNode) => void;
}

function getNodeSize(): [number, number] {
  return [48 * 4, 16 * 4];
}

function getNodeHTML(data?: TaskNode, active?: boolean) {
  const statusStr = data?.status === 'pending' ? '待执行' : data?.status === 'running' ? '执行中' : data?.status === 'success' ? '成功' : data?.status === 'failed' ? '失败' : '';
  var borderClass = data?.status === 'pending' ? 'border-gray-500' : data?.status === 'running' ? 'border-blue-500' : data?.status === 'success' ? 'border-green-500' : data?.status === 'failed' ? 'border-red-500' : '';

  if (active === true) {
    borderClass = 'border-orange-500 outline outline-orange-200 outline-4 bg-orange-50';
  } else if (active === false) {
    borderClass = 'border-gray-200';
  }

  // <div class="flex justify-end gap-2 mt-2">
  //   <button class="px-4 py-1 text-xs font-normal bg-white border border-gray-300 rounded-md shadow-sm hover:text-blue-600 hover:border-blue-600 focus:outline-none focus:border-blue-600 focus:ring-2 focus:ring-blue-500/20 transition-colors">配置</button>
  //   <button class="px-4 py-1 text-xs font-normal bg-white border border-gray-300 rounded-md shadow-sm hover:text-blue-600 hover:border-blue-600 focus:outline-none focus:border-blue-600 focus:ring-2 focus:ring-blue-500/20 transition-colors">输出</button>
  // </div>

  return `<div class="w-48 h-16 border flex justify-center flex-col border-l-8 ${borderClass} rounded-md shadow-sm bg-white p-2">
            <h3 class="text-sm font-bold">${data?.title}</h3>
            <div class="text-xs">当前状态：${statusStr}</div>
          </div>`

}


export default React.forwardRef<CanvasAction, Props>((props, ref) => {
  const { data, onSelect } = props;
  const canvasRef = useRef<HTMLDivElement>(null);
  const graphRef = useRef<Graph | null>(null);

  React.useImperativeHandle(ref, () => ({
    relayout: () => {
      const graph = graphRef.current;
      if (!graph) return;
      graph.layout();
      graph.fitView();
    }
  }))

  // useWhyDidYouUpdate('Canvas', props);

  const dataToGraphData = (data?: TaskNode[]): GraphData => {
    if (!data) return { nodes: [], edges: [] };

    const nodes = data.map((item, index) => {
      return {
        id: item.name,
        label: item.name,
        data: item,
      }
    }) as GraphData['nodes'];

    const edges = data.map((item, index) => ({
      ss: item.dependencies,
      t: item.name
    })).flatMap((item) => item.ss.map((s) => ({
      source: s,
      target: item.t,
    }))) as GraphData['edges'];

    return {
      nodes,
      edges
    }
  }

  const handleOnClick = (e: any) => {
    let selected = undefined;
    if (e && e.targetType === 'node' && e.target?.id) {
      selected = data?.find((item => item.name === e.target.id));
    }
    onSelect?.(selected);
  };



  useEffect(() => {
    if (!graphRef.current) return;
    const graph = graphRef.current;
    graph.clear();
    graph.addData(dataToGraphData(data));
    graph.render();
  }, [data]);

  useEffect(() => {
    const observer = new ResizeObserver(() => {
      const canvas = canvasRef.current;
      const graph = graphRef.current;
      if (!canvas || !graph || graph.destroyed) return;
      graph.resize();
      // graph.fitView();
    });
    if (canvasRef.current) {
      observer.observe(canvasRef.current);
    }
    return () => {
      observer.disconnect();
    };
  }, [])


  useEffect(() => {
    const graph = new Graph({
      container: canvasRef.current!,
      autoFit: {
        type: 'view',
        options: {
          when: 'overflow',
          direction: 'both',
        },
      },
      behaviors: ['drag-canvas', 'zoom-canvas', 'drag-element',
        {
          type: 'click-select',
          enable: (event: any) => ['node', 'canvas'].includes(event.targetType),
          key: 'click-select-1',
          state: 'active',
          neighborState: 'neighborActive',
          unselectedState: 'inactive',
          onClick: handleOnClick,
        },
      ],
      data: dataToGraphData(data),
      node: {
        type: "html",
        state: {
          active: {
            innerHTML: ({ data }: { data?: TaskNode }) => getNodeHTML(data, true),
          },
          inactive: {
            innerHTML: ({ data }: { data?: TaskNode }) => getNodeHTML(data, false),
          }
        },
        style: {
          // port: true,
          // portR: 3,
          // portLineWidth: 1,
          // portStroke: '#fff',
          ports: [
            { key: 'top', placement: 'top', },
            { key: 'right', placement: 'right', },
            { key: 'bottom', placement: 'bottom', },
            { key: 'left', placement: 'left', linkToCenter: false },
          ],
          size: getNodeSize(),
          innerHTML: ({ data }: {
            data?: TaskNode
          }) => getNodeHTML(data),
        },
        // palette: {
        //   field: 'group',
        //   color: 'tableau',
        // },
      },
      layout: {
        type: 'dagre',
        rankdir: 'LR',
        nodesep: 40,
        ranksep: 80,
      },
      edge: {
        type: 'cubic-horizontal',
        style: {
          // labelText: (d) => d.id,
          // labelBackground: true,
          stroke: (d) => {
            const targetName = d.target;
            const targetNode = data?.find((item) => item.name === targetName);
            if (!targetNode) return '#ccc';
            return targetNode.status === 'pending' ? '#7E92B5' : targetNode.status === 'running' ? '#2e67c3' : targetNode.status === 'success' ? '#199245' : targetNode.status === 'failed' ? '#FF3D00' : '#ccc';
          },
          endArrow: true,
        },
      },

    });
    graphRef.current = graph;
    graph.render();

    return () => {
      graphRef.current = null;
      graph.destroy();
    };
  }, []);

  // useEffect(() => {
  //   const canvas = canvasRef.current;
  //   const graph = graphRef.current;

  //   if (!canvas || !graph || graph.destroyed) return;

  //   graph.render();
  // });

  return <div className="w-full h-full" ref={canvasRef} />;
})