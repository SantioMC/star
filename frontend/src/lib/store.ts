import { browser } from "$app/environment";
import { writable, type Writable } from "svelte/store";

export const persistentStore = (
  name: string,
  initialValue: string | null = null,
): Writable<string | null> => {
  const store = writable(initialValue);
  if (!browser) return store;

  const storedValue = localStorage.getItem(name);
  if (storedValue !== null) {
    store.set(storedValue);
  }

  store.subscribe((value) => {
    if (!value) localStorage.removeItem(name);
    else localStorage.setItem(name, value);
  });

  return store;
};

export const typedPersistentStore = <T>(
  name: string,
  initialValue: T,
): Writable<T> => {
  const store = writable(initialValue);
  if (!browser) return store;

  const storedValue = localStorage.getItem(name);
  if (storedValue !== null) {
    store.set(JSON.parse(storedValue) as T);
  }

  store.subscribe((value) => {
    if (value == null) localStorage.removeItem(name);
    else localStorage.setItem(name, JSON.stringify(value));
  });

  return store;
};
