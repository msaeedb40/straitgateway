import { StraitGatewayApiClient } from './client'
import { DataplaneMetrics, EbpfProgramStatus, FlowLogEntry } from '../types/observability'

export class ObservabilityApi {
  constructor(private client: StraitGatewayApiClient) {}

  public getMetrics(signal?: AbortSignal): Promise<DataplaneMetrics> {
    return this.client.get<DataplaneMetrics>('/observability/metrics', undefined, signal)
  }

  public getFlowLogs(limit = 100, signal?: AbortSignal): Promise<FlowLogEntry[]> {
    return this.client.get<FlowLogEntry[]>('/observability/flows', { limit }, signal)
  }

  public getEbpfPrograms(signal?: AbortSignal): Promise<EbpfProgramStatus[]> {
    return this.client.get<EbpfProgramStatus[]>('/observability/ebpf/programs', undefined, signal)
  }
}
