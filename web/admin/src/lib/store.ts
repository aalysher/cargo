/** Локальное состояние админки: JWT + профиль (роль). */

export interface AdminProfile {
  id: number;
  username: string;
  display_name: string;
  role: 'operator' | 'admin';
}

export function getToken(): string {
  return localStorage.getItem('admin_token') ?? '';
}

export function getProfile(): AdminProfile | null {
  const raw = localStorage.getItem('admin_profile');
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AdminProfile;
  } catch {
    return null;
  }
}

export function saveAuth(token: string, profile: AdminProfile) {
  localStorage.setItem('admin_token', token);
  localStorage.setItem('admin_profile', JSON.stringify(profile));
}

export function clearAuth() {
  localStorage.removeItem('admin_token');
  localStorage.removeItem('admin_profile');
}
