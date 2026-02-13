"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import type { Step } from "@/lib/types";
import { useWorkflowStore } from "@/stores/workflow-store";
import { 
  ChevronDown, 
  ChevronUp, 
  Brain, 
  Wrench, 
  FileOutput,
  Edit2,
  Check
} from "lucide-react";
import { cn } from "@/lib/utils";

interface TimelineStepProps {
  step: Step;
  nodeId: string;
  isLast: boolean;
}

export function TimelineStep({ step, nodeId, isLast }: TimelineStepProps) {
  const [expanded, setExpanded] = useState(false);
  const [editing, setEditing] = useState(false);
  const [editedOutput, setEditedOutput] = useState(step.output || "");
  
  const { replayMode, modifyStep, modifiedSteps } = useWorkflowStore();

  const isModified = modifiedSteps.has(step.step_id);
  const stepTypeIcon = {
    think: Brain,
    tool_call: Wrench,
    result: FileOutput,
  }[step.type] || Brain;

  const stepTypeColor = {
    think: "text-blue-500",
    tool_call: "text-green-500",
    result: "text-purple-500",
  }[step.type] || "text-gray-500";

  const StepIcon = stepTypeIcon;

  const handleSaveEdit = () => {
    modifyStep(step.step_id, editedOutput);
    setEditing(false);
  };

  return (
    <div className={cn(
      "timeline-step",
      step.type,
      isLast && "pb-0"
    )}>
      <div 
        className="flex items-start gap-3 cursor-pointer"
        onClick={() => setExpanded(!expanded)}
      >
        <div className={cn(
          "mt-0.5 p-1 rounded",
          step.type === "think" && "bg-blue-100 dark:bg-blue-900/30",
          step.type === "tool_call" && "bg-green-100 dark:bg-green-900/30",
          step.type === "result" && "bg-purple-100 dark:bg-purple-900/30"
        )}>
          <StepIcon className={cn("h-4 w-4", stepTypeColor)} />
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between">
            <span className="font-medium text-sm capitalize">
              {step.type.replace("_", " ")}
            </span>
            <div className="flex items-center gap-1">
              {step.latency_ms && (
                <span className="text-xs text-muted-foreground">
                  {step.latency_ms}ms
                </span>
              )}
              {step.tokens && (
                <span className="text-xs text-muted-foreground">
                  {step.tokens} tokens
                </span>
              )}
              {expanded ? (
                <ChevronUp className="h-4 w-4 text-muted-foreground" />
              ) : (
                <ChevronDown className="h-4 w-4 text-muted-foreground" />
              )}
            </div>
          </div>

          {/* Preview */}
          {!expanded && (
            <p className="text-xs text-muted-foreground truncate mt-1">
              {step.output || step.input || step.tool || "No content"}
            </p>
          )}

          {/* Modified indicator */}
          {isModified && (
            <span className="text-xs text-amber-500 mt-1 block">
              已修改
            </span>
          )}
        </div>
      </div>

      {/* Expanded content */}
      {expanded && (
        <div className="mt-3 pl-8 space-y-3">
          {step.input && (
            <div>
              <span className="text-xs font-medium text-muted-foreground block mb-1">
                Input
              </span>
              <pre className="text-xs bg-muted p-2 rounded overflow-x-auto whitespace-pre-wrap">
                {step.input}
              </pre>
            </div>
          )}

          {step.prompt && (
            <div>
              <span className="text-xs font-medium text-muted-foreground block mb-1">
                Prompt
              </span>
              <pre className="text-xs bg-muted p-2 rounded overflow-x-auto whitespace-pre-wrap">
                {step.prompt}
              </pre>
            </div>
          )}

          {step.tool && (
            <div>
              <span className="text-xs font-medium text-muted-foreground block mb-1">
                Tool: {step.tool}
              </span>
              {step.arguments && (
                <pre className="text-xs bg-muted p-2 rounded overflow-x-auto">
                  {JSON.stringify(step.arguments, null, 2)}
                </pre>
              )}
            </div>
          )}

          {/* Output with edit capability */}
          {(step.output || step.result) && (
            <div>
              <span className="text-xs font-medium text-muted-foreground block mb-1">
                Output
              </span>
              {editing && replayMode ? (
                <div className="space-y-2">
                  <Textarea
                    value={editedOutput}
                    onChange={(e) => setEditedOutput(e.target.value)}
                    className="min-h-[100px] text-xs"
                  />
                  <div className="flex gap-2">
                    <Button size="sm" onClick={handleSaveEdit}>
                      <Check className="h-3 w-3 mr-1" />
                      保存修改
                    </Button>
                    <Button 
                      size="sm" 
                      variant="outline" 
                      onClick={() => setEditing(false)}
                    >
                      取消
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="relative group">
                  <pre className={cn(
                    "text-xs bg-muted p-2 rounded overflow-x-auto whitespace-pre-wrap",
                    isModified && "border border-amber-500"
                  )}>
                    {isModified ? modifiedSteps.get(step.step_id) : (step.output || step.result)}
                  </pre>
                  {replayMode && (
                    <Button
                      size="sm"
                      variant="ghost"
                      className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity"
                      onClick={() => {
                        setEditedOutput(step.output || "");
                        setEditing(true);
                      }}
                    >
                      <Edit2 className="h-3 w-3" />
                    </Button>
                  )}
                </div>
              )}
            </div>
          )}

          {/* Timestamp */}
          <div className="text-xs text-muted-foreground">
            {new Date(step.timestamp).toLocaleString()}
          </div>
        </div>
      )}
    </div>
  );
}
