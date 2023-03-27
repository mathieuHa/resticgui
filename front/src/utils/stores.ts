import { writable } from 'svelte/store';

interface Toast {
  id: number
  type: string
  value: string
  timeout: unknown
}

const toasts = (() => {
  const { subscribe, update } = writable([] as Toast[]);

  return {
    subscribe,
    clear: () => update(() => ([])),
    addToast: (value: string, type = 'success') => update((array: Toast[]) => {
      array.push({
        id: Math.round(Math.random() * 1000),
        type,
        value,
        timeout: setTimeout(() => update((array2: Toast[]) => {
          array2.shift();
          return array2;
        }), 5000),
      });
      return array;
    }),
  };
})();

const refresh = (() => {
  const { subscribe, update } = writable(0);
  return {
    subscribe,
    update: () => update((data) => data + 1),
  };
})();

export { toasts, refresh };
