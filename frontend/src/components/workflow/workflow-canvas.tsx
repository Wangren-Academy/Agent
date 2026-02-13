"use client";

import { useCallback, useRef, DragEvent } from "react";
import { ReactFlow, Connection, addEdge, Node, Edge, NodeTypes, ReactFlowInstance } from "reactflow";
import { Background } from "@reactflow/background";
import { Controls } from "@reactflow/controls";
import { MiniMap } from "@reactflow/minimap";
import "@reactflow/core/dist/style.css";
import { useWorkflowStore } from "@/stores/workflow-store";
import { AgentNode } from "./agent-node";

const nodeTypes: NodeTypes = {
  agent: AgentNode,
};

export function WorkflowCanvas() {
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const { 
    nodes, 
    edges, 
    addNode, 
    addEdge: storeAddEdge, 
    setSelectedNode,
    workflow 
  } = useWorkflowStore();

  const onConnect = useCallback(
    (connection: Connection) => {
      if (connection.source && connection.target) {
        storeAddEdge(connection.source, connection.target);
      }
    },
    [storeAddEdge]
  );

  const onDragOver = useCallback((event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  }, []);

  const onDrop = useCallback(
    (event: DragEvent<HTMLDivElement>) => {
      event.preventDefault();

      const data = event.dataTransfer.getData("application/agent");
      if (!data) return;

      const { agentId, agentName } = JSON.parse(data);
      
      // Calculate position
      const bounds = reactFlowWrapper.current?.getBoundingClientRect();
      if (!bounds) return;

      const position = {
        x: event.clientX - bounds.left - 100,
        y: event.clientY - bounds.top - 50,
      };

      addNode({ id: agentId, name: agentName } as any, position);
    },
    [addNode]
  );

  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      setSelectedNode(node.id);
    },
    [setSelectedNode]
  );

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, [setSelectedNode]);

  const onEdgesChange = useCallback(
    (changes: any[]) => {
      // Handle edge removal
      changes.forEach((change) => {
        if (change.type === "remove") {
          // Edge removal is handled by the store
        }
      });
    },
    []
  );

  return (
    <div ref={reactFlowWrapper} className="h-full w-full">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onConnect={onConnect}
        onDragOver={onDragOver}
        onDrop={onDrop}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
        nodeTypes={nodeTypes}
        fitView
        snapToGrid
        snapGrid={[15, 15]}
        defaultEdgeOptions={{
          type: "smoothstep",
          animated: true,
        }}
      >
        <Background gap={15} size={1} />
        <Controls />
        <MiniMap 
          nodeColor={(node) => {
            if (node.type === "agent") return "#3b82f6";
            return "#94a3b8";
          }}
          maskColor="rgba(0, 0, 0, 0.1)"
        />
      </ReactFlow>

      {!workflow && nodes.length === 0 && (
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div className="text-center">
            <h2 className="text-xl font-semibold text-muted-foreground">
              开始创建工作流
            </h2>
            <p className="text-sm text-muted-foreground mt-2">
              从左侧面板拖拽智能体到画布
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
