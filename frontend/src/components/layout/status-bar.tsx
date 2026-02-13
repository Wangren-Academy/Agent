"use client";

import { AlertCircle, CheckCircle, Loader2 } from "lucide-react";
import { useWorkflowStore } from "@/stores/workflow-store";

interface StatusBarProps {
  isLoading: boolean;
  error: string | null;
}

export function StatusBar({ isLoading, error }: StatusBarProps) {
  const { currentExecution, workflow } = useWorkflowStore();

  return (
    <footer className="h-8 border-t bg-card flex items-center justify-between px-4 text-xs">
      <div className="flex items-center gap-4">
        {isLoading && (
          <div className="flex items-center gap-1 text-muted-foreground">
            <Loader2 className="h-3 w-3 animate-spin" />
            <span>处理中...</span>
          </div>
        )}

        {error && (
          <div className="flex items-center gap-1 text-destructive">
            <AlertCircle className="h-3 w-3" />
            <span>{error}</span>
          </div>
        )}

        {!isLoading && !error && (
          <div className="flex items-center gap-1 text-muted-foreground">
            <CheckCircle className="h-3 w-3 text-green-500" />
            <span>就绪</span>
          </div>
        )}
      </div>

      <div className="flex items-center gap-4 text-muted-foreground">
        {currentExecution && (
          <>
            <span>状态: {currentExecution.status}</span>
            {currentExecution.snapshot?.execution_meta && (
              <>
                <span>
                  Tokens: {currentExecution.snapshot.execution_meta.total_tokens}
                </span>
                <span>
                  耗时: {currentExecution.snapshot.execution_meta.duration_ms}ms
                </span>
              </>
            )}
          </>
        )}
        {workflow && (
          <span>版本: v{workflow.version}</span>
        )}
      </div>
    </footer>
  );
}
