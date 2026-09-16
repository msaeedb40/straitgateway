import {
  Component, input, output, computed, signal, ContentChildren, QueryList,
  TemplateRef, Directive, AfterContentInit
} from '@angular/core';
import { SkeletonListComponent } from '../../components/skeleton/skeleton.component';
import { ErrorStateComponent, EmptyStateComponent } from '../../components/error-state/error-state.component';
import { ApiError } from '../../../core/api/api-error';

export interface TableColumn<T = unknown> {
  readonly key: string;
  readonly label: string;
  readonly sortable?: boolean;
  readonly width?: string;
  readonly ariaLabel?: string;
}

export interface SortState {
  readonly key: string;
  readonly direction: 'asc' | 'desc';
}

@Component({
  selector: 'sg-data-table',
  imports: [SkeletonListComponent, ErrorStateComponent, EmptyStateComponent],
  template: `
    @if (loading()) {
      <sg-skeleton-list [count]="8" label="Loading table data…" />
    } @else if (error()) {
      <sg-error-state [error]="error()" (retry)="retry.emit()" />
    } @else if (rows().length === 0) {
      <sg-empty-state [title]="emptyTitle()" [message]="emptyMessage()" />
    } @else {
      <div class="sg-table-wrapper" role="region" [attr.aria-label]="caption()">
        <table class="sg-table" [attr.aria-rowcount]="totalRows()">
          <caption class="sg-sr-only">{{ caption() }}</caption>
          <thead>
            <tr>
              @for (col of columns(); track col.key) {
                <th
                  scope="col"
                  [style.width]="col.width"
                  [attr.aria-sort]="ariaSort(col)"
                  [class.sg-sortable]="col.sortable"
                  (click)="col.sortable && onSort(col.key)"
                  (keydown.enter)="col.sortable && onSort(col.key)"
                  [attr.tabindex]="col.sortable ? 0 : null"
                  [attr.role]="col.sortable ? 'button' : null"
                >
                  {{ col.label }}
                  @if (col.sortable) {
                    <span class="sg-sort-icon" aria-hidden="true">
                      {{ sortState()?.key === col.key ? (sortState()!.direction === 'asc' ? '↑' : '↓') : '↕' }}
                    </span>
                  }
                </th>
              }
            </tr>
          </thead>
          <tbody>
            @for (row of rows(); track trackBy()(row, $index); let i = $index) {
              <tr
                class="sg-table-row"
                [class.sg-row-clickable]="rowClickable()"
                (click)="rowClickable() && rowClick.emit(row)"
                (keydown.enter)="rowClickable() && rowClick.emit(row)"
                [attr.tabindex]="rowClickable() ? 0 : null"
                [attr.aria-rowindex]="i + 1"
              >
                <ng-content />
              </tr>
            }
          </tbody>
        </table>

        @if (totalRows() > rows().length) {
          <div class="sg-table-pagination" role="navigation" aria-label="Table pagination">
            <span class="sg-pagination-info">
              {{ rows().length }} of {{ totalRows() }} rows
            </span>
            <button
              class="sg-btn sg-btn-ghost"
              [disabled]="page() <= 1"
              (click)="pageChange.emit(page() - 1)"
              aria-label="Previous page"
            >‹ Prev</button>
            <span aria-live="polite" aria-atomic="true">Page {{ page() }}</span>
            <button
              class="sg-btn sg-btn-ghost"
              [disabled]="page() * pageSize() >= totalRows()"
              (click)="pageChange.emit(page() + 1)"
              aria-label="Next page"
            >Next ›</button>
          </div>
        }
      </div>
    }
  `,
  styles: [`
    .sg-table-wrapper { overflow-x: auto; }
    .sg-sortable { cursor: pointer; user-select: none; }
    .sg-sortable:hover { color: var(--sg-accent); }
    .sg-sort-icon { margin-left: 4px; opacity: 0.6; font-size: 0.7em; }
    .sg-row-clickable { cursor: pointer; }
    .sg-table-pagination {
      display: flex; align-items: center; gap: 12px;
      padding: 10px 12px; border-top: 1px solid var(--sg-border);
      font-size: 0.8125rem; color: var(--sg-text-secondary);
    }
    .sg-pagination-info { flex: 1; }
  `],
})
export class DataTableComponent<T extends object> {
  readonly columns     = input.required<TableColumn<T>[]>();
  readonly rows        = input.required<T[]>();
  readonly loading     = input<boolean>(false);
  readonly error       = input<ApiError | null>(null);
  readonly caption     = input<string>('Data table');
  readonly emptyTitle  = input<string>('No results');
  readonly emptyMessage = input<string | undefined>(undefined);
  readonly totalRows   = input<number>(0);
  readonly page        = input<number>(1);
  readonly pageSize    = input<number>(20);
  readonly rowClickable = input<boolean>(false);
  readonly trackBy     = input<(row: T, index: number) => unknown>((row, i) => i);

  readonly sortState   = signal<SortState | null>(null);
  readonly retry       = output<void>();
  readonly rowClick    = output<T>();
  readonly sortChange  = output<SortState>();
  readonly pageChange  = output<number>();

  onSort(key: string): void {
    const current = this.sortState();
    const next: SortState = current?.key === key
      ? { key, direction: current.direction === 'asc' ? 'desc' : 'asc' }
      : { key, direction: 'asc' };
    this.sortState.set(next);
    this.sortChange.emit(next);
  }

  ariaSort(col: TableColumn<T>): string | null {
    if (!col.sortable) return null;
    const s = this.sortState();
    if (!s || s.key !== col.key) return 'none';
    return s.direction === 'asc' ? 'ascending' : 'descending';
  }
}
