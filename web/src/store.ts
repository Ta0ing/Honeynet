import { create } from 'zustand';

export type AuthUser = { id: string; username: string; role: 'admin' | 'operator' | 'viewer'; display_name?: string };

type AuthState = {
  token: string | null;
  user: AuthUser | null;
  login: (token: string, user: AuthUser) => void;
  logout: () => void;
};

const savedUser = localStorage.getItem('honeynet_user');
export const useAuth = create<AuthState>((set) => ({
  token: localStorage.getItem('honeynet_token'),
  user: savedUser ? JSON.parse(savedUser) : null,
  login: (token, user) => {
    localStorage.setItem('honeynet_token', token);
    localStorage.setItem('honeynet_user', JSON.stringify(user));
    set({ token, user });
  },
  logout: () => {
    localStorage.removeItem('honeynet_token');
    localStorage.removeItem('honeynet_user');
    set({ token: null, user: null });
  },
}));

export type ThemeMode = 'light' | 'dark';

type ThemeState = {
  mode: ThemeMode;
  setMode: (mode: ThemeMode) => void;
  toggle: () => void;
};

function savedTheme(): ThemeMode {
  return localStorage.getItem('honeynet_theme') === 'dark' ? 'dark' : 'light';
}

export function applyTheme(mode: ThemeMode) {
  document.documentElement.dataset.theme = mode;
  document.documentElement.style.colorScheme = mode;
  if (mode === 'dark') document.body.setAttribute('arco-theme', 'dark');
  else document.body.removeAttribute('arco-theme');
}

export const useTheme = create<ThemeState>((set, get) => ({
  mode: savedTheme(),
  setMode: (mode) => {
    localStorage.setItem('honeynet_theme', mode);
    applyTheme(mode);
    set({ mode });
  },
  toggle: () => get().setMode(get().mode === 'light' ? 'dark' : 'light'),
}));
