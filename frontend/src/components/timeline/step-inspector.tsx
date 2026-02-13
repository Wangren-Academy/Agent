"use client";

import { useWorkflowStore } from "@/stores/workflow-store";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { 
  FileCode, 
  MessageSquare, 
  Timer,
  Coins,
  Bot
} from "lucide-react";

export function StepInspector() {
  const { selectedNodeId, nodes, agents, executionSteps } = useWorkflowStore();

  if (!selectedNodeId) {
    return (
      <div className="flex-1 flex items-center justify-center text-muted-foreground text-sm p-4">
        <div className="text-center">
          <FileCode className="h-8 w-8 mx-auto mb-2 opacity-50" />
          <p>选择一个节点查看详情</p>
        </div>
      </div>
    );
  }

  const node = nodes.find((n) => n.id === selectedNodeId);
  const steps = executionSteps.get(selectedNodeId) || [];
  const agent = agents.find((a) => a.id === node?.data?.agent_id);

  if (!node) {
    return null;
  }

  const totalTokens = steps.reduce((sum, s) => sum + (s.tokens || 0), 0);
  const totalLatency = steps.reduce((sum, s) => sum + (s.latency_ms || 0), 0);

  return (
    <ScrollArea className="flex-1">
      <div className="p-4 space-y-4">
        {/* Node Info */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Bot className="h-5 w-5 text-primary" />
            <h3 className="font-semibold">
              {node.data?.agent_name || "Agent Node"}
            </h3>
          </div>
          
          <div className="flex flex-wrap gap-2">
            <Badge variant="secondary">
              Node ID: {selectedNodeId.substring(0, 8)}...
            </Badge>
            {agent?.model_config?.provider && (
              <Badge variant="outline">
                {agent.model_config.provider}
              </Badge>
            )}
            {agent?.model_config?.model && (
              <Badge variant="outline">
                {agent.model_config.model}
              </Badge>
            )}
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 gap-3">
          <div className="p-3 rounded-lg border bg-background">
            <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
              <Coins className="h-3 w-3" />
              Tokens
            </div>
            <div className="text-lg font-semibold">{totalTokens}</div>
          </div>
          <div className="p-3 rounded-lg border bg-background">
            <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
              <Timer className="h-3 w-3" />
              延迟
            </div>
            <div className="text-lg font-semibold">{totalLatency}ms</div>
          </div>
        </div>

        {/* Agent Config */}
        {agent && (
          <div className="space-y-2">
            <h4 className="text-sm font-medium flex items-center gap-2">
              <MessageSquare className="h-4 w-4" />
              System Prompt
            </h4>
            <pre className="text-xs bg-muted p-3 rounded-lg overflow-x-auto whitespace-pre-wrap">
              {agent.system_prompt}
            </pre>
          </div>
        )}

        {/* Steps List */}
        {steps.length > 0 && (
          <div className="space-y-2">
            <h4 className="text-sm font-medium">执行步骤 ({steps.length})</h4>
            <div className="space-y-2">
              {steps.map((step, index) => (
                <div 
                  key={step.step_id}
                  className="p-3 rounded-lg border bg-background text-xs"
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium">
                      {index + 1}. {step.type}
                    </span>
                    {step.tokens && (
                      <span className="text-muted-foreground">
                        {step.tokens} tokens
                      </span>
                    )}
                  </div>
                  {step.output && (
                    <p className="text-muted-foreground line-clamp-3">
                      {step.output}
                    </p>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Node Position */}
        <div className="text-xs text-muted-foreground">
          Position: ({Math.round(node.position.x)}, {Math.round(node.position.y)})
        </div>
      </div>
    </ScrollArea>
  );
}
