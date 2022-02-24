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
  if (response.status >= 400) throw Error(response.status.toString());
  try {
    return await response.json();
  } catch (err) {
    return true;
  }
}

const agent = {
  get: async (url, options) => call('GET', url, options),
  delete: async (url, options) => call('DELETE', url, options),
  post: async (url, data, options) => call('POST', url, options, data),
  put: async (url, data, options) => call('PUT', url, options, data),
  patch: async (url, data, options) => call('PATCH', url, options, data),
};

export default agent;
