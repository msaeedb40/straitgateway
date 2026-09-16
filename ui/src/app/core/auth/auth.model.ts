export type OidcState = 'Unauthenticated' | 'Authenticating' | 'Authenticated' | 'Error';

export interface OidcConfig {
  readonly authority: string;
  readonly clientId: string;
  readonly redirectUri: string;
  readonly scope: string;
  readonly responseType: string;
}

export interface AuthUser {
  readonly sub: string;
  readonly name?: string;
  readonly email?: string;
  readonly groups?: string[];
  readonly accessToken: string;
  readonly idToken: string;
  readonly expiresAt: number; // Unix epoch seconds
}

export interface AuthState {
  readonly status: OidcState;
  readonly user: AuthUser | null;
  readonly error: string | null;
}
