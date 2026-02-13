"use client";

import { useState } from "react";
import { ReactFlowProvider } from "@reactflow/core";
import { MainLayout } from "@/components/layout/main-layout";
import { WorkflowCanvas } from "@/components/workflow/workflow-canvas";
import { Sidebar } from "@/components/layout/sidebar";
import { TimelinePanel } from "@/components/timeline/timeline-panel";
import { useWorkflowStore } from "@/stores/workflow-store";

export default function Home() {
  const { selectedNodeId, executionHistory } = useWorkflowStore();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  return (
    <ReactFlowProvider>
      <MainLayout>
        <div className="flex h-full">
          {/* Left Sidebar - Agent Library */}
          <Sidebar 
            collapsed={sidebarCollapsed}
            onToggle={() => setSidebarCollapsed(!sidebarCollapsed)}
          />

          {/* Main Canvas */}
          <div className="flex-1 relative">
            <WorkflowCanvas />
          </div>

          {/* Right Panel - Timeline */}
          {(selectedNodeId || executionHistory.length > 0) && (
            <TimelinePanel />
          )}
        </div>
      </MainLayout>
    </ReactFlowProvider>
  );
}
