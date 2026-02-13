// Agent types
export interface Agent {
  id: string;
  name: string;
  description?: string;
  system_prompt: string;
  model_config: ModelConfig;
  created_at: string;
  updated_at: string;
}

export interface ModelConfig {
  provider: string;
  model: string;
  temperature?: number;
  max_tokens?: number;
  top_p?: number;
}

// Workflow types
export interface Workflow {
  id: string;
  name: string;
  description?: string;
  nodes: NodeConfig[];
  edges: EdgeConfig[];
  version: number;
  created_at: string;
  updated_at: string;
}

export interface NodeConfig {
  id: string;
  agent_id: string;
  position: Position;
  data?: Record<string, unknown>;
}

export interface Position {
  x: number;
  y: number;
}

export interface EdgeConfig {
  id: string;
  source: string;
  target: string;
}

// Execution types
export interface Execution {
  id: string;
  workflow_id: string;
  status: ExecutionStatus;
  snapshot: Snapshot;
  started_at: string;
  finished_at?: string;
  created_at: string;
}

export type ExecutionStatus = "running" | "success" | "failed" | "replaying";

export interface Snapshot {
  workflow_id: string;
  execution_id: string;
  nodes: NodeSnapshot[];
  edges: EdgeConfig[];
  execution_meta: MetaInfo;
}

export interface NodeSnapshot {
  node_id: string;
  agent_name: string;
  steps: Step[];
  final_output: string;
}

export interface Step {
  step_id: string;
  type: StepType;
  input?: string;
  output?: string;
  prompt?: string;
  tokens?: number;
  latency_ms?: number;
  timestamp: string;
  tool?: string;
  arguments?: Record<string, unknown>;
  result?: string;
}

export type StepType = "think" | "tool_call" | "result";

export interface MetaInfo {
  total_tokens: number;
  total_cost: number;
  duration_ms: number;
}

// WebSocket event types
export interface WebSocketEvent {
  type: string;
  execution_id?: string;
  data?: ExecutionEventData;
  timestamp?: string;
}

export interface ExecutionEventData {
  node_id?: string;
  step?: Step;
  result?: NodeResult;
}

export interface NodeResult {
  node_id: string;
  output: string;
  steps: Step[];
  start_time: string;
  end_time: string;
  error?: string;
}

// API Response types
export interface ApiResponse<T> {
  data?: T;
  error?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
