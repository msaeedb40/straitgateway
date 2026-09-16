export type CaptureProtocol = 'TCP' | 'UDP' | 'ICMP' | 'ARP' | 'ANY';
export type CaptureState = 'Pending' | 'Running' | 'Stopped' | 'Failed' | 'Completed';

export interface CaptureSpec {
  readonly nodeName: string;
  readonly namespace?: string;
  readonly podName?: string;
  readonly interfaceName?: string;
  readonly protocol: CaptureProtocol;
  readonly sourceFilter?: string;
  readonly destinationFilter?: string;
  readonly port?: number;
  readonly durationSeconds?: number;
  readonly packetLimit?: number;
}

export interface CaptureSession {
  readonly id: string;
  readonly spec: CaptureSpec;
  readonly state: CaptureState;
  readonly startTime?: string;
  readonly endTime?: string;
  readonly packetCount: number;
  readonly error?: string;
}

export type PacketDirection = 'Inbound' | 'Outbound';

export interface CapturedPacket {
  readonly index: number;
  readonly timestamp: string;
  readonly sourceIP: string;
  readonly destinationIP: string;
  readonly sourcePort: number | null;
  readonly destinationPort: number | null;
  readonly protocol: string;
  readonly length: number;
  readonly direction: PacketDirection;
  readonly flags?: string[];
  readonly payload?: string;
}
