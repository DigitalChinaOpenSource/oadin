import { AUTH_TOKEN } from '../constants';

export function baseHeaders(providedToken = null) {
  const token = providedToken || window.localStorage.getItem(AUTH_TOKEN);
  return {
    Authorization: token ? `Bearer ${token}` : null,
  } as any;
}
