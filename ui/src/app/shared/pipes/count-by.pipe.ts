import { Pipe, PipeTransform } from '@angular/core';
@Pipe({ name: 'countBy', standalone: true })
export class CountByPipe implements PipeTransform {
  transform(items: any[] | null, field: string, value: string): number {
    if (!items) return 0;
    return items.filter(i => i[field] === value).length;
  }
}
