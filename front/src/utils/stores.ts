import { writable } from 'svelte/store';

const toasts = (() => {
  const { subscribe, update } = writable([]);

  return {
    subscribe,
    addToast: (value, type = 'success') => update((array) => {
      array.push({
        id: Math.round(Math.random() * 1000),
        type,
        value,
        timeout: setTimeout(() => update((array) => {
          array.shift();
          return array;
        }), 3000),
      });

      return array;
    }),
  }
})();

export { toasts };
