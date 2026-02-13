"use client";

import { memo } from "react";
import { Handle, Position, NodeProps } from "@reactflow/core";
import { Bot, MoreVertical } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useWorkflowStore } from "@/stores/workflow-store";

interface AgentNodeData {
  agent_id: string;
  agent_name?: string;
  label?: string;
}

export const AgentNode = memo(({ id, data, selected }: NodeProps<AgentNodeData>) => {
  const { removeNode, executionSteps, currentExecution } = useWorkflowStore();
  const steps = executionSteps.get(id) || [];
  const isRunning = currentExecution?.status === "running";
  const hasOutput = steps.length > 0;

  return (
    <div
      className={`
        px-4 py-3 rounded-lg border-2 bg-card min-w-[200px] shadow-sm
        ${selected ? "border-primary" : "border-border"}
        ${isRunning ? "animate-pulse" : ""}
        ${hasOutput ? "border-green-500/50" : ""}
      `}
    >
      {/* Input Handle */}
      <Handle
        type="target"
        position={Position.Left}
        className="w-3 h-3 !bg-primary"
      />

      {/* Node Content */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="p-1.5 rounded-md bg-primary/10">
            <Bot className="h-4 w-4 text-primary" />
          </div>
          <div>
            <div className="font-medium text-sm">
              {data.agent_name || data.label || "Agent"}
            </div>
            <div className="text-xs text-muted-foreground">
              {steps.length > 0 ? `${steps.length} 步骤` : "待执行"}
            </div>
          </div>
        </div>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="h-6 w-6 p-0">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => {/* TODO: Edit node */}}>
              编辑配置
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => {/* TODO: Duplicate */}}>
              复制节点
            </DropdownMenuItem>
            <DropdownMenuItem 
              onClick={() => removeNode(id)}
              className="text-destructive"
            >
              删除节点
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Status Indicator */}
      {hasOutput && (
        <div className="mt-2 pt-2 border-t border-border">
          <div className="text-xs text-muted-foreground">
            最后输出: {steps[steps.length - 1]?.output?.substring(0, 50)}...
          </div>
        </div>
      )}

      {/* Output Handle */}
      <Handle
        type="source"
        position={Position.Right}
        className="w-3 h-3 !bg-primary"
      />
    </div>
  );
});

AgentNode.displayName = "AgentNode";
