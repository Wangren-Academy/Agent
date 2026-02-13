"use client";

import { Button } from "@/components/ui/button";
import { 
  DropdownMenu, 
  DropdownMenuContent, 
  DropdownMenuItem, 
  DropdownMenuTrigger 
} from "@/components/ui/dropdown-menu";
import { useWorkflowStore } from "@/stores/workflow-store";
import { 
  Play, 
  Save, 
  FolderOpen, 
  Plus, 
  ChevronDown,
  Settings,
  User
} from "lucide-react";

export function Header() {
  const { workflow, saveWorkflow, executeWorkflow, newWorkflow, isDirty } = useWorkflowStore();

  return (
    <header className="h-14 border-b bg-card flex items-center justify-between px-4">
      <div className="flex items-center gap-4">
        <h1 className="text-xl font-bold text-primary">AgentForge</h1>
        
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="sm" onClick={newWorkflow}>
            <Plus className="h-4 w-4 mr-1" />
            新建
          </Button>
          
          <Button variant="ghost" size="sm" onClick={() => {/* TODO: Open dialog */}}>
            <FolderOpen className="h-4 w-4 mr-1" />
            打开
          </Button>
          
          <Button 
            variant="ghost" 
            size="sm" 
            onClick={saveWorkflow}
            disabled={!workflow || !isDirty}
          >
            <Save className="h-4 w-4 mr-1" />
            保存
          </Button>
        </div>
      </div>

      <div className="flex items-center gap-2">
        {workflow && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">
              {workflow.name}
              {isDirty && " *"}
            </span>
            
            <Button 
              size="sm"
              onClick={() => executeWorkflow()}
              className="bg-primary hover:bg-primary/90"
            >
              <Play className="h-4 w-4 mr-1" />
              执行
            </Button>
          </div>
        )}

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm">
              <User className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem>
              <Settings className="h-4 w-4 mr-2" />
              设置
            </DropdownMenuItem>
            <DropdownMenuItem>
              API 密钥配置
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
