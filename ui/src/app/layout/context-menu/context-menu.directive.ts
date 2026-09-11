// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Directive, HostListener, inject, input } from '@angular/core';
import { ContextMenuItem, ContextMenuService } from './context-menu.service';

@Directive({
  selector: '[sgContextMenu]',
  standalone: true,
})
export class ContextMenuDirective {
  private menuService = inject(ContextMenuService);

  sgContextMenu = input.required<ContextMenuItem[]>();
  sgContextData = input<any>();

  @HostListener('contextmenu', ['$event'])
  onRightClick(event: MouseEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.menuService.open(
      event.clientX,
      event.clientY,
      this.sgContextMenu(),
      this.sgContextData()
    );
  }
}
