import { inject, Injectable, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { AuthUser, OidcConfig } from './auth.model';
import { RuntimeConfigService } from '../config/runtime-config.service';

interface OidcDiscovery {
  authorization_endpoint: string;
  token_endpoint: string;
  end_session_endpoint?: string;
  userinfo_endpoint: string;
}

interface TokenResponse {
  access_token: string;
  id_token: string;
  expires_in: number;
  token_type: string;
}

interface UserInfo {
  sub: string;
  name?: string;
  email?: string;
  groups?: string[];
}

/**
 * Lightweight OIDC implementation using Authorization Code + PKCE flow.
 * Tokens are held in memory only — never in localStorage.
 * The runtime-config.json must supply oidc configuration.
 */
@Injectable({ providedIn: 'root' })
export class OidcService {
  private readonly http = inject(HttpClient);
  private readonly platformId = inject(PLATFORM_ID);
  private readonly runtimeConfig = inject(RuntimeConfigService);

  private discovery: OidcDiscovery | null = null;
  private currentUser: AuthUser | null = null;

  async initialize(): Promise<AuthUser | null> {
    if (!isPlatformBrowser(this.platformId)) return null;
    const config = this.getOidcConfig();
    if (!config) return null;

    await this.loadDiscovery(config.authority);

    // Check for callback params in URL
    const url = new URL(window.location.href);
    if (url.searchParams.has('code')) {
      return this.handleCallback();
    }

    return null;
  }

  async startLogin(): Promise<void> {
    if (!isPlatformBrowser(this.platformId)) return;
    const config = this.getOidcConfig();
    if (!config || !this.discovery) throw new Error('OIDC not configured');

    const codeVerifier = this.generateCodeVerifier();
    const codeChallenge = await this.generateCodeChallenge(codeVerifier);
    const state = this.generateState();

    sessionStorage.setItem('sg:oidc:verifier', codeVerifier);
    sessionStorage.setItem('sg:oidc:state', state);
    sessionStorage.setItem('sg:oidc:return', window.location.pathname);

    const params = new URLSearchParams({
      response_type: config.responseType,
      client_id: config.clientId,
      redirect_uri: config.redirectUri,
      scope: config.scope,
      state,
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
    });

    window.location.href = `${this.discovery.authorization_endpoint}?${params}`;
  }

  async handleCallback(): Promise<AuthUser> {
    if (!isPlatformBrowser(this.platformId)) throw new Error('SSR context');
    const config = this.getOidcConfig();
    if (!config || !this.discovery) throw new Error('OIDC not initialized');

    const url = new URL(window.location.href);
    const code = url.searchParams.get('code');
    const state = url.searchParams.get('state');

    const savedState = sessionStorage.getItem('sg:oidc:state');
    const codeVerifier = sessionStorage.getItem('sg:oidc:verifier');
    sessionStorage.removeItem('sg:oidc:state');
    sessionStorage.removeItem('sg:oidc:verifier');

    if (!code || state !== savedState || !codeVerifier) {
      throw new Error('Invalid OIDC callback parameters');
    }

    const tokenRes = await firstValueFrom(
      this.http.post<TokenResponse>(
        this.discovery.token_endpoint,
        new URLSearchParams({
          grant_type: 'authorization_code',
          client_id: config.clientId,
          redirect_uri: config.redirectUri,
          code,
          code_verifier: codeVerifier,
        }).toString(),
        { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
      )
    );

    const userInfo = await firstValueFrom(
      this.http.get<UserInfo>(this.discovery.userinfo_endpoint, {
        headers: { Authorization: `Bearer ${tokenRes.access_token}` },
      })
    );

    this.currentUser = {
      sub: userInfo.sub,
      name: userInfo.name,
      email: userInfo.email,
      groups: userInfo.groups,
      accessToken: tokenRes.access_token,
      idToken: tokenRes.id_token,
      expiresAt: Math.floor(Date.now() / 1000) + tokenRes.expires_in,
    };

    // Clean callback params from URL without reload
    const returnPath = sessionStorage.getItem('sg:oidc:return') ?? '/';
    sessionStorage.removeItem('sg:oidc:return');
    window.history.replaceState({}, '', returnPath);

    return this.currentUser;
  }

  async logout(): Promise<void> {
    if (!isPlatformBrowser(this.platformId)) return;
    const config = this.getOidcConfig();
    this.currentUser = null;
    if (config && this.discovery?.end_session_endpoint) {
      window.location.href = this.discovery.end_session_endpoint;
    }
  }

  private async loadDiscovery(authority: string): Promise<void> {
    if (this.discovery) return;
    this.discovery = await firstValueFrom(
      this.http.get<OidcDiscovery>(`${authority}/.well-known/openid-configuration`)
    );
  }

  private getOidcConfig(): OidcConfig | null {
    const raw = (this.runtimeConfig as unknown as { config?: { auth?: OidcConfig } }).config;
    return raw?.auth ?? null;
  }

  private generateCodeVerifier(): string {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return btoa(String.fromCharCode(...array))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  private async generateCodeChallenge(verifier: string): Promise<string> {
    const encoder = new TextEncoder();
    const data = encoder.encode(verifier);
    const hash = await crypto.subtle.digest('SHA-256', data);
    return btoa(String.fromCharCode(...new Uint8Array(hash)))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  private generateState(): string {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return Array.from(array, (b) => b.toString(16).padStart(2, '0')).join('');
  }
}
