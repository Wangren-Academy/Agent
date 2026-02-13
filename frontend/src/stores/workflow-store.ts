import { create } from "zustand";
import type { Node, Edge } from "@reactflow/core";
import type { Agent, Workflow, Execution, Step, NodeSnapshot } from "@/lib/types";
import { workflowApi, executionApi } from "@/lib/api";
import { agentApi } from "@/lib/api";

interface WorkflowState {
  // Current workflow
  workflow: Workflow | null;
  nodes: Node[];
  edges: Edge[];
  selectedNodeId: string | null;
  isDirty: boolean;

  // Agents
  agents: Agent[];

  // Execution
  currentExecution: Execution | null;
  executionHistory: Execution[];
  executionSteps: Map<string, Step[]>;
  replayMode: boolean;
  modifiedSteps: Map<string, string>;

  // UI State
  isLoading: boolean;
  error: string | null;

  // Actions - Workflow
  setWorkflow: (workflow: Workflow | null) => void;
  addNode: (agent: Agent, position: { x: number; y: number }) => void;
  updateNode: (nodeId: string, data: Partial<Node["data"]>) => void;
  removeNode: (nodeId: string) => void;
  addEdge: (source: string, target: string) => void;
  removeEdge: (edgeId: string) => void;
  setSelectedNode: (nodeId: string | null) => void;
  saveWorkflow: () => Promise<void>;
  loadWorkflow: (id: string) => Promise<void>;
  newWorkflow: () => void;

  // Actions - Agents
  loadAgents: () => Promise<void>;

  // Actions - Execution
  executeWorkflow: (inputData?: Record<string, unknown>) => Promise<void>;
  loadExecution: (executionId: string) => Promise<void>;
  loadExecutionHistory: (workflowId: string) => Promise<void>;
  addExecutionStep: (nodeId: string, step: Step) => void;
  toggleReplayMode: () => void;
  modifyStep: (stepId: string, newOutput: string) => void;
  replayExecution: () => Promise<void>;

  // Actions - General
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  reset: () => void;
}

const initialState = {
  workflow: null,
  nodes: [],
  edges: [],
  selectedNodeId: null,
  isDirty: false,
  agents: [],
  currentExecution: null,
  executionHistory: [],
  executionSteps: new Map(),
  replayMode: false,
  modifiedSteps: new Map(),
  isLoading: false,
  error: null,
};

export const useWorkflowStore = create<WorkflowState>((set, get) => ({
  ...initialState,

  // Workflow Actions
  setWorkflow: (workflow) => {
    if (workflow) {
      const nodes: Node[] = workflow.nodes.map((n) => ({
        id: n.id,
        type: "agent",
        position: n.position,
        data: { agent_id: n.agent_id, ...n.data },
      }));
      const edges: Edge[] = workflow.edges.map((e) => ({
        id: e.id,
        source: e.source,
        target: e.target,
      }));
      set({ workflow, nodes, edges, isDirty: false });
    } else {
      set({ workflow: null, nodes: [], edges: [], isDirty: false });
    }
  },

  addNode: (agent, position) => {
    const nodeId = `node_${Date.now()}`;
    const newNode: Node = {
      id: nodeId,
      type: "agent",
      position,
      data: { agent_id: agent.id, agent_name: agent.name },
    };
    set((state) => ({
      nodes: [...state.nodes, newNode],
      isDirty: true,
    }));
  },

  updateNode: (nodeId, data) => {
    set((state) => ({
      nodes: state.nodes.map((n) =>
        n.id === nodeId ? { ...n, data: { ...n.data, ...data } } : n
      ),
      isDirty: true,
    }));
  },

  removeNode: (nodeId) => {
    set((state) => ({
      nodes: state.nodes.filter((n) => n.id !== nodeId),
      edges: state.edges.filter((e) => e.source !== nodeId && e.target !== nodeId),
      selectedNodeId: state.selectedNodeId === nodeId ? null : state.selectedNodeId,
      isDirty: true,
    }));
  },

  addEdge: (source, target) => {
    const edgeId = `edge_${source}_${target}`;
    const newEdge: Edge = {
      id: edgeId,
      source,
      target,
    };
    set((state) => ({
      edges: [...state.edges, newEdge],
      isDirty: true,
    }));
  },

  removeEdge: (edgeId) => {
    set((state) => ({
      edges: state.edges.filter((e) => e.id !== edgeId),
      isDirty: true,
    }));
  },

  setSelectedNode: (nodeId) => set({ selectedNodeId: nodeId }),

  saveWorkflow: async () => {
    const { workflow, nodes, edges } = get();
    if (!workflow) return;

    const nodeConfigs = nodes.map((n) => ({
      id: n.id,
      agent_id: n.data.agent_id as string,
      position: n.position,
      data: n.data,
    }));
    const edgeConfigs = edges.map((e) => ({
      id: e.id,
      source: e.source,
      target: e.target,
    }));

    const result = await workflowApi.update(workflow.id, {
      nodes: nodeConfigs,
      edges: edgeConfigs,
    });

    if (result.data) {
      set({ isDirty: false });
    } else if (result.error) {
      set({ error: result.error });
    }
  },

  loadWorkflow: async (id) => {
    set({ isLoading: true, error: null });
    const result = await workflowApi.get(id);
    if (result.data) {
      get().setWorkflow(result.data);
    } else if (result.error) {
      set({ error: result.error });
    }
    set({ isLoading: false });
  },

  newWorkflow: () => {
    set({
      workflow: null,
      nodes: [],
      edges: [],
      selectedNodeId: null,
      currentExecution: null,
      executionSteps: new Map(),
      replayMode: false,
      modifiedSteps: new Map(),
      isDirty: false,
    });
  },

  // Agent Actions
  loadAgents: async () => {
    const result = await agentApi.list();
    if (result.data) {
      set({ agents: result.data });
    } else if (result.error) {
      set({ error: result.error });
    }
  },

  // Execution Actions
  executeWorkflow: async (inputData) => {
    const { workflow } = get();
    if (!workflow) return;

    set({ isLoading: true, error: null, executionSteps: new Map() });
    
    const result = await workflowApi.execute(workflow.id, inputData);
    if (result.data) {
      // Execution started, WebSocket will handle updates
      set({ currentExecution: { id: result.data.execution_id, workflow_id: workflow.id, status: "running" } as Execution });
    } else if (result.error) {
      set({ error: result.error });
    }
    set({ isLoading: false });
  },

  loadExecution: async (executionId) => {
    set({ isLoading: true, error: null });
    const result = await executionApi.get(executionId);
    if (result.data) {
      // Rebuild execution steps from snapshot
      const steps = new Map<string, Step[]>();
      result.data.snapshot.nodes.forEach((node: NodeSnapshot) => {
        steps.set(node.node_id, node.steps);
      });
      set({
        currentExecution: result.data,
        executionSteps: steps,
      });
    } else if (result.error) {
      set({ error: result.error });
    }
    set({ isLoading: false });
  },

  loadExecutionHistory: async (workflowId) => {
    const result = await executionApi.list(workflowId);
    if (result.data) {
      set({ executionHistory: result.data });
    } else if (result.error) {
      set({ error: result.error });
    }
  },

  addExecutionStep: (nodeId, step) => {
    set((state) => {
      const newSteps = new Map(state.executionSteps);
      const nodeSteps = newSteps.get(nodeId) || [];
      newSteps.set(nodeId, [...nodeSteps, step]);
      return { executionSteps: newSteps };
    });
  },

  toggleReplayMode: () => set((state) => ({ replayMode: !state.replayMode })),

  modifyStep: (stepId, newOutput) => {
    set((state) => {
      const newModified = new Map(state.modifiedSteps);
      newModified.set(stepId, newOutput);
      return { modifiedSteps: newModified };
    });
  },

  replayExecution: async () => {
    const { currentExecution, modifiedSteps } = get();
    if (!currentExecution) return;

    const modifications = Array.from(modifiedSteps.entries()).map(([stepId, newOutput]) => ({
      step_id: stepId,
      new_output: newOutput,
    }));

    const result = await executionApi.replay(currentExecution.id, modifications);
    if (result.data) {
      set({
        currentExecution: {
          ...currentExecution,
          id: result.data.new_execution_id,
          status: "replaying",
        },
        modifiedSteps: new Map(),
      });
    } else if (result.error) {
      set({ error: result.error });
    }
  },

  // General Actions
  setLoading: (loading) => set({ isLoading: loading }),
  setError: (error) => set({ error }),
  reset: () => set(initialState),
}));
