import { describe, expect, it } from 'vitest';
import { canAccess, requiredRole } from './authorization';
import { defaultRoute } from './session';

describe('role-based route boundaries', () => {
  it('keeps contributor routes exclusive to contributors', () => {
    expect(requiredRole('/dashboard')).toBe('contributor');
    expect(requiredRole('/contributor/uploads')).toBe('contributor');
    expect(canAccess('contributor', '/dashboard')).toBe(true);
    expect(canAccess('buyer', '/dashboard')).toBe(false);
    expect(canAccess('buyer', '/contributor/uploads')).toBe(false);
  });

  it('keeps buyer routes exclusive to buyers', () => {
    expect(requiredRole('/buyer')).toBe('buyer');
    expect(requiredRole('/buyer/collections')).toBe('buyer');
    expect(canAccess('buyer', '/buyer')).toBe(true);
    expect(canAccess('contributor', '/buyer')).toBe(false);
    expect(canAccess('contributor', '/buyer/collections')).toBe(false);
  });

  it('allows both roles onto public routes', () => {
    expect(requiredRole('/search')).toBeUndefined();
    expect(canAccess('buyer', '/search')).toBe(true);
    expect(canAccess('contributor', '/search')).toBe(true);
  });

  it('sends each role to its own workspace after login', () => {
    expect(defaultRoute('buyer')).toBe('/buyer');
    expect(defaultRoute('contributor')).toBe('/dashboard');
  });
});
