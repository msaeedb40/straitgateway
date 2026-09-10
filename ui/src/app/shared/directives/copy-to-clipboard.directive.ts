// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Directive, HostListener, input, inject } from '@angular/core';

/** Click-to-copy directive. Copies the input value to the clipboard. */
@Directive({ selector: '[sgCopy]', standalone: true })
export class CopyToClipboardDirective {
  sgCopy = input.required<string>();

  @HostListener('click')
  async onClick() {
    try {
      await navigator.clipboard.writeText(this.sgCopy());
    } catch {
      // Fallback for older browsers.
      const ta = document.createElement('textarea');
      ta.value = this.sgCopy();
      ta.style.position = 'fixed';
      ta.style.left = '-9999px';
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
    }
  }
}
