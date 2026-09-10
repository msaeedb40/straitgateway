// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ShellComponent } from './layout/shell/shell.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, ShellComponent],
  template: `<sg-shell></sg-shell>`,
})
export class AppComponent {}

// Export alias for main.ts compatibility
export { AppComponent as App };
