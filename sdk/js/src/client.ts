import fetch, { Headers, type RequestInit } from "node-fetch";

import { StremThruError } from "./error";
import { StoreMagnetStatus, StoreUserSubscriptionStatus } from "./types";
import { VERSION } from "./version";

const USER_AGENT = `stremthru:sdk:js/${VERSION}`;

export type StremThruConfig = {
  auth?:
    | string
    | { pass: string; user: string }
    | { store: string; token: string };
} & {
  baseUrl: string;
  clientIp?: string;
  timeout?: number;
  userAgent?: string;
};

type ResponseMeta = {
  headers: Record<string, string>;
  statusCode: number;
  statusText: string;
};

class StremThruStore {
  #client: StremThru;
  #clientIp?: string;

  constructor(client: StremThru, clientIp?: string) {
    this.#client = client;
    if (clientIp) {
      this.#clientIp = clientIp;
    }
  }

  async addMagnet({
    clientIp = this.#clientIp,
    magnet,
  }: {
    clientIp?: string;
    magnet: string;
  }) {
    return await this.#client.request<{
      added_at: string;
      files: Array<{
        index: number;
        link: string;
        name: string;
        path: string;
        size: number;
        video_hash: string;
      }>;
      hash: string;
      id: string;
      magnet: string;
      name: string;
      status: StoreMagnetStatus;
    }>("/v0/store/magnets", {
      body: { magnet },
      method: "POST",
      params: clientIp ? { client_ip: clientIp } : {},
    });
  }

  async checkMagnet(params: { magnet: string[]; sid?: string }) {
    return await this.#client.request<{
      items: Array<{
        files: Array<{
          index: number;
          name: string;
          size: number;
        }>;
        hash: string;
        magnet: string;
        status: StoreMagnetStatus;
      }>;
    }>("/v0/store/magnets/check", {
      method: "GET",
      params,
    });
  }

  async generateLink({
    clientIp = this.#clientIp,
    link,
  }: {
    clientIp?: string;
    link: string;
  }) {
    return await this.#client.request<{
      link: string;
    }>(`/v0/store/link/generate`, {
      body: { link },
      method: "POST",
      params: clientIp ? { client_ip: clientIp } : {},
    });
  }

  async getMagnet(magnetId: string) {
    return await this.#client.request<{
      added_at: string;
      files: Array<{
        index: number;
        link: string;
        name: string;
        path: string;
        size: number;
        video_hash: string;
      }>;
      hash: string;
      id: string;
      name: string;
      status: StoreMagnetStatus;
    }>(`/v0/store/magnets/${magnetId}`, { method: "GET" });
  }

  async getUser() {
    return await this.#client.request<{
      email: string;
      id: string;
      subscription_status: StoreUserSubscriptionStatus;
    }>("/v0/store/user", { method: "GET" });
  }

  async listMagnets({
    limit,
    offset,
  }: {
    // min `1`, max `500`, default `100`
    limit?: number;
    // min `0`, default `0`
    offset?: number;
  }) {
    const params: Record<string, string> = {};
    if (limit) {
      params["limit"] = String(limit);
    }
    if (offset) {
      params["offset"] = String(offset);
    }
    return await this.#client.request<{
      items: Array<{
        added_at: string;
        hash: string;
        id: string;
        name: string;
        status: StoreMagnetStatus;
      }>;
      total_items: number;
    }>("/v0/store/magnets", { method: "GET", params });
  }

  async removeMagnet(magnetId: string) {
    return await this.#client.request<null>(`/v0/store/magnets/${magnetId}`, {
      method: "DELETE",
    });
  }
}

export class StremThru {
  baseUrl: string;

  store: StremThruStore;

  #headers: Record<string, unknown>;
  #timeout?: number;

  constructor(config: StremThruConfig) {
    this.baseUrl = config.baseUrl;

    this.#headers = {
      "User-Agent": [USER_AGENT, config.userAgent].filter(Boolean).join(" "),
    };
    if (config.timeout) {
      this.#timeout = config.timeout;
    }

    if (config.auth) {
      if (typeof config.auth === "object" && "user" in config.auth) {
        config.auth = `${config.auth.user}:${config.auth.pass}`;
      }

      if (typeof config.auth === "string") {
        if (config.auth.includes(":")) {
          config.auth = Buffer.from(config.auth.trim()).toString("base64");
        }
        this.#headers["X-StremThru-Authorization"] = `Basic ${config.auth}`;
      } else if ("store" in config.auth) {
        this.#headers["X-StremThru-Store-Name"] = config.auth.store;
        this.#headers["X-StremThru-Store-Authorization"] =
          `Bearer ${config.auth.token}`;
      }
    }

    this.store = new StremThruStore(this, config.clientIp);
  }

  async health() {
    return await this.request<{ status: "ok" }>(`/v0/health`, {});
  }

  async request<T>(
    endpoint: string,
    {
      body,
      headers,
      method = "GET",
      params,
      ...options
    }: Omit<RequestInit, "body"> & {
      body?: Record<string, unknown> | URLSearchParams;
      params?: Record<string, string | string[]> | URLSearchParams;
    } = {},
  ): Promise<{
    data: T;
    meta: ResponseMeta;
  }> {
    const url = new URL(endpoint, this.baseUrl);
    if (params) {
      url.search = new URLSearchParams(params).toString();
    }

    headers = new Headers({
      accept: "*/*",
      "accept-encoding": "gzip,deflate",
      ...this.#headers,
      ...headers,
    });

    const req: RequestInit = {
      ...options,
      method,
      timeout: this.#timeout,
    };

    if (body instanceof URLSearchParams) {
      headers.set("Content-Type", "application/x-www-form-urlencoded");
      req.body = body;
    } else if (typeof body === "object") {
      headers.set("Content-Type", "application/json");
      req.body = JSON.stringify(body);
    }

    req.headers = headers;

    const res = await fetch(url, req);

    const contentType = res.headers.get("content-type") ?? "";

    const resBody = contentType.includes("application/json")
      ? await res.json()
      : await res.text();

    const meta: ResponseMeta = {
      headers: Object.fromEntries(res.headers.entries()),
      statusCode: res.status,
      statusText: res.statusText,
    };

    if (!res.ok) {
      const opts: ConstructorParameters<typeof StremThruError>[1] = {
        ...meta,
        body: resBody,
      };
      if (typeof resBody === "object") {
        const error = resBody.error;
        opts.type = error.type;
        opts.code = error.code;
      }
      throw new StremThruError(
        typeof resBody === "string"
          ? resBody
          : "message" in resBody.error
            ? `(${resBody.error.type}) ${resBody.error.message}`
            : JSON.stringify(resBody.error),
        opts,
      );
    }

    return {
      data: resBody.data,
      meta,
    };
  }
}
