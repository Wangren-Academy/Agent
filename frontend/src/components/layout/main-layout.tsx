"use client";

import { ReactNode } from "react";
import { useWorkflowStore } from "@/stores/workflow-store";
import { Header } from "./header";
import { StatusBar } from "./status-bar";

interface MainLayoutProps {
  children: ReactNode;
}

export function MainLayout({ children }: MainLayoutProps) {
  const { isLoading, error } = useWorkflowStore();

  return (
    <div className="h-screen flex flex-col bg-background">
      <Header />
      
      <main className="flex-1 overflow-hidden">
        {children}
      </main>

      <StatusBar isLoading={isLoading} error={error} />
    </div>
  );
}
