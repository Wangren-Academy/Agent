"use client";

import { useEffect } from "react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useWorkflowStore } from "@/stores/workflow-store";
import { useExecutionWebSocket } from "@/lib/websocket";
import { 
  Play, 
  Pause, 
  RotateCcw, 
  SkipForward, 
  Clock,
  Zap
} from "lucide-react";
import { TimelineStep } from "./timeline-step";
import { StepInspector } from "./step-inspector";

export function TimelinePanel() {
  const { 
    currentExecution, 
    executionSteps, 
    selectedNodeId,
    replayMode,
    toggleReplayMode,
    replayExecution,
    loadExecution
  } = useWorkflowStore();

  const { connected } = useExecutionWebSocket(
    currentExecution?.status === "running" ? currentExecution.id : null
  );

  const allSteps = Array.from(executionSteps.entries()).flatMap(([nodeId, steps]) =>
    steps.map((step) => ({ ...step, nodeId }))
  );

  return (
    <div className="w-96 border-l bg-card flex flex-col">
      <div className="h-12 flex items-center justify-between px-4 border-b">
        <div className="flex items-center gap-2">
          <Clock className="h-4 w-4 text-primary" />
          <span className="font-medium">执行时序图</span>
        </div>
        {currentExecution && (
          <div className="flex items-center gap-1">
            {connected && (
              <span className="flex items-center gap-1 text-xs text-green-500">
                <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                实时
              </span>
            )}
            <span className="text-xs text-muted-foreground">
              {currentExecution.status}
            </span>
          </div>
        )}
      </div>

      {/* Replay Controls */}
      {currentExecution && currentExecution.status !== "running" && (
        <div className="p-2 border-b flex items-center gap-2">
          <Button 
            size="sm" 
            variant={replayMode ? "default" : "outline"}
            onClick={toggleReplayMode}
          >
            {replayMode ? (
              <>
                <Zap className="h-3 w-3 mr-1" />
                沙盘模式
              </>
            ) : (
              <>
                <RotateCcw className="h-3 w-3 mr-1" />
                重放
              </>
            )}
          </Button>
          
          {replayMode && (
            <Button size="sm" onClick={replayExecution}>
              <Play className="h-3 w-3 mr-1" />
              应用修改
            </Button>
          )}
        </div>
      )}

      <Tabs defaultValue="timeline" className="flex-1 flex flex-col">
        <TabsList className="grid w-full grid-cols-2 m-2">
          <TabsTrigger value="timeline">时序图</TabsTrigger>
          <TabsTrigger value="inspector">详情</TabsTrigger>
        </TabsList>

        <TabsContent value="timeline" className="flex-1 mt-0">
          <ScrollArea className="h-full">
            <div className="p-4">
              {allSteps.length === 0 ? (
                <div className="text-center text-muted-foreground text-sm py-8">
                  <Clock className="h-8 w-8 mx-auto mb-2 opacity-50" />
                  <p>暂无执行记录</p>
                  <p className="text-xs mt-1">执行工作流后将在此显示时序图</p>
                </div>
              ) : (
                <div className="relative">
                  {allSteps.map((step, index) => (
                    <TimelineStep
                      key={step.step_id}
                      step={step}
                      nodeId={step.nodeId}
                      isLast={index === allSteps.length - 1}
                    />
                  ))}
                </div>
              )}
            </div>
          </ScrollArea>
        </TabsContent>

        <TabsContent value="inspector" className="flex-1 mt-0">
          <StepInspector />
        </TabsContent>
      </Tabs>

      {/* Execution Meta */}
      {currentExecution?.snapshot?.execution_meta && (
        <div className="p-4 border-t text-xs text-muted-foreground">
          <div className="grid grid-cols-3 gap-2">
            <div>
              <span className="block font-medium">Tokens</span>
              <span>{currentExecution.snapshot.execution_meta.total_tokens}</span>
            </div>
            <div>
              <span className="block font-medium">耗时</span>
              <span>{currentExecution.snapshot.execution_meta.duration_ms}ms</span>
            </div>
            <div>
              <span className="block font-medium">成本</span>
              <span>${currentExecution.snapshot.execution_meta.total_cost.toFixed(4)}</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
