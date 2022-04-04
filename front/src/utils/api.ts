async function call(method, url, options, data?) {
  let response;
  try {
    response = await fetch(`/api/v1${url}`, {
      method,
      ...data ? { body: JSON.stringify(data) } : {},
      headers: {
        'Content-Type': 'application/json'
      }
    });
  } catch (err) {
    throw err;
  }

  let body;
  try {
    body = await response.json();
  } catch (err) {
    body = {};
  }
  if (response.status >= 400) {
    throw new Error(body.error || response.status);
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
