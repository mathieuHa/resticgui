async function call(method: string, url: string, options: any = {}, data?: any) {
  const response = await fetch(`/api/v1${url}`, {
    ...options,
    method,
    ...data ? { body: JSON.stringify(data) } : {},
    headers: {
      'Content-Type': 'application/json',
    },
  });

  let body: any;
  try {
    body = await response.json();
  } catch (err) {
    body = {};
  }
  if (response.status >= 400) {
    throw new Error(`${body.error ? `${body.error}: ${body.message}` : ''} ${response.status}`);
  }
  return body;
}

const agent = {
  get: async (url: string, options?: any) => call('GET', url, options),
  delete: async (url: string, options?: any) => call('DELETE', url, options),
  post: async (url: string, data: any, options?: any) => call('POST', url, options, data),
  put: async (url: string, data: any, options?: any) => call('PUT', url, options, data),
  patch: async (url: string, data: any, options?: any) => call('PATCH', url, options, data),
};

export default agent;
