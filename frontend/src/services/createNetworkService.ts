import { NetworkService } from './NetworkService';
import { defaultNetworkService } from './MockNetworkService';
import { EngineConnection, HttpNetworkService } from './HttpNetworkService';

interface WailsWindow {
  go?: { main?: { App?: { GetConnection?: () => Promise<EngineConnection> } } };
}

const sleep = (ms: number) => new Promise(r => setTimeout(r, ms));

/** Wails host (production): the Go side hands us the loopback URL + session token. */
async function fromWails(): Promise<EngineConnection | null> {
  const get = (window as unknown as WailsWindow).go?.main?.App?.GetConnection;
  if (!get) return null;
  for (let i = 0; i < 50; i++) { // the engine starts asynchronously
    const c = await get();
    if (c?.baseUrl && c?.token) return c;
    await sleep(100);
  }
  throw new Error('NetWatch engine did not start.');
}

/** Browser development against `netwatchd`: ?api=http://127.0.0.1:PORT&token=... (remembered per tab). */
function fromQuery(): EngineConnection | null {
  try {
    const q = new URLSearchParams(window.location.search);
    const api = q.get('api');
    const token = q.get('token');
    if (api && token) {
      sessionStorage.setItem('netwatch.conn', JSON.stringify({ baseUrl: api, token }));
      window.history.replaceState(null, '', window.location.pathname); // keep the token out of the URL bar
      return { baseUrl: api, token };
    }
    const saved = sessionStorage.getItem('netwatch.conn');
    return saved ? (JSON.parse(saved) as EngineConnection) : null;
  } catch {
    return null;
  }
}

export interface ServiceHandle {
  service: NetworkService;
  /** Set when a real engine was expected but could not be reached. */
  startupError?: string;
}

/** Real engine when a connection is available, otherwise the prototype mock. */
export async function createNetworkService(): Promise<ServiceHandle> {
  try {
    const conn = (await fromWails()) ?? fromQuery();
    if (conn) {
      const svc = new HttpNetworkService(conn);
      await svc.getNetworkInfo().catch(() => undefined); // warm-up; errors surface in the UI
      return { service: svc };
    }
  } catch (e) {
    return { service: defaultNetworkService, startupError: e instanceof Error ? e.message : String(e) };
  }
  return { service: defaultNetworkService };
}
