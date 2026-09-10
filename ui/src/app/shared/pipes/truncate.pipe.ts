// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Pipe, PipeTransform } from '@angular/core';

/** Truncates a string to a given length with ellipsis. */
@Pipe({ name: 'truncate', standalone: true })
export class TruncatePipe implements PipeTransform {
  transform(value: string | null, maxLength = 50): string {
    if (!value) return '';
    return value.length > maxLength ? value.substring(0, maxLength) + '…' : value;
  }
}
