// src/lib/httpClient.ts

type RequestOptions = Omit<RequestInit, 'body'> & {
    body?: Record<string, any> | FormData | null;
};

export interface SuccessResponse<T> {
    data: T;
    message: string;
}

export interface ErrorResponse {
    error: string;
}

const BASE_URL = import.meta.env.VITE_API_URL || '';

async function request<T>(method: string, url: string, options: RequestOptions = {}): Promise<SuccessResponse<T> | ErrorResponse> {
    const fullUrl = BASE_URL + url;

    const headers: HeadersInit = {
        'Accept': 'application/json',
        ...options.headers,
    };

    let body: BodyInit | undefined;
    if (options.body instanceof FormData) {
        body = options.body;
    } else if (options.body) {
        headers['Content-Type'] = 'application/json';
        body = JSON.stringify(options.body);
    }

    try {
        const response = await fetch(fullUrl, {
            method,
            ...options,
            headers,
            body,
        });

        const contentType = response.headers.get('Content-Type');
        const isJson = contentType && contentType.includes('application/json');

        if (response.ok) {
            if (isJson) {
                const data = await response.json();
                return {
                    data: data as T,
                    message: 'Request successful',
                };
            }
            return {
                data: {} as T,
                message: 'Request successful (no content)',
            };
        } else {
            if (isJson) {
                const errorBody = await response.json();
                return {
                    error: errorBody.error || response.statusText,
                };
            } else {
                const errorText = await response.text();
                return {
                    error: errorText || response.statusText,
                };
            }
        }
    } catch (err: any) {
        return {
            error: err.message || 'Unknown error',
        };
    }
}

export const httpClient = {
    get: <T>(url: string, options?: RequestOptions) => request<T>('GET', url, options),
    post: <T>(url: string, body?: Record<string, any> | FormData, options?: RequestOptions) =>
        request<T>('POST', url, { ...options, body }),
    put: <T>(url: string, body?: Record<string, any> | FormData, options?: RequestOptions) =>
        request<T>('PUT', url, { ...options, body }),
    delete: <T>(url: string, options?: RequestOptions) => request<T>('DELETE', url, options),
    patch: <T>(url: string, options?: RequestOptions) => request<T>('PATCH', url, options),
};
