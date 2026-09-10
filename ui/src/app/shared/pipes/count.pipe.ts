// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Pipe, PipeTransform } from '@angular/core';

/** Counts items in an array where a given boolean property is truthy. */
@Pipe({ name: 'count', standalone: true })
export class CountPipe implements PipeTransform {
  transform(items: any[] | null, property: string): number {
    if (!items) return 0;
    return items.filter(item => !!item[property]).length;
  }
}
