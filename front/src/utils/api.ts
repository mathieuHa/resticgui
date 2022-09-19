async function call(method: string, url: string, options?: any, data?: any) {
  const response = await fetch(`/api/v1${url}`, {
    method,
    ...data ? { body: JSON.stringify(data) } : {},
    headers: {
      'Content-Type': 'application/json'
    }
  });

  let body: any;
  try {
    body = await response.json();
  } catch (err) {
    body = {};
  }
  if (response.status >= 400) {
    throw new Error(body.error ? `${body.error}: ${body.message}` : '' + response.status);
  }
  return body;
}

const agent = {
  get: async (url, options) => call('GET', url, options),
  delete: async (url, options) => call('DELETE', url, options),
  post: async (url, data, options) => call('POST', url, options, data),
  put: async (url, data, options) => call('PUT', url, options, data),
  patch: async (url, data, options) => call('PATCH', url, options, data),
};

export default agent;
