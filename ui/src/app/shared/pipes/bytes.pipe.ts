// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Pipe, PipeTransform } from '@angular/core';

/** Formats a byte count to a human-readable string (e.g. "1.4 GiB"). */
@Pipe({ name: 'bytes', standalone: true })
export class BytesPipe implements PipeTransform {
  transform(value: number | null | undefined, decimals = 1): string {
    if (value == null || value === 0) return '0 B';
    const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB'];
    const k = 1024;
    const i = Math.floor(Math.log(value) / Math.log(k));
    return (value / Math.pow(k, i)).toFixed(decimals) + ' ' + units[i];
  }
}
