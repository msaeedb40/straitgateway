import { Component, inject, computed } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ActivatedRoute, Router, NavigationEnd } from '@angular/router';
import { filter, map } from 'rxjs/operators';
import { toSignal } from '@angular/core/rxjs-interop';

interface Crumb {
  readonly label: string;
  readonly path: string | null;
}

@Component({
  selector: 'sg-breadcrumbs',
  imports: [RouterLink],
  template: `
    @if (crumbs().length > 1) {
      <nav class="sg-breadcrumbs" aria-label="Breadcrumb">
        <ol class="sg-crumb-list" role="list">
          @for (crumb of crumbs(); track crumb.path; let last = $last) {
            <li class="sg-crumb-item" role="listitem">
              @if (crumb.path && !last) {
                <a [routerLink]="crumb.path" class="sg-crumb-link">{{ crumb.label }}</a>
              } @else {
                <span class="sg-crumb-current" [attr.aria-current]="last ? 'page' : null">
                  {{ crumb.label }}
                </span>
              }
              @if (!last) {
                <span class="sg-crumb-sep" aria-hidden="true">/</span>
              }
            </li>
          }
        </ol>
      </nav>
    }
  `,
  styles: [`
    .sg-breadcrumbs { padding: 0 1.5rem; }
    .sg-crumb-list {
      display: flex;
      align-items: center;
      gap: 6px;
      list-style: none;
      margin: 0;
      padding: 8px 0;
      font-size: 0.75rem;
    }
    .sg-crumb-item { display: flex; align-items: center; gap: 6px; }
    .sg-crumb-link {
      color: var(--sg-text-secondary);
      text-decoration: none;
      transition: color var(--sg-transition-fast);
    }
    .sg-crumb-link:hover { color: var(--sg-accent); }
    .sg-crumb-current { color: var(--sg-text-primary); font-weight: 500; }
    .sg-crumb-sep { color: var(--sg-text-muted); }
  `],
})
export class BreadcrumbsComponent {
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);

  private readonly url$ = toSignal(
    this.router.events.pipe(
      filter((e): e is NavigationEnd => e instanceof NavigationEnd),
      map((e) => e.urlAfterRedirects)
    ),
    { initialValue: this.router.url }
  );

  readonly crumbs = computed<Crumb[]>(() => {
    const url = this.url$() ?? '/';
    const segments = url.split('/').filter(Boolean);
    const crumbs: Crumb[] = [{ label: 'StraitGateway', path: '/dashboard' }];
    let path = '';
    for (const seg of segments) {
      path += `/${seg}`;
      crumbs.push({
        label: this.formatLabel(seg),
        path,
      });
    }
    return crumbs;
  });

  private formatLabel(segment: string): string {
    // UUID-like segments shown as IDs, others title-cased
    if (/^[0-9a-f-]{32,}$/i.test(segment)) return segment.slice(0, 8) + '…';
    return segment.charAt(0).toUpperCase() + segment.slice(1).replace(/-/g, ' ');
  }
}
