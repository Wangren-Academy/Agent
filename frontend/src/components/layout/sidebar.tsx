"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useWorkflowStore } from "@/stores/workflow-store";
import { ChevronLeft, ChevronRight, Bot, FileJson, Plus } from "lucide-react";
import { agentApi, workflowApi } from "@/lib/api";

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
}

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const { agents, setWorkflow, loadAgents } = useWorkflowStore();

  useEffect(() => {
    loadAgents();
  }, [loadAgents]);

  const handleDragStart = (e: React.DragEvent, agentId: string, agentName: string) => {
    e.dataTransfer.setData("application/agent", JSON.stringify({ agentId, agentName }));
    e.dataTransfer.effectAllowed = "move";
  };

  if (collapsed) {
    return (
      <div className="w-12 border-r bg-card flex flex-col items-center py-4">
        <Button variant="ghost" size="sm" onClick={onToggle}>
          <ChevronRight className="h-4 w-4" />
        </Button>
        <div className="mt-4 flex flex-col gap-2">
          <Bot className="h-5 w-5 text-muted-foreground" />
          <FileJson className="h-5 w-5 text-muted-foreground" />
        </div>
      </div>
    );
  }

  return (
    <div className="w-64 border-r bg-card flex flex-col">
      <div className="h-12 flex items-center justify-between px-4 border-b">
        <span className="font-medium">工具面板</span>
        <Button variant="ghost" size="sm" onClick={onToggle}>
          <ChevronLeft className="h-4 w-4" />
        </Button>
      </div>

      <Tabs defaultValue="agents" className="flex-1 flex flex-col">
        <TabsList className="grid w-full grid-cols-2 m-2">
          <TabsTrigger value="agents">智能体</TabsTrigger>
          <TabsTrigger value="templates">模板</TabsTrigger>
        </TabsList>

        <TabsContent value="agents" className="flex-1 mt-0">
          <ScrollArea className="h-full">
            <div className="p-2 space-y-1">
              <Button 
                variant="outline" 
                size="sm" 
                className="w-full justify-start"
                onClick={() => {/* TODO: Open create agent dialog */}}
              >
                <Plus className="h-4 w-4 mr-2" />
                创建智能体
              </Button>

              {agents.map((agent) => (
                <div
                  key={agent.id}
                  draggable
                  onDragStart={(e) => handleDragStart(e, agent.id, agent.name)}
                  className="p-3 rounded-lg border bg-background hover:bg-accent cursor-move transition-colors"
                >
                  <div className="flex items-center gap-2">
                    <Bot className="h-4 w-4 text-primary" />
                    <span className="font-medium text-sm">{agent.name}</span>
                  </div>
                  {agent.description && (
                    <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                      {agent.description}
                    </p>
                  )}
                  <div className="flex items-center gap-1 mt-2 text-xs text-muted-foreground">
                    <span className="px-1.5 py-0.5 rounded bg-muted">
                      {agent.model_config?.provider || "unknown"}
                    </span>
                    <span className="px-1.5 py-0.5 rounded bg-muted">
                      {agent.model_config?.model || "unknown"}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>
        </TabsContent>

        <TabsContent value="templates" className="flex-1 mt-0">
          <ScrollArea className="h-full">
            <div className="p-2 space-y-2">
              <TemplateCard
                name="串行处理"
                description="多个智能体依次执行"
              />
              <TemplateCard
                name="并行处理"
                description="多个智能体同时执行"
              />
              <TemplateCard
                name="路由分发"
                description="根据条件分发到不同智能体"
              />
            </div>
          </ScrollArea>
        </TabsContent>
      </Tabs>
    </div>
  );
}

function TemplateCard({ name, description }: { name: string; description: string }) {
  return (
    <div className="p-3 rounded-lg border bg-background hover:bg-accent cursor-pointer transition-colors">
      <div className="font-medium text-sm">{name}</div>
      <p className="text-xs text-muted-foreground mt-1">{description}</p>
    </div>
  );
}
