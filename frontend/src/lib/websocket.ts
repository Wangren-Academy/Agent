import type { WebSocketEvent, Step } from "./types";

const WS_BASE = process.env.NEXT_PUBLIC_API_URL?.replace("http", "ws") || "ws://localhost:8080";

export type StepUpdateHandler = (step: Step, nodeId: string) => void;
export type ConnectionHandler = (connected: boolean) => void;

export class ExecutionWebSocket {
  private ws: WebSocket | null = null;
  private executionId: string;
  private onStepUpdate: StepUpdateHandler;
  private onConnectionChange?: ConnectionHandler;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  constructor(
    executionId: string,
    onStepUpdate: StepUpdateHandler,
    onConnectionChange?: ConnectionHandler
  ) {
    this.executionId = executionId;
    this.onStepUpdate = onStepUpdate;
    this.onConnectionChange = onConnectionChange;
    this.connect();
  }

  private connect() {
    try {
      this.ws = new WebSocket(`${WS_BASE}/ws/executions/${this.executionId}`);

      this.ws.onopen = () => {
        console.log(`[WebSocket] Connected to execution ${this.executionId}`);
        this.reconnectAttempts = 0;
        this.onConnectionChange?.(true);
      };

      this.ws.onmessage = (event) => {
        try {
          const payload: WebSocketEvent = JSON.parse(event.data);
          this.handleMessage(payload);
        } catch (error) {
          console.error("[WebSocket] Failed to parse message:", error);
        }
      };

      this.ws.onclose = (event) => {
        console.log("[WebSocket] Connection closed:", event.code);
        this.onConnectionChange?.(false);
        this.attemptReconnect();
      };

      this.ws.onerror = (error) => {
        console.error("[WebSocket] Error:", error);
      };
    } catch (error) {
      console.error("[WebSocket] Failed to connect:", error);
      this.attemptReconnect();
    }
  }

  private handleMessage(event: WebSocketEvent) {
    switch (event.type) {
      case "step_update":
      case "step_complete":
        if (event.data?.step && event.data?.node_id) {
          this.onStepUpdate(event.data.step, event.data.node_id);
        }
        break;

      case "node_complete":
      case "node_failed":
        console.log(`[WebSocket] Node event: ${event.type}`, event.data);
        break;

      case "execution_complete":
        console.log("[WebSocket] Execution complete");
        break;

      default:
        console.log("[WebSocket] Unknown event type:", event.type);
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
      setTimeout(() => this.connect(), delay);
    } else {
      console.error("[WebSocket] Max reconnect attempts reached");
    }
  }

  sendModification(stepId: string, newOutput: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(
        JSON.stringify({
          type: "modify_step",
          data: {
            step_id: stepId,
            new_output: newOutput,
          },
        })
      );
    } else {
      console.error("[WebSocket] Cannot send - connection not open");
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Hook for using WebSocket in components
import { useEffect, useState, useCallback } from "react";

export function useExecutionWebSocket(executionId: string | null) {
  const [socket, setSocket] = useState<ExecutionWebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [steps, setSteps] = useState<Map<string, Step[]>>(new Map());

  const handleStepUpdate = useCallback((step: Step, nodeId: string) => {
    setSteps((prev) => {
      const newMap = new Map(prev);
      const nodeSteps = newMap.get(nodeId) || [];
      newMap.set(nodeId, [...nodeSteps, step]);
      return newMap;
    });
  }, []);

  useEffect(() => {
    if (!executionId) {
      socket?.disconnect();
      setSocket(null);
      setConnected(false);
      return;
    }

    const ws = new ExecutionWebSocket(
      executionId,
      handleStepUpdate,
      setConnected
    );
    setSocket(ws);

    return () => {
      ws.disconnect();
    };
  }, [executionId, handleStepUpdate]);

  return {
    socket,
    connected,
    steps,
    clearSteps: () => setSteps(new Map()),
  };
}
